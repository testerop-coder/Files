package handlers

import (
	"fmt"
	"runtime"
	"time"

	"telegram-bot/database"
	"telegram-bot/middleware"
	"telegram-bot/utils"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// ─── /status ─────────────────────────────────────────────────────────────────

func (h *Handler) handleStatus(msg *tgbotapi.Message) {
	if !middleware.IsAdmin(msg.From.ID, h.cfg) {
		h.sendMsg(msg.Chat.ID, "❌ Sirf admins status dekh sakte hain!", msg.MessageID)
		return
	}

	// Ping calculation
	start := time.Now()
	pingMsg := h.sendMsg(msg.Chat.ID, "⏳ Calculating...", 0)
	ping := time.Since(start).Milliseconds()

	// Stats
	userCount, _ := database.CountUsers()
	uptime := utils.FormatDuration(time.Since(h.startTime))
	settings := database.GetBotSettings()

	// Memory stats
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)
	memUsage := float64(memStats.Alloc) / 1024 / 1024

	// FSub channels
	fsubChannels, _ := database.GetAllFSubChannels()
	admins, _ := database.GetAllAdmins()

	autoDelete := "❌ Off"
	if settings.AutoDeleteSeconds > 0 {
		autoDelete = fmt.Sprintf("✅ %d seconds", settings.AutoDeleteSeconds)
	}

	text := fmt.Sprintf(`📊 <b>Bot Status</b>

🤖 <b>Bot:</b> @%s
⚡ <b>Ping:</b> %dms
⏱ <b>Uptime:</b> %s

👥 <b>Total Users:</b> %d
👮 <b>Admins:</b> %d
📢 <b>FSub Channels:</b> %d

⏳ <b>Auto Delete:</b> %s
🔗 <b>FSub Expire:</b> %d minutes

💾 <b>Memory:</b> %.2f MB
🖥 <b>Go Version:</b> %s
🌐 <b>Platform:</b> %s/%s`,
		h.bot.Self.UserName,
		ping,
		uptime,
		userCount,
		len(admins)+1, // +1 for owner
		len(fsubChannels),
		autoDelete,
		settings.FSubExpireMinutes,
		memUsage,
		runtime.Version(),
		runtime.GOOS,
		runtime.GOARCH,
	)

	if pingMsg != nil {
		editMsg := tgbotapi.NewEditMessageText(msg.Chat.ID, pingMsg.MessageID, text)
		editMsg.ParseMode = "HTML"
		h.bot.Send(editMsg)
	} else {
		h.sendMsg(msg.Chat.ID, text, msg.MessageID)
	}
}

// ─── /ping ───────────────────────────────────────────────────────────────────

func (h *Handler) handlePing(msg *tgbotapi.Message) {
	start := time.Now()
	sent := h.sendMsg(msg.Chat.ID, "🏓 Pong!", 0)
	ping := time.Since(start).Milliseconds()

	if sent != nil {
		editMsg := tgbotapi.NewEditMessageText(msg.Chat.ID, sent.MessageID,
			fmt.Sprintf("🏓 <b>Pong!</b>\n⚡ <b>%dms</b>", ping))
		editMsg.ParseMode = "HTML"
		h.bot.Send(editMsg)
	}
}

// ─── /help ───────────────────────────────────────────────────────────────────

func (h *Handler) handleHelp(msg *tgbotapi.Message) {
	isAdmin := middleware.IsAdmin(msg.From.ID, h.cfg)
	isOwner := middleware.IsOwner(msg.From.ID, h.cfg)

	userHelp := `📖 <b>File Provider Bot - Help</b>

<b>User Commands:</b>
/start - Bot start karein
/ping - Bot ki speed check karein
/help - Ye help message`

	adminHelp := `

<b>File Commands:</b>
/getlink - Single file ka shareable link banao
/batch - Multiple files ka batch link banao

<b>Settings:</b>
/setdelete [seconds] - Auto delete time set karein
/setexpire [minutes] - FSub link expire time set karein

<b>FSub Management:</b>
/addfsub [channel_id] - FSub channel add karein
/removefsub [channel_id] - FSub channel hatain
/fsubs - FSub channels ki list
/fsubmode [id] [normal/request] - Mode set karein

<b>Broadcast:</b>
/broadcast [message] - Sabko message bhejein
/pbroadcast [message] - Pin karke message bhejein

<b>Admin Management:</b>
/admins - Admins ki list
/status - Bot ka status dekho
/ping - Ping check`

	ownerHelp := `

<b>Owner Only:</b>
/addadmin [user_id/reply] - Admin add karein
/removeadmin [user_id/reply] - Admin hatain`

	text := userHelp
	if isAdmin {
		text += adminHelp
	}
	if isOwner {
		text += ownerHelp
	}

	h.sendMsg(msg.Chat.ID, text, msg.MessageID)
}
