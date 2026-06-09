# 🤖 File Provider Telegram Bot (Go)

Ek powerful Telegram bot jo private channel se files share karne ki facility deta hai.

---

## ✨ Features

| Feature | Status |
|---|---|
| `/getlink` - Single file link | ✅ |
| `/batch` - Multiple files batch link | ✅ |
| DB Channel se URL copy (auto-index) | ✅ |
| Auto Delete (admin set kare) | ✅ |
| Owner → Admin add/remove | ✅ |
| FSub channel add/remove/show | ✅ |
| FSub Normal + Request mode | ✅ |
| FSub invite link expiry | ✅ |
| FSub expire time bot se set | ✅ |
| Status (users, ping, uptime) | ✅ |
| 24hr auto-restart (systemd) | ✅ |
| Broadcast + Pin Broadcast | ✅ |
| MongoDB + API credentials | ✅ |

---

## 📁 Project Structure

```
telegram-bot/
├── main.go                 # Entry point
├── go.mod                  # Dependencies
├── .env.example            # Config template
├── telegram-bot.service    # Systemd service
├── deploy.sh               # VPS deploy script
├── config/
│   └── config.go           # Config loader
├── database/
│   └── database.go         # MongoDB operations
├── handlers/
│   ├── handler.go          # Main router
│   ├── files.go            # /getlink, /batch, file delivery
│   ├── admin.go            # Admin management, auto-delete
│   ├── fsub.go             # FSub channels management
│   ├── broadcast.go        # /broadcast, /pbroadcast
│   ├── status.go           # /status, /ping, /help
│   └── callback.go         # Inline button callbacks
├── middleware/
│   └── auth.go             # Auth checks
├── models/
│   └── models.go           # MongoDB models
└── utils/
    └── utils.go            # Helper functions
```

---

## 🚀 Setup Guide

### Step 1: Credentials lein

1. **Bot Token**: [@BotFather](https://t.me/BotFather) se naya bot banao → token copy karein
2. **API ID & Hash**: [my.telegram.org](https://my.telegram.org) → App → API credentials
3. **Owner ID**: [@userinfobot](https://t.me/userinfobot) ko message karein
4. **DB Channel**: Ek private channel banao → bot ko admin banao → Channel ID copy karein

### Step 2: DB Channel ID kaise milega?

1. Channel mein koi bhi message forward karein [@JsonDumpBot](https://t.me/JsonDumpBot) ko
2. `forward_from_chat.id` field mein ID milegi (e.g. `-1001234567890`)

### Step 3: Local Setup

```bash
# Clone/copy project
cd telegram-bot

# .env file banao
cp .env.example .env
nano .env   # apni values fill karein

# Dependencies install karein
go mod tidy

# Run karein
go run main.go
```

### Step 4: VPS Deploy (Auto-start + 24hr restart)

```bash
# Project VPS pe copy karein
scp -r telegram-bot/ ubuntu@YOUR_VPS_IP:/home/ubuntu/

# VPS pe login karein
ssh ubuntu@YOUR_VPS_IP

# Deploy script run karein
cd telegram-bot
chmod +x deploy.sh
./deploy.sh
```

---

## 📋 Commands Reference

### 👤 User Commands
| Command | Description |
|---|---|
| `/start` | Bot start, file/batch receive |
| `/ping` | Speed check |
| `/help` | Help message |

### 👮 Admin Commands
| Command | Description |
|---|---|
| `/getlink` | File ko reply karke single link banao |
| `/batch` | First aur last URL se batch link banao |
| `/setdelete [seconds]` | Auto delete time (0 = off) |
| `/setexpire [minutes]` | FSub link expire time |
| `/addfsub [channel_id]` | FSub channel add |
| `/removefsub [channel_id]` | FSub channel remove |
| `/fsubs` | FSub channels list |
| `/fsubmode [id] [normal/request]` | Mode change |
| `/broadcast [text]` | Sabko message bhejo |
| `/pbroadcast [text]` | Pin karke broadcast |
| `/admins` | Admins list |
| `/status` | Full bot status |

### 👑 Owner Commands
| Command | Description |
|---|---|
| `/addadmin [id/reply]` | Admin add karein |
| `/removeadmin [id/reply]` | Admin remove karein |

---

## 🔄 Batch Link Usage

1. Admin `/batch` command bhejein
2. DB Channel se **pehli file** forward karein
3. DB Channel se **aakhri file** forward karein
4. Bot automatically batch link generate karega
5. Link share karein → users ko saari files milegi

---

## 📢 FSub Modes

| Mode | Description |
|---|---|
| `normal` | Direct invite link — user click karke join kare |
| `request` | Join request link — bot automatically approve kare |

Links automatically expire ho jaate hain (default: 5 minutes, bot se change karein)

---

## ⚙️ Systemd Service (24hr Auto-restart)

```bash
# Status check
sudo systemctl status telegram-bot

# Live logs
sudo journalctl -u telegram-bot -f

# Manual restart
sudo systemctl restart telegram-bot

# 24hr RuntimeMaxSec already set hai service file mein
```

---

## 🛠 Troubleshooting

**Bot respond nahi kar raha:**
```bash
sudo journalctl -u telegram-bot -n 50
```

**MongoDB connect nahi ho raha:**
```bash
sudo systemctl status mongod
sudo systemctl start mongod
```

**Bot channel admin nahi:**
> Channel settings → Administrators → Bot add karein → All permissions dein
