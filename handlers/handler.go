package handlers

import (
	"log"
	"strings"
	"time"

	"telegram-bot/config"
	"telegram-bot/database"
	"telegram-bot/middleware"
	"telegram-bot/models"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"go.mongodb.org/mongo-driver/mongo"
)

type Handler struct {
	bot       *tgbotapi.BotAPI
	db        *mongo.Database
	cfg       *config.Config
	startTime time.Time

	// Batch state: stores first URL per user
	batchState map[int64]int // userID -> first message ID
}

func New(bot *tgbotapi.BotAPI, db *mongo.Database, cfg *config.Config, startTime time.Time) *Handler {
	return &Handler{
		bot:        bot,
		db:         db,
		cfg:        cfg,
		startTime:  startTime,
		batchState: make(map[int64]int),
	}
}

func (h *Handler) HandleUpdate(update tgbotapi.Update) {
	// Handle DB channel posts
	if update.ChannelPost != nil {
		if update.ChannelPost.Chat.ID == h.cfg.DBChannelID {
			h.handleDBChannelPost(update.ChannelPost)
		}
		return
	}

	if update.Message != nil {
		// Register user
		user := &models.User{
			UserID:    update.Message.From.ID,
			Username:  update.Message.From.UserName,
			FirstName: update.Message.From.FirstName,
			JoinedAt:  time.Now(),
		}
		_ = database.UpsertUser(user)

		// Check FSub
		if !middleware.IsAdmin(update.Message.From.ID, h.cfg) {
			if blocked, msg := h.checkFSub(update.Message.From.ID, update.Message.Chat.ID); blocked {
				h.sendMsg(update.Message.Chat.ID, msg, update.Message.MessageID)
				return
			}
		}

		h.handleMessage(update.Message)
	} else if update.CallbackQuery != nil {
		h.handleCallback(update.CallbackQuery)
	} else if update.ChatJoinRequest != nil {
		h.handleJoinRequest(update.ChatJoinRequest)
	}
}

func (h *Handler) handleMessage(msg *tgbotapi.Message) {
	if msg.IsCommand() {
		h.handleCommand(msg)
		return
	}

	// Check if user is in batch mode
	if _, ok := h.batchState[msg.From.ID]; ok && msg.ForwardOrigin != nil {
		h.handleBatchSecondURL(msg)
		return
	}
}

func (h *Handler) handleCommand(msg *tgbotapi.Message) {
	cmd := msg.Command()
	args := msg.CommandArguments()

	switch cmd {
	case "start":
		h.handleStart(msg, args)
	case "getlink":
		h.handleGetLink(msg)
	case "batch":
		h.handleBatchStart(msg)
	case "status":
		h.handleStatus(msg)
	case "addadmin":
		h.handleAddAdmin(msg)
	case "removeadmin":
		h.handleRemoveAdmin(msg)
	case "admins":
		h.handleListAdmins(msg)
	case "addfsub":
		h.handleAddFSub(msg)
	case "removefsub":
		h.handleRemoveFSub(msg)
	case "fsubs":
		h.handleListFSubs(msg)
	case "fsubmode":
		h.handleFSubMode(msg)
	case "setexpire":
		h.handleSetExpire(msg)
	case "setdelete":
		h.handleSetAutoDelete(msg)
	case "broadcast":
		h.handleBroadcast(msg, false)
	case "pbroadcast":
		h.handleBroadcast(msg, true)
	case "ping":
		h.handlePing(msg)
	case "help":
		h.handleHelp(msg)
	default:
		log.Printf("Unknown command: %s", cmd)
	}

	_ = args
}

// ─── SEND HELPERS ────────────────────────────────────────────────────────────

func (h *Handler) sendMsg(chatID int64, text string, replyTo int) *tgbotapi.Message {
	msg := tgbotapi.NewMessage(chatID, text)
	msg.ParseMode = "HTML"
	if replyTo > 0 {
		msg.ReplyToMessageID = replyTo
	}
	sent, err := h.bot.Send(msg)
	if err != nil {
		log.Printf("sendMsg error: %v", err)
		return nil
	}
	return &sent
}

func (h *Handler) sendMsgWithKeyboard(chatID int64, text string, keyboard tgbotapi.InlineKeyboardMarkup) *tgbotapi.Message {
	msg := tgbotapi.NewMessage(chatID, text)
	msg.ParseMode = "HTML"
	msg.ReplyMarkup = keyboard
	sent, err := h.bot.Send(msg)
	if err != nil {
		log.Printf("sendMsgWithKeyboard error: %v", err)
		return nil
	}
	return &sent
}

func (h *Handler) deleteMsg(chatID int64, msgID int) {
	del := tgbotapi.NewDeleteMessage(chatID, msgID)
	_, _ = h.bot.Request(del)
}

func (h *Handler) scheduleDelete(chatID int64, msgID int, seconds int) {
	if seconds <= 0 {
		return
	}
	go func() {
		time.Sleep(time.Duration(seconds) * time.Second)
		h.deleteMsg(chatID, msgID)
	}()
}

func (h *Handler) isPrivate(msg *tgbotapi.Message) bool {
	return msg.Chat.Type == "private"
}

func (h *Handler) extractUserIDFromMsg(msg *tgbotapi.Message) int64 {
	if msg.ReplyToMessage != nil {
		return msg.ReplyToMessage.From.ID
	}
	if len(msg.Entities) > 0 {
		for _, e := range msg.Entities {
			if e.Type == "mention" {
				// @username mention — can't get ID directly without API call
			}
			if e.Type == "text_mention" && e.User != nil {
				return e.User.ID
			}
		}
	}
	return 0
}

func (h *Handler) buildBotLink(token string, isBatch bool) string {
	prefix := "file"
	if isBatch {
		prefix = "batch"
	}
	return "https://t.me/" + h.bot.Self.UserName + "?start=" + prefix + "_" + token
}

func cleanArgs(args string) string {
	return strings.TrimSpace(args)
}
