package handlers

import (
	"fmt"
	"log"
	"strings"
	"time"

	"telegram-bot/database"
	"telegram-bot/middleware"
	"telegram-bot/models"
	"telegram-bot/utils"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// ─── /start ──────────────────────────────────────────────────────────────────

func (h *Handler) handleStart(msg *tgbotapi.Message, args string) {
	if args == "" {
		h.sendWelcome(msg)
		return
	}

	args = cleanArgs(args)

	if strings.HasPrefix(args, "file_") {
		token := strings.TrimPrefix(args, "file_")
		h.deliverFile(msg, token)
	} else if strings.HasPrefix(args, "batch_") {
		token := strings.TrimPrefix(args, "batch_")
		h.deliverBatch(msg, token)
	} else {
		h.sendWelcome(msg)
	}
}

func (h *Handler) sendWelcome(msg *tgbotapi.Message) {
	text := fmt.Sprintf(`👋 <b>Hello, %s!</b>

🤖 <b>File Provider Bot</b> ke sawagat hai!

📂 Ye bot aapko private channel se files share karne ki facility deta hai.

<b>Commands:</b>
/getlink - Single file ka link banao
/batch - Multiple files ka link banao
/help - Poori help list

<i>Admin se link maango aur /start se file pao!</i>`,
		msg.From.FirstName)

	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonURL("📢 Updates Channel", "https://t.me/"),
		),
	)
	h.sendMsgWithKeyboard(msg.Chat.ID, text, keyboard)
}

// ─── /getlink ─────────────────────────────────────────────────────────────────

func (h *Handler) handleGetLink(msg *tgbotapi.Message) {
	if !middleware.IsAdmin(msg.From.ID, h.cfg) {
		h.sendMsg(msg.Chat.ID, "❌ Sirf admins ye command use kar sakte hain!", msg.MessageID)
		return
	}
	if !h.isPrivate(msg) {
		h.sendMsg(msg.Chat.ID, "📩 Ye command sirf private chat mein use karein.", msg.MessageID)
		return
	}
	if msg.ReplyToMessage == nil {
		h.sendMsg(msg.Chat.ID, "📎 Kisi file message ko reply karke /getlink bhejein.\n\n<i>DB channel se file forward karein phir reply karein.</i>", msg.MessageID)
		return
	}

	replied := msg.ReplyToMessage

	// Extract file info
	fileID, fileName, fileSize, fileType := extractFileInfo(replied)
	if fileID == "" {
		h.sendMsg(msg.Chat.ID, "❌ Koi file nahi mili. Kisi document/video/audio/photo ko reply karein.", msg.MessageID)
		return
	}

	token := utils.GenerateToken(16)

	file := &models.File{
		ID:          database.NewObjectID(),
		FileID:      fileID,
		MessageID:   replied.MessageID,
		FileName:    fileName,
		FileSize:    fileSize,
		FileType:    fileType,
		UniqueToken: token,
		AddedBy:     msg.From.ID,
		AddedAt:     time.Now(),
	}

	if err := database.SaveFile(file); err != nil {
		log.Printf("SaveFile error: %v", err)
		h.sendMsg(msg.Chat.ID, "❌ File save karne mein error aaya.", msg.MessageID)
		return
	}

	link := h.buildBotLink(token, false)

	text := fmt.Sprintf(`✅ <b>File Link Ready!</b>

%s <b>File:</b> %s
📦 <b>Size:</b> %s
🔗 <b>Link:</b>
<code>%s</code>

<i>Link share karein, users /start se file paa sakte hain.</i>`,
		utils.FileTypeEmoji(fileType), fileName,
		database.FormatSize(fileSize), link)

	h.sendMsg(msg.Chat.ID, text, msg.MessageID)
}

// ─── /batch ──────────────────────────────────────────────────────────────────

func (h *Handler) handleBatchStart(msg *tgbotapi.Message) {
	if !middleware.IsAdmin(msg.From.ID, h.cfg) {
		h.sendMsg(msg.Chat.ID, "❌ Sirf admins ye command use kar sakte hain!", msg.MessageID)
		return
	}
	if !h.isPrivate(msg) {
		h.sendMsg(msg.Chat.ID, "📩 Ye command sirf private chat mein use karein.", msg.MessageID)
		return
	}

	h.batchState[msg.From.ID] = -1 // Mark as waiting for first URL

	text := `📦 <b>Batch Link Mode</b>

<b>Step 1:</b> DB Channel se <b>pehli file</b> ka message forward karein.

<i>⚠️ Sirf DB channel se forward karein!</i>
/cancel - Batch cancel karein`

	h.sendMsg(msg.Chat.ID, text, msg.MessageID)
}

func (h *Handler) handleBatchSecondURL(msg *tgbotapi.Message) {
	userID := msg.From.ID
	firstMsgID, waiting := h.batchState[userID]

	if !waiting {
		return
	}

	// Check it's from DB channel
	if !h.isForwardFromDBChannel(msg) {
		h.sendMsg(msg.Chat.ID, "❌ Sirf DB Channel se messages forward karein!", msg.MessageID)
		return
	}

	msgID := h.getForwardedMsgID(msg)
	if msgID == 0 {
		h.sendMsg(msg.Chat.ID, "❌ Valid forwarded message nahi mila.", msg.MessageID)
		return
	}

	if firstMsgID == -1 {
		// Save first message ID
		h.batchState[userID] = msgID
		text := fmt.Sprintf(`✅ <b>Pehli file set!</b>
📌 Message ID: <code>%d</code>

<b>Step 2:</b> DB Channel se <b>aakhri file</b> ka message forward karein.`, msgID)
		h.sendMsg(msg.Chat.ID, text, msg.MessageID)
		return
	}

	// We have both - create batch link
	startMsgID := firstMsgID
	endMsgID := msgID

	if startMsgID > endMsgID {
		startMsgID, endMsgID = endMsgID, startMsgID
	}

	token := utils.GenerateToken(16)
	batch := &models.BatchLink{
		ID:          database.NewObjectID(),
		UniqueToken: token,
		StartMsgID:  startMsgID,
		EndMsgID:    endMsgID,
		CreatedBy:   userID,
		CreatedAt:   time.Now(),
	}

	if err := database.SaveBatch(batch); err != nil {
		log.Printf("SaveBatch error: %v", err)
		h.sendMsg(msg.Chat.ID, "❌ Batch save karne mein error.", msg.MessageID)
		delete(h.batchState, userID)
		return
	}

	delete(h.batchState, userID)
	link := h.buildBotLink(token, true)
	count := endMsgID - startMsgID + 1

	text := fmt.Sprintf(`✅ <b>Batch Link Ready!</b>

📁 <b>Files:</b> %d files (MSG %d → %d)
🔗 <b>Link:</b>
<code>%s</code>

<i>Users ko ye link share karein!</i>`, count, startMsgID, endMsgID, link)

	h.sendMsg(msg.Chat.ID, text, msg.MessageID)
}

// ─── FILE DELIVERY ────────────────────────────────────────────────────────────

func (h *Handler) deliverFile(msg *tgbotapi.Message, token string) {
	file, err := database.GetFileByToken(token)
	if err != nil {
		h.sendMsg(msg.Chat.ID, "❌ File nahi mili ya link expired ho gaya!", msg.MessageID)
		return
	}

	settings := database.GetBotSettings()

	// Forward from DB channel
	fwd := tgbotapi.NewForward(msg.Chat.ID, h.cfg.DBChannelID, file.MessageID)
	sent, err := h.bot.Send(fwd)
	if err != nil {
		log.Printf("Forward file error: %v", err)
		// Try sending by file ID
		h.sendFileByID(msg.Chat.ID, file, msg.MessageID)
		return
	}

	// Schedule auto-delete
	if settings.AutoDeleteSeconds > 0 {
		info := h.sendMsg(msg.Chat.ID, fmt.Sprintf("⏱ Ye file <b>%d seconds</b> mein delete ho jayegi!", settings.AutoDeleteSeconds), 0)
		h.scheduleDelete(msg.Chat.ID, sent.MessageID, settings.AutoDeleteSeconds)
		if info != nil {
			h.scheduleDelete(msg.Chat.ID, info.MessageID, settings.AutoDeleteSeconds)
		}
	}
}

func (h *Handler) deliverBatch(msg *tgbotapi.Message, token string) {
	batch, err := database.GetBatchByToken(token)
	if err != nil {
		h.sendMsg(msg.Chat.ID, "❌ Batch link nahi mili ya expired ho gayi!", msg.MessageID)
		return
	}

	settings := database.GetBotSettings()
	count := batch.EndMsgID - batch.StartMsgID + 1

	info := h.sendMsg(msg.Chat.ID, fmt.Sprintf("📦 <b>%d files</b> send ho rahi hain, thoda wait karein...", count), msg.MessageID)

	var sentIDs []int

	for msgID := batch.StartMsgID; msgID <= batch.EndMsgID; msgID++ {
		fwd := tgbotapi.NewForward(msg.Chat.ID, h.cfg.DBChannelID, msgID)
		sent, err := h.bot.Send(fwd)
		if err != nil {
			log.Printf("Batch forward msgID %d error: %v", msgID, err)
			continue
		}
		sentIDs = append(sentIDs, sent.MessageID)
		time.Sleep(300 * time.Millisecond) // Anti-flood
	}

	if info != nil {
		h.deleteMsg(msg.Chat.ID, info.MessageID)
	}

	doneMsg := h.sendMsg(msg.Chat.ID, fmt.Sprintf("✅ <b>%d/%d files</b> successfully send ho gayi!", len(sentIDs), count), 0)

	// Schedule auto-delete
	if settings.AutoDeleteSeconds > 0 {
		for _, id := range sentIDs {
			h.scheduleDelete(msg.Chat.ID, id, settings.AutoDeleteSeconds)
		}
		if doneMsg != nil {
			h.scheduleDelete(msg.Chat.ID, doneMsg.MessageID, settings.AutoDeleteSeconds)
		}
		h.sendMsg(msg.Chat.ID, fmt.Sprintf("⏱ Files <b>%d seconds</b> mein delete ho jayengi!", settings.AutoDeleteSeconds), 0)
	}
}

func (h *Handler) sendFileByID(chatID int64, file *models.File, replyTo int) {
	var sendable tgbotapi.Chattable
	switch file.FileType {
	case "document":
		d := tgbotapi.NewDocument(chatID, tgbotapi.FileID(file.FileID))
		d.ReplyToMessageID = replyTo
		sendable = d
	case "video":
		v := tgbotapi.NewVideo(chatID, tgbotapi.FileID(file.FileID))
		v.ReplyToMessageID = replyTo
		sendable = v
	case "audio":
		a := tgbotapi.NewAudio(chatID, tgbotapi.FileID(file.FileID))
		a.ReplyToMessageID = replyTo
		sendable = a
	case "photo":
		p := tgbotapi.NewPhoto(chatID, tgbotapi.FileID(file.FileID))
		p.ReplyToMessageID = replyTo
		sendable = p
	default:
		h.sendMsg(chatID, "❌ File type unsupported.", replyTo)
		return
	}
	_, err := h.bot.Send(sendable)
	if err != nil {
		log.Printf("sendFileByID error: %v", err)
	}
}

// ─── DB CHANNEL POST HANDLER ─────────────────────────────────────────────────

func (h *Handler) handleDBChannelPost(post *tgbotapi.Message) {
	// Auto-index files posted to DB channel
	fileID, fileName, fileSize, fileType := extractFileInfo(post)
	if fileID == "" {
		return
	}

	token := utils.GenerateToken(16)
	file := &models.File{
		ID:          database.NewObjectID(),
		FileID:      fileID,
		MessageID:   post.MessageID,
		FileName:    fileName,
		FileSize:    fileSize,
		FileType:    fileType,
		UniqueToken: token,
		AddedBy:     0, // system
		AddedAt:     time.Now(),
	}
	if err := database.SaveFile(file); err != nil {
		// May already exist - ok
		log.Printf("DB Channel file index: %v", err)
	}
}

// ─── HELPERS ─────────────────────────────────────────────────────────────────

func extractFileInfo(msg *tgbotapi.Message) (fileID, fileName string, fileSize int64, fileType string) {
	if msg.Document != nil {
		return msg.Document.FileID, msg.Document.FileName, int64(msg.Document.FileSize), "document"
	}
	if msg.Video != nil {
		name := msg.Video.FileName
		if name == "" {
			name = fmt.Sprintf("video_%d.mp4", msg.MessageID)
		}
		return msg.Video.FileID, name, int64(msg.Video.FileSize), "video"
	}
	if msg.Audio != nil {
		name := msg.Audio.Title
		if name == "" {
			name = fmt.Sprintf("audio_%d.mp3", msg.MessageID)
		}
		return msg.Audio.FileID, name, int64(msg.Audio.FileSize), "audio"
	}
	if len(msg.Photo) > 0 {
		p := msg.Photo[len(msg.Photo)-1] // Largest
		return p.FileID, fmt.Sprintf("photo_%d.jpg", msg.MessageID), int64(p.FileSize), "photo"
	}
	return "", "", 0, ""
}

func (h *Handler) isForwardFromDBChannel(msg *tgbotapi.Message) bool {
	if msg.ForwardFromChat != nil && msg.ForwardFromChat.ID == h.cfg.DBChannelID {
		return true
	}
	// ForwardOrigin check (newer API)
	if msg.ForwardOrigin != nil {
		if msg.ForwardOrigin.Type == "channel" {
			return true // Assume correct channel for now
		}
	}
	return false
}

func (h *Handler) getForwardedMsgID(msg *tgbotapi.Message) int {
	if msg.ForwardFromMessageID != 0 {
		return msg.ForwardFromMessageID
	}
	return 0
}

func init() {
	_ = strings.TrimSpace
}
