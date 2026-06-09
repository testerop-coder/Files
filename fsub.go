package handlers

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"telegram-bot/database"
	"telegram-bot/middleware"
	"telegram-bot/models"
	"telegram-bot/utils"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// ─── /addfsub ─────────────────────────────────────────────────────────────────

func (h *Handler) handleAddFSub(msg *tgbotapi.Message) {
	if !middleware.IsAdmin(msg.From.ID, h.cfg) {
		h.sendMsg(msg.Chat.ID, "❌ Sirf admins ye command use kar sakte hain!", msg.MessageID)
		return
	}

	args := cleanArgs(msg.CommandArguments())
	if args == "" {
		h.sendMsg(msg.Chat.ID, `❌ Channel ID dijiye:
<code>/addfsub -100XXXXXXXXXX</code>

<i>Bot ko us channel ka admin banana zaruri hai!</i>`, msg.MessageID)
		return
	}

	channelID, err := strconv.ParseInt(args, 10, 64)
	if err != nil {
		h.sendMsg(msg.Chat.ID, "❌ Valid channel ID dijiye. Example: <code>-1001234567890</code>", msg.MessageID)
		return
	}

	// Verify bot is admin in that channel
	chatConfig := tgbotapi.ChatInfoConfig{ChatConfig: tgbotapi.ChatConfig{ChatID: channelID}}
	chat, err := h.bot.GetChat(chatConfig)
	if err != nil {
		h.sendMsg(msg.Chat.ID, fmt.Sprintf("❌ Channel access nahi ho pa raha.\nError: <code>%v</code>\n\n<i>Bot ko channel ka admin banao pehle!</i>", err), msg.MessageID)
		return
	}

	ch := &models.FSubChannel{
		ID:          database.NewObjectID(),
		ChannelID:   channelID,
		ChannelName: chat.Title,
		Mode:        "normal",
		AddedBy:     msg.From.ID,
		AddedAt:     time.Now(),
	}

	if err := database.AddFSubChannel(ch); err != nil {
		h.sendMsg(msg.Chat.ID, "❌ FSub channel save karne mein error.", msg.MessageID)
		return
	}

	h.sendMsg(msg.Chat.ID, fmt.Sprintf(`✅ <b>FSub Channel Added!</b>

📢 <b>Channel:</b> %s
🆔 <b>ID:</b> <code>%d</code>
🔧 <b>Mode:</b> Normal

Mode change karne ke liye: <code>/fsubmode %d normal</code> ya <code>/fsubmode %d request</code>`,
		chat.Title, channelID, channelID, channelID), msg.MessageID)
}

// ─── /removefsub ──────────────────────────────────────────────────────────────

func (h *Handler) handleRemoveFSub(msg *tgbotapi.Message) {
	if !middleware.IsAdmin(msg.From.ID, h.cfg) {
		h.sendMsg(msg.Chat.ID, "❌ Sirf admins ye command use kar sakte hain!", msg.MessageID)
		return
	}

	args := cleanArgs(msg.CommandArguments())
	if args == "" {
		h.sendMsg(msg.Chat.ID, "❌ Channel ID dijiye: <code>/removefsub -100XXXXXXXXXX</code>", msg.MessageID)
		return
	}

	channelID, err := strconv.ParseInt(args, 10, 64)
	if err != nil {
		h.sendMsg(msg.Chat.ID, "❌ Valid channel ID dijiye.", msg.MessageID)
		return
	}

	if err := database.RemoveFSubChannel(channelID); err != nil {
		h.sendMsg(msg.Chat.ID, "❌ Channel remove karne mein error.", msg.MessageID)
		return
	}

	h.sendMsg(msg.Chat.ID, fmt.Sprintf("✅ FSub Channel <code>%d</code> hata diya gaya!", channelID), msg.MessageID)
}

// ─── /fsubs ───────────────────────────────────────────────────────────────────

func (h *Handler) handleListFSubs(msg *tgbotapi.Message) {
	if !middleware.IsAdmin(msg.From.ID, h.cfg) {
		h.sendMsg(msg.Chat.ID, "❌ Access denied!", msg.MessageID)
		return
	}

	channels, err := database.GetAllFSubChannels()
	if err != nil {
		h.sendMsg(msg.Chat.ID, "❌ Channels fetch karne mein error.", msg.MessageID)
		return
	}

	if len(channels) == 0 {
		h.sendMsg(msg.Chat.ID, "📭 Koi FSub channel set nahi hai.\n\n<code>/addfsub CHANNEL_ID</code> se add karein.", msg.MessageID)
		return
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("📢 <b>FSub Channels (%d)</b>\n\n", len(channels)))

	for i, ch := range channels {
		modeEmoji := "🟢"
		if ch.Mode == "request" {
			modeEmoji = "🔵"
		}
		sb.WriteString(fmt.Sprintf("%d. %s <b>%s</b>\n   🆔 <code>%d</code>\n   %s Mode: <b>%s</b>\n\n",
			i+1, modeEmoji, ch.ChannelName, ch.ChannelID, modeEmoji, strings.Title(ch.Mode)))
	}

	sb.WriteString(`<b>Commands:</b>
/addfsub - Channel add karein
/removefsub - Channel hatain
/fsubmode - Mode change karein
/setexpire - Link expire time set karein`)

	h.sendMsg(msg.Chat.ID, sb.String(), msg.MessageID)
}

// ─── /fsubmode ────────────────────────────────────────────────────────────────

func (h *Handler) handleFSubMode(msg *tgbotapi.Message) {
	if !middleware.IsAdmin(msg.From.ID, h.cfg) {
		h.sendMsg(msg.Chat.ID, "❌ Sirf admins ye command use kar sakte hain!", msg.MessageID)
		return
	}

	args := strings.Fields(msg.CommandArguments())
	if len(args) < 2 {
		h.sendMsg(msg.Chat.ID, `❌ Usage: <code>/fsubmode CHANNEL_ID MODE</code>

Modes:
• <code>normal</code> - Direct join link
• <code>request</code> - Join request link

Example: <code>/fsubmode -1001234567890 request</code>`, msg.MessageID)
		return
	}

	channelID, err := strconv.ParseInt(args[0], 10, 64)
	if err != nil {
		h.sendMsg(msg.Chat.ID, "❌ Valid channel ID dijiye.", msg.MessageID)
		return
	}

	mode := strings.ToLower(args[1])
	if mode != "normal" && mode != "request" {
		h.sendMsg(msg.Chat.ID, "❌ Mode sirf <b>normal</b> ya <b>request</b> ho sakta hai.", msg.MessageID)
		return
	}

	if err := database.UpdateFSubMode(channelID, mode); err != nil {
		h.sendMsg(msg.Chat.ID, "❌ Mode update karne mein error.", msg.MessageID)
		return
	}

	modeEmoji := "🟢"
	if mode == "request" {
		modeEmoji = "🔵"
	}
	h.sendMsg(msg.Chat.ID, fmt.Sprintf("✅ Channel <code>%d</code> ka mode <b>%s %s</b> set kar diya!", channelID, modeEmoji, strings.Title(mode)), msg.MessageID)
}

// ─── /setexpire ───────────────────────────────────────────────────────────────

func (h *Handler) handleSetExpire(msg *tgbotapi.Message) {
	if !middleware.IsAdmin(msg.From.ID, h.cfg) {
		h.sendMsg(msg.Chat.ID, "❌ Sirf admins ye command use kar sakte hain!", msg.MessageID)
		return
	}

	args := cleanArgs(msg.CommandArguments())
	if args == "" {
		settings := database.GetBotSettings()
		keyboard := tgbotapi.NewInlineKeyboardMarkup(
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData("5 Min", "expire_5"),
				tgbotapi.NewInlineKeyboardButtonData("10 Min", "expire_10"),
				tgbotapi.NewInlineKeyboardButtonData("15 Min", "expire_15"),
			),
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData("30 Min", "expire_30"),
				tgbotapi.NewInlineKeyboardButtonData("1 Hour", "expire_60"),
				tgbotapi.NewInlineKeyboardButtonData("2 Hours", "expire_120"),
			),
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData("6 Hours", "expire_360"),
				tgbotapi.NewInlineKeyboardButtonData("12 Hours", "expire_720"),
				tgbotapi.NewInlineKeyboardButtonData("24 Hours", "expire_1440"),
			),
		)
		h.sendMsgWithKeyboard(msg.Chat.ID, fmt.Sprintf(`⏰ <b>FSub Link Expire Time</b>

Current: <b>%d minutes</b>

Neeche se select karein ya manually:
<code>/setexpire MINUTES</code>`, settings.FSubExpireMinutes), keyboard)
		return
	}

	minutes, err := strconv.Atoi(args)
	if err != nil || minutes < 1 {
		h.sendMsg(msg.Chat.ID, "❌ Valid minutes enter karein (minimum 1)", msg.MessageID)
		return
	}

	if err := database.SetSetting("fsub_expire_minutes", strconv.Itoa(minutes)); err != nil {
		h.sendMsg(msg.Chat.ID, "❌ Setting save karne mein error.", msg.MessageID)
		return
	}

	h.sendMsg(msg.Chat.ID, fmt.Sprintf("✅ FSub link expire time <b>%d minutes</b> set kar diya!", minutes), msg.MessageID)
}

// ─── FSUB CHECK ───────────────────────────────────────────────────────────────

// checkFSub returns (blocked bool, message string)
func (h *Handler) checkFSub(userID, chatID int64) (bool, string) {
	channels, err := database.GetAllFSubChannels()
	if err != nil || len(channels) == 0 {
		return false, ""
	}

	settings := database.GetBotSettings()
	var notJoined []models.FSubChannel

	for _, ch := range channels {
		member, err := h.bot.GetChatMember(tgbotapi.GetChatMemberConfig{
			ChatConfigWithUser: tgbotapi.ChatConfigWithUser{
				ChatID: ch.ChannelID,
				UserID: userID,
			},
		})
		if err != nil {
			notJoined = append(notJoined, ch)
			continue
		}
		status := member.Status
		if status == "left" || status == "kicked" {
			notJoined = append(notJoined, ch)
		}
	}

	if len(notJoined) == 0 {
		return false, ""
	}

	// Build message with join buttons
	var sb strings.Builder
	sb.WriteString("⚠️ <b>Pehle in channels ko join karein:</b>\n\n")

	var rows [][]tgbotapi.InlineKeyboardButton

	for i, ch := range notJoined {
		sb.WriteString(fmt.Sprintf("%d. 📢 <b>%s</b>\n", i+1, ch.ChannelName))

		// Generate invite link with expiry
		expireTime := time.Now().Add(time.Duration(settings.FSubExpireMinutes) * time.Minute)

		var inviteLink string
		if ch.Mode == "request" {
			// Request mode - creates approval-based link
			linkConfig := tgbotapi.CreateChatInviteLinkConfig{
				ChatConfig:         tgbotapi.ChatConfig{ChatID: ch.ChannelID},
				ExpireDate:         int(expireTime.Unix()),
				CreatesJoinRequest: true,
			}
			link, err := h.bot.CreateChatInviteLink(linkConfig)
			if err == nil {
				inviteLink = link.InviteLink
			}
		} else {
			// Normal mode
			linkConfig := tgbotapi.CreateChatInviteLinkConfig{
				ChatConfig: tgbotapi.ChatConfig{ChatID: ch.ChannelID},
				ExpireDate: int(expireTime.Unix()),
			}
			link, err := h.bot.CreateChatInviteLink(linkConfig)
			if err == nil {
				inviteLink = link.InviteLink
			}
		}

		if inviteLink != "" {
			btnText := fmt.Sprintf("📢 Join %s", ch.ChannelName)
			if ch.Mode == "request" {
				btnText = fmt.Sprintf("🔵 Request %s", ch.ChannelName)
			}
			rows = append(rows, tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonURL(btnText, inviteLink),
			))
		}
	}

	sb.WriteString(fmt.Sprintf("\n⏰ Links <b>%d minutes</b> mein expire ho jayenge!", settings.FSubExpireMinutes))
	sb.WriteString("\n\nJoin karne ke baad <b>Try Again</b> press karein! 👇")

	rows = append(rows, tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData("🔄 Try Again", fmt.Sprintf("fsub_check_%d", userID)),
	))

	keyboard := tgbotapi.NewInlineKeyboardMarkup(rows...)
	_ = utils.EscapeMarkdown
	_ = keyboard

	// Send with keyboard
	sendMsg := tgbotapi.NewMessage(chatID, sb.String())
	sendMsg.ParseMode = "HTML"
	sendMsg.ReplyMarkup = keyboard
	h.bot.Send(sendMsg)

	return true, ""
}

// ─── JOIN REQUEST HANDLER ─────────────────────────────────────────────────────

func (h *Handler) handleJoinRequest(req *tgbotapi.ChatJoinRequest) {
	// Auto-approve join requests for request-mode FSub channels
	channels, err := database.GetAllFSubChannels()
	if err != nil {
		return
	}

	for _, ch := range channels {
		if ch.ChannelID == req.Chat.ID && ch.Mode == "request" {
			approveConfig := tgbotapi.ApproveChatJoinRequestConfig{
				ChatConfig: tgbotapi.ChatConfig{ChatID: req.Chat.ID},
				UserID:     req.From.ID,
			}
			_, _ = h.bot.Request(approveConfig)

			// Notify user
			h.sendMsg(req.From.ID, fmt.Sprintf("✅ Aapki join request <b>%s</b> mein approve ho gayi!", req.Chat.Title), 0)
			break
		}
	}
}
