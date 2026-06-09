package middleware

import (
	"telegram-bot/config"
	"telegram-bot/database"
)

// IsOwner checks if user is the bot owner
func IsOwner(userID int64, cfg *config.Config) bool {
	return userID == cfg.OwnerID
}

// IsAdmin checks if user is admin or owner
func IsAdmin(userID int64, cfg *config.Config) bool {
	if IsOwner(userID, cfg) {
		return true
	}
	// Check config admins
	for _, id := range cfg.AdminIDs {
		if id == userID {
			return true
		}
	}
	// Check DB admins
	ok, _ := database.IsAdmin(userID)
	return ok
}
