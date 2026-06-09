package utils

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"strings"
	"time"
)

// GenerateToken creates a unique URL-safe token
func GenerateToken(length int) string {
	b := make([]byte, length)
	_, _ = rand.Read(b)
	token := base64.URLEncoding.EncodeToString(b)
	token = strings.ReplaceAll(token, "=", "")
	if len(token) > length {
		token = token[:length]
	}
	return token
}

// FormatDuration formats seconds to human-readable
func FormatDuration(d time.Duration) string {
	days := int(d.Hours()) / 24
	hours := int(d.Hours()) % 24
	minutes := int(d.Minutes()) % 60
	seconds := int(d.Seconds()) % 60

	if days > 0 {
		return fmt.Sprintf("%dd %dh %dm", days, hours, minutes)
	}
	if hours > 0 {
		return fmt.Sprintf("%dh %dm %ds", hours, minutes, seconds)
	}
	if minutes > 0 {
		return fmt.Sprintf("%dm %ds", minutes, seconds)
	}
	return fmt.Sprintf("%ds", seconds)
}

// ParseMentionOrID extracts user ID from message entities
func MentionToText(firstName, username string, userID int64) string {
	if username != "" {
		return fmt.Sprintf("@%s", username)
	}
	return fmt.Sprintf("[%s](tg://user?id=%d)", firstName, userID)
}

// EscapeMarkdown escapes special MarkdownV2 chars
func EscapeMarkdown(s string) string {
	replacer := strings.NewReplacer(
		"_", "\\_", "*", "\\*", "[", "\\[", "]", "\\]",
		"(", "\\(", ")", "\\)", "~", "\\~", "`", "\\`",
		">", "\\>", "#", "\\#", "+", "\\+", "-", "\\-",
		"=", "\\=", "|", "\\|", "{", "\\{", "}", "\\}",
		".", "\\.", "!", "\\!",
	)
	return replacer.Replace(s)
}

// FileTypeEmoji returns emoji for file type
func FileTypeEmoji(fileType string) string {
	switch fileType {
	case "video":
		return "🎬"
	case "audio":
		return "🎵"
	case "photo":
		return "🖼️"
	case "document":
		return "📄"
	default:
		return "📁"
	}
}
