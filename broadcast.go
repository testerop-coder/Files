package handlers

import (
	"fmt"
	"log"
	"strings"
	"time"

	"telegram-bot/database"
	"telegram-bot/middleware"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// ─── /broadcast ───────────────────────────────────────────────────────────────

func (h *Handler) handleBroadcast(msg *tgbotapi.Message, pin bool) {
	if !middleware.IsAdmin(msg.From.ID, h.cfg) {
		h.sendMsg(msg.Chat.ID, "❌ Sirf admins broadcast kar sakte hain!", msg.MessageID)
		return
	}

	var broadcastText string
	var broadcastMsg *tgbotapi.Message

	// Check if replying to a message
	if msg.ReplyToMessage != nil {
		broadcastMsg = msg.ReplyToMessage
		broadcastText = msg.ReplyToMessage.Text
		if broadcastText == "" {
			broadcastText = msg.ReplyToMessage.Caption
		}
	} else {
		broadcastText = strings.TrimSpace(msg.CommandArguments())
	}

	if broadcastText == "" && broadcastMsg == nil {
		cmdName := "broadcast"
		if pin {
			cmdName = "pbroadcast"
		}
		h.sendMsg(msg.Chat.ID, fmt.Sprintf(`❌ Message dijiye:

Option 1: <code>/%s Aapka message yahan</code>
Option 2: Kisi message ko reply karke <code>/%s</code> bhejein`, cmdName, cmdName), msg.MessageID)
		return
	}

	// Get all user IDs
	userIDs, err := database.GetAllUserIDs()
	if err != nil {
		h.sendMsg(msg.Chat.ID, "❌ Users fetch karne mein error.", msg.MessageID)
		return
	}

	total := len(userIDs)
	statusMsg := h.sendMsg(msg.Chat.ID, fmt.Sprintf("📤 Broadcasting to <b>%d users</b>...\n⏳ Please wait...", total), msg.MessageID)

	success, failed, blocked := 0, 0, 0

	for _, userID := range userIDs {
		var sendErr error

		if broadcastMsg != nil {
			// Forward the replied message
			fwd := tgbotapi.NewForward(userID, broadcastMsg.Chat.ID, broadcastMsg.MessageID)
			sent, err := h.bot.Send(fwd)
			sendErr = err

			// Pin if requested
			if err == nil && pin {
				pinConfig := tgbotapi.PinChatMessageConfig{
					ChatID:              userID,
					MessageID:           sent.MessageID,
					DisableNotification: false,
				}
				_, _ = h.bot.Request(pinConfig)
			}
		} else {
			// Send text message
			newMsg := tgbotapi.NewMessage(userID, broadcastText)
			newMsg.ParseMode = "HTML"
			sent, err := h.bot.Send(newMsg)
			sendErr = err

			// Pin if requested
			if err == nil && pin {
				pinConfig := tgbotapi.PinChatMessageConfig{
					ChatID:              userID,
					MessageID:           sent.MessageID,
					DisableNotification: false,
				}
				_, _ = h.bot.Request(pinConfig)
			}
		}

		if sendErr != nil {
			errStr := sendErr.Error()
			if strings.Contains(errStr, "blocked") || strings.Contains(errStr, "deactivated") || strings.Contains(errStr, "not found") {
				blocked++
			} else {
				failed++
			}
			log.Printf("Broadcast to %d failed: %v", userID, sendErr)
		} else {
			success++
		}

		// Anti-flood: 25 messages per second max
		time.Sleep(40 * time.Millisecond)
	}

	// Update status message
	pinText := ""
	if pin {
		pinText = "\n📌 Messages pinned"
	}
	resultText := fmt.Sprintf(`✅ <b>Broadcast Complete!</b>%s

📊 <b>Results:</b>
✅ Sent: <b>%d</b>
❌ Failed: <b>%d</b>
🚫 Blocked: <b>%d</b>
👥 Total: <b>%d</b>`,
		pinText, success, failed, blocked, total)

	if statusMsg != nil {
		editMsg := tgbotapi.NewEditMessageText(msg.Chat.ID, statusMsg.MessageID, resultText)
		editMsg.ParseMode = "HTML"
		h.bot.Send(editMsg)
	} else {
		h.sendMsg(msg.Chat.ID, resultText, msg.MessageID)
	}
}
