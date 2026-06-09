package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"telegram-bot/config"
	"telegram-bot/database"
	"telegram-bot/handlers"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

var startTime = time.Now()

func main() {
	cfg := config.Load()

	// Connect MongoDB
	db, err := database.Connect(cfg.MongoURI)
	if err != nil {
		log.Fatalf("MongoDB connection failed: %v", err)
	}
	defer database.Disconnect()
	log.Println("✅ MongoDB connected")

	// Init Bot
	bot, err := tgbotapi.NewBotAPI(cfg.BotToken)
	if err != nil {
		log.Fatalf("Bot init failed: %v", err)
	}
	bot.Debug = false
	log.Printf("✅ Bot started: @%s", bot.Self.UserName)

	// Init Handlers
	h := handlers.New(bot, db, cfg, startTime)

	// Start update polling
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	updates := bot.GetUpdatesChan(u)

	// Graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	log.Println("🚀 Bot is running...")

	for {
		select {
		case update := <-updates:
			go h.HandleUpdate(update)
		case <-sigChan:
			log.Println("Shutting down...")
			return
		}
	}
}
