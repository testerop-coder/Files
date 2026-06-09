package config

import (
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
)

type Config struct {
	BotToken      string
	APIId         string
	APIHash       string
	OwnerID       int64
	MongoURI      string
	DBChannelID   int64
	AdminIDs      []int64
}

func Load() *Config {
	_ = godotenv.Load(".env")

	ownerID, err := strconv.ParseInt(getEnv("OWNER_ID", "0"), 10, 64)
	if err != nil || ownerID == 0 {
		log.Fatal("❌ OWNER_ID is required and must be a valid integer")
	}

	dbChannelID, _ := strconv.ParseInt(getEnv("DB_CHANNEL_ID", "0"), 10, 64)

	cfg := &Config{
		BotToken:    getEnv("BOT_TOKEN", ""),
		APIId:       getEnv("API_ID", ""),
		APIHash:     getEnv("API_HASH", ""),
		OwnerID:     ownerID,
		MongoURI:    getEnv("MONGO_URI", "mongodb://localhost:27017"),
		DBChannelID: dbChannelID,
	}

	if cfg.BotToken == "" {
		log.Fatal("❌ BOT_TOKEN is required")
	}

	// Parse extra admin IDs from env
	adminStr := getEnv("ADMIN_IDS", "")
	if adminStr != "" {
		for _, s := range strings.Split(adminStr, ",") {
			id, err := strconv.ParseInt(strings.TrimSpace(s), 10, 64)
			if err == nil {
				cfg.AdminIDs = append(cfg.AdminIDs, id)
			}
		}
	}

	return cfg
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
