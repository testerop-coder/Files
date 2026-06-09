#!/bin/bash
# ─────────────────────────────────────────────
# File Provider Bot - VPS Deploy Script
# ─────────────────────────────────────────────

set -e

BOT_DIR="/home/ubuntu/telegram-bot"
SERVICE_NAME="telegram-bot"

echo "🚀 File Provider Bot - Deploy Script"
echo "======================================"

# ── 1. Install Go if not present ──────────────
if ! command -v go &> /dev/null; then
    echo "📦 Installing Go..."
    wget -q https://go.dev/dl/go1.21.6.linux-amd64.tar.gz
    sudo tar -C /usr/local -xzf go1.21.6.linux-amd64.tar.gz
    echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc
    export PATH=$PATH:/usr/local/go/bin
    rm go1.21.6.linux-amd64.tar.gz
    echo "✅ Go installed: $(go version)"
else
    echo "✅ Go already installed: $(go version)"
fi

# ── 2. Install MongoDB if not present ─────────
if ! command -v mongod &> /dev/null; then
    echo "📦 Installing MongoDB..."
    curl -fsSL https://www.mongodb.org/static/pgp/server-7.0.asc | sudo gpg -o /usr/share/keyrings/mongodb-server-7.0.gpg --dearmor
    echo "deb [ arch=amd64,arm64 signed-by=/usr/share/keyrings/mongodb-server-7.0.gpg ] https://repo.mongodb.org/apt/ubuntu jammy/mongodb-org/7.0 multiverse" | sudo tee /etc/apt/sources.list.d/mongodb-org-7.0.list
    sudo apt-get update -qq
    sudo apt-get install -y mongodb-org
    sudo systemctl start mongod
    sudo systemctl enable mongod
    echo "✅ MongoDB installed and started"
else
    echo "✅ MongoDB already installed"
fi

# ── 3. Build the bot ──────────────────────────
echo "🔨 Building bot..."
cd "$BOT_DIR"

# Download dependencies
go mod tidy
go mod download

# Build binary
go build -o telegram-bot ./main.go
echo "✅ Bot built successfully"

# ── 4. Setup .env if not exists ───────────────
if [ ! -f "$BOT_DIR/.env" ]; then
    cp "$BOT_DIR/.env.example" "$BOT_DIR/.env"
    echo ""
    echo "⚠️  .env file created from example!"
    echo "📝 Edit it now: nano $BOT_DIR/.env"
    echo "   Fill in: BOT_TOKEN, API_ID, API_HASH, OWNER_ID, DB_CHANNEL_ID"
    echo ""
    read -p "Press Enter after editing .env to continue..."
fi

# ── 5. Install systemd service ────────────────
echo "⚙️  Installing systemd service..."
sudo cp "$BOT_DIR/telegram-bot.service" /etc/systemd/system/
sudo systemctl daemon-reload
sudo systemctl enable $SERVICE_NAME
sudo systemctl restart $SERVICE_NAME

echo ""
echo "✅ Bot deployed and started!"
echo ""
echo "📋 Useful commands:"
echo "  Status:  sudo systemctl status $SERVICE_NAME"
echo "  Logs:    sudo journalctl -u $SERVICE_NAME -f"
echo "  Restart: sudo systemctl restart $SERVICE_NAME"
echo "  Stop:    sudo systemctl stop $SERVICE_NAME"
