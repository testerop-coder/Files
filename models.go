package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// User model
type User struct {
	ID        primitive.ObjectID `bson:"_id,omitempty"`
	UserID    int64              `bson:"user_id"`
	Username  string             `bson:"username"`
	FirstName string             `bson:"first_name"`
	JoinedAt  time.Time          `bson:"joined_at"`
	IsBanned  bool               `bson:"is_banned"`
}

// File model - stores file info from DB channel
type File struct {
	ID          primitive.ObjectID `bson:"_id,omitempty"`
	FileID      string             `bson:"file_id"`
	MessageID   int                `bson:"message_id"`
	FileName    string             `bson:"file_name"`
	FileSize    int64              `bson:"file_size"`
	FileType    string             `bson:"file_type"` // document, video, photo, audio
	UniqueToken string             `bson:"unique_token"`
	AddedBy     int64              `bson:"added_by"`
	AddedAt     time.Time          `bson:"added_at"`
}

// BatchLink model - stores batch of files
type BatchLink struct {
	ID          primitive.ObjectID `bson:"_id,omitempty"`
	UniqueToken string             `bson:"unique_token"`
	StartMsgID  int                `bson:"start_msg_id"`
	EndMsgID    int                `bson:"end_msg_id"`
	CreatedBy   int64              `bson:"created_by"`
	CreatedAt   time.Time          `bson:"created_at"`
}

// Admin model
type Admin struct {
	ID       primitive.ObjectID `bson:"_id,omitempty"`
	UserID   int64              `bson:"user_id"`
	Username string             `bson:"username"`
	AddedBy  int64              `bson:"added_by"`
	AddedAt  time.Time          `bson:"added_at"`
}

// FSub channel model
type FSubChannel struct {
	ID          primitive.ObjectID `bson:"_id,omitempty"`
	ChannelID   int64              `bson:"channel_id"`
	ChannelName string             `bson:"channel_name"`
	Mode        string             `bson:"mode"` // "normal" or "request"
	AddedBy     int64              `bson:"added_by"`
	AddedAt     time.Time          `bson:"added_at"`
}

// Settings model
type Settings struct {
	ID              primitive.ObjectID `bson:"_id,omitempty"`
	Key             string             `bson:"key"`
	Value           string             `bson:"value"`
	UpdatedAt       time.Time          `bson:"updated_at"`
}

// BotSettings aggregated
type BotSettings struct {
	AutoDeleteSeconds int    // 0 = disabled
	FSubExpireMinutes int    // default 5
}
