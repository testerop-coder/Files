package handlers

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"telegram-bot/database"
	"telegram-bot/middleware"
	"telegram-bot/models"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// ─── /addadmin ────────────────────────────────────────────────────────────────

func (h *Handler) handleAddAdmin(msg *tgbotapi.Message) {
	if !middleware.IsOwner(msg.From.ID, h.cfg) {
		h.sendMsg(msg.Chat.ID, "❌ Sirf <b>Owner</b> admin add kar sakta hai!", msg.MessageID)
		return
	}

	targetID := h.extractTargetUserID(msg)
	if targetID == 0 {
		h.sendMsg(msg.Chat.ID, `❌ User specify karein:
• Kisi message ko reply karein
• Ya <code>/addadmin USER_ID</code> use karein`, msg.MessageID)
		return
	}

	if middleware.IsOwner(targetID, h.cfg) {
		h.sendMsg(msg.Chat.ID, "ℹ️ Ye user already owner hai!", msg.MessageID)
		return
	}

	admin := &models.Admin{
		ID:      database.NewObjectID(),
		UserID:  targetID,
		AddedBy: msg.From.ID,
		AddedAt: time.Now(),
	}

	if err := database.AddAdmin(admin); err != nil {
		if strings.Contains(err.Error(), "duplicate") {
			h.sendMsg(msg.Chat.ID, "⚠️ Ye user already admin hai!", msg.MessageID)
			return
		}
		h.sendMsg(msg.Chat.ID, "❌ Admin add karne mein error.", msg.MessageID)
		return
	}

	h.sendMsg(msg.Chat.ID, fmt.Sprintf("✅ User <code>%d</code> ko Admin bana diya gaya!", targetID), msg.MessageID)
}

// ─── /removeadmin ─────────────────────────────────────────────────────────────

func (h *Handler) handleRemoveAdmin(msg *tgbotapi.Message) {
	if !middleware.IsOwner(msg.From.ID, h.cfg) {
		h.sendMsg(msg.Chat.ID, "❌ Sirf <b>Owner</b> admin remove kar sakta hai!", msg.MessageID)
		return
	}

	targetID := h.extractTargetUserID(msg)
	if targetID == 0 {
		h.sendMsg(msg.Chat.ID, `❌ User specify karein:
• Kisi message ko reply karein
• Ya <code>/removeadmin USER_ID</code> use karein`, msg.MessageID)
		return
	}

	if err := database.RemoveAdmin(targetID); err != nil {
		h.sendMsg(msg.Chat.ID, "❌ Admin remove karne mein error.", msg.MessageID)
		return
	}

	h.sendMsg(msg.Chat.ID, fmt.Sprintf("✅ User <code>%d</code> ko Admin list se hata diya!", targetID), msg.MessageID)
}

// ─── /admins ──────────────────────────────────────────────────────────────────

func (h *Handler) handleListAdmins(msg *tgbotapi.Message) {
	if !middleware.IsAdmin(msg.From.ID, h.cfg) {
		h.sendMsg(msg.Chat.ID, "❌ Access denied!", msg.MessageID)
		return
	}

	admins, err := database.GetAllAdmins()
	if err != nil {
		h.sendMsg(msg.Chat.ID, "❌ Admins fetch karne mein error.", msg.MessageID)
		return
	}

	var sb strings.Builder
	sb.WriteString("👮 <b>Bot Admins List</b>\n\n")

	// Owner
	sb.WriteString(fmt.Sprintf("👑 <b>Owner:</b> <code>%d</code>\n\n", h.cfg.OwnerID))

	if len(admins) == 0 {
		sb.WriteString("📭 Koi admin nahi hai abhi.")
	} else {
		sb.WriteString(fmt.Sprintf("📋 <b>Total Admins: %d</b>\n", len(admins)))
		for i, a := range admins {
			name := a.Username
			if name == "" {
				name = fmt.Sprintf("ID: %d", a.UserID)
			} else {
				name = "@" + name
			}
			sb.WriteString(fmt.Sprintf("%d. %s (<code>%d</code>)\n", i+1, name, a.UserID))
		}
	}

	h.sendMsg(msg.Chat.ID, sb.String(), msg.MessageID)
}

// ─── /setdelete ───────────────────────────────────────────────────────────────

func (h *Handler) handleSetAutoDelete(msg *tgbotapi.Message) {
	if !middleware.IsAdmin(msg.From.ID, h.cfg) {
		h.sendMsg(msg.Chat.ID, "❌ Sirf admins ye command use kar sakte hain!", msg.MessageID)
		return
	}

	args := cleanArgs(msg.CommandArguments())
	if args == "" {
		settings := database.GetBotSettings()
		current := "Off"
		if settings.AutoDeleteSeconds > 0 {
			current = fmt.Sprintf("%d seconds", settings.AutoDeleteSeconds)
		}
		h.sendMsg(msg.Chat.ID, fmt.Sprintf(`⏱ <b>Auto Delete Settings</b>

Current: <b>%s</b>

Usage: <code>/setdelete SECONDS</code>
Examples:
• <code>/setdelete 300</code> - 5 minutes
• <code>/setdelete 600</code> - 10 minutes  
• <code>/setdelete 3600</code> - 1 hour
• <code>/setdelete 0</code> - Disable`, current), msg.MessageID)
		return
	}

	seconds, err := strconv.Atoi(args)
	if err != nil || seconds < 0 {
		h.sendMsg(msg.Chat.ID, "❌ Valid seconds enter karein (0 = disable)", msg.MessageID)
		return
	}

	if err := database.SetSetting("auto_delete_seconds", strconv.Itoa(seconds)); err != nil {
		h.sendMsg(msg.Chat.ID, "❌ Setting save karne mein error.", msg.MessageID)
		return
	}

	if seconds == 0 {
		h.sendMsg(msg.Chat.ID, "✅ Auto Delete <b>Disabled</b> kar diya gaya!", msg.MessageID)
	} else {
		h.sendMsg(msg.Chat.ID, fmt.Sprintf("✅ Auto Delete <b>%d seconds</b> set kar diya!", seconds), msg.MessageID)
	}
}

// ─── HELPER ───────────────────────────────────────────────────────────────────

func (h *Handler) extractTargetUserID(msg *tgbotapi.Message) int64 {
	// From reply
	if msg.ReplyToMessage != nil {
		return msg.ReplyToMessage.From.ID
	}

	// From text_mention entity
	for _, e := range msg.Entities {
		if e.Type == "text_mention" && e.User != nil {
			return e.User.ID
		}
	}

	// From args as user ID
	args := cleanArgs(msg.CommandArguments())
	if args != "" {
		id, err := strconv.ParseInt(args, 10, 64)
		if err == nil {
			return id
		}
	}

	return 0
}
