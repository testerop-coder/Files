package handlers

import (
	"fmt"
	"strconv"
	"strings"

	"telegram-bot/database"
	"telegram-bot/middleware"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func (h *Handler) handleCallback(cb *tgbotapi.CallbackQuery) {
	data := cb.Data
	chatID := cb.Message.Chat.ID
	msgID := cb.Message.MessageID
	userID := cb.From.ID

	// Always answer callback to remove loading state
	defer func() {
		answer := tgbotapi.NewCallback(cb.ID, "")
		h.bot.Request(answer)
	}()

	switch {
	// FSub expire time quick buttons
	case strings.HasPrefix(data, "expire_"):
		h.handleExpireCallback(cb, data, chatID, msgID, userID)

	// FSub re-check button
	case strings.HasPrefix(data, "fsub_check_"):
		h.handleFSubCheckCallback(cb, chatID, msgID, userID)

	default:
		answer := tgbotapi.NewCallback(cb.ID, "❓ Unknown action")
		h.bot.Request(answer)
	}
}

// ─── EXPIRE CALLBACK ──────────────────────────────────────────────────────────

func (h *Handler) handleExpireCallback(cb *tgbotapi.CallbackQuery, data string, chatID int64, msgID int, userID int64) {
	if !middleware.IsAdmin(userID, h.cfg) {
		answer := tgbotapi.NewCallback(cb.ID, "❌ Permission denied!")
		h.bot.Request(answer)
		return
	}

	minutesStr := strings.TrimPrefix(data, "expire_")
	minutes, err := strconv.Atoi(minutesStr)
	if err != nil {
		answer := tgbotapi.NewCallback(cb.ID, "❌ Invalid value")
		h.bot.Request(answer)
		return
	}

	if err := database.SetSetting("fsub_expire_minutes", minutesStr); err != nil {
		answer := tgbotapi.NewCallback(cb.ID, "❌ Save failed!")
		h.bot.Request(answer)
		return
	}

	// Format display time
	displayTime := fmt.Sprintf("%d minutes", minutes)
	if minutes >= 60 {
		hours := minutes / 60
		mins := minutes % 60
		if mins == 0 {
			displayTime = fmt.Sprintf("%d hour(s)", hours)
		} else {
			displayTime = fmt.Sprintf("%d hour(s) %d min", hours, mins)
		}
	}

	// Edit the message
	newText := fmt.Sprintf(`✅ <b>FSub Expire Time Updated!</b>

⏰ New Time: <b>%s</b>

<i>Naye links ab %s mein expire honge.</i>`, displayTime, displayTime)

	editMsg := tgbotapi.NewEditMessageText(chatID, msgID, newText)
	editMsg.ParseMode = "HTML"
	h.bot.Send(editMsg)

	answer := tgbotapi.NewCallback(cb.ID, fmt.Sprintf("✅ Set to %s", displayTime))
	h.bot.Request(answer)
}

// ─── FSUB CHECK CALLBACK ──────────────────────────────────────────────────────

func (h *Handler) handleFSubCheckCallback(cb *tgbotapi.CallbackQuery, chatID int64, msgID int, userID int64) {
	// Re-check subscription
	blocked, _ := h.checkFSubSilent(userID)

	if blocked {
		answer := tgbotapi.NewCallback(cb.ID, "❌ Abhi bhi kuch channels join nahi kiye!")
		h.bot.Request(answer)
		return
	}

	// User has joined - delete the fsub message
	h.deleteMsg(chatID, msgID)

	answer := tgbotapi.NewCallback(cb.ID, "✅ Verified! Ab aap bot use kar sakte hain.")
	h.bot.Request(answer)

	h.sendMsg(chatID, "✅ <b>Verification successful!</b>\n\nAb apna kaam dobara karein. 😊", 0)
}

// checkFSubSilent returns blocked=true if user hasn't joined all channels
func (h *Handler) checkFSubSilent(userID int64) (bool, []string) {
	channels, err := database.GetAllFSubChannels()
	if err != nil || len(channels) == 0 {
		return false, nil
	}

	var notJoined []string

	for _, ch := range channels {
		member, err := h.bot.GetChatMember(tgbotapi.GetChatMemberConfig{
			ChatConfigWithUser: tgbotapi.ChatConfigWithUser{
				ChatID: ch.ChannelID,
				UserID: userID,
			},
		})
		if err != nil {
			notJoined = append(notJoined, ch.ChannelName)
			continue
		}
		status := member.Status
		if status == "left" || status == "kicked" {
			notJoined = append(notJoined, ch.ChannelName)
		}
	}

	return len(notJoined) > 0, notJoined
}
