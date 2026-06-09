package database

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"time"

	"telegram-bot/models"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var client *mongo.Client
var db *mongo.Database

func Connect(uri string) (*mongo.Database, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	c, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	if err != nil {
		return nil, err
	}
	if err = c.Ping(ctx, nil); err != nil {
		return nil, err
	}
	client = c
	db = c.Database("fileproviderbot")
	ensureIndexes()
	return db, nil
}

func Disconnect() {
	if client != nil {
		_ = client.Disconnect(context.Background())
	}
}

func ensureIndexes() {
	ctx := context.Background()
	db.Collection("users").Indexes().CreateOne(ctx, mongo.IndexModel{Keys: bson.M{"user_id": 1}, Options: options.Index().SetUnique(true)})
	db.Collection("files").Indexes().CreateOne(ctx, mongo.IndexModel{Keys: bson.M{"unique_token": 1}, Options: options.Index().SetUnique(true)})
	db.Collection("batch_links").Indexes().CreateOne(ctx, mongo.IndexModel{Keys: bson.M{"unique_token": 1}, Options: options.Index().SetUnique(true)})
	db.Collection("admins").Indexes().CreateOne(ctx, mongo.IndexModel{Keys: bson.M{"user_id": 1}, Options: options.Index().SetUnique(true)})
}

// ─── USER ────────────────────────────────────────────────────────────────────

func UpsertUser(user *models.User) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	filter := bson.M{"user_id": user.UserID}
	update := bson.M{"$setOnInsert": user}
	opts := options.Update().SetUpsert(true)
	_, err := db.Collection("users").UpdateOne(ctx, filter, update, opts)
	return err
}

func CountUsers() (int64, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return db.Collection("users").CountDocuments(ctx, bson.M{})
}

func GetAllUserIDs() ([]int64, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	cursor, err := db.Collection("users").Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)
	var ids []int64
	for cursor.Next(ctx) {
		var u models.User
		if err := cursor.Decode(&u); err == nil {
			ids = append(ids, u.UserID)
		}
	}
	return ids, nil
}

// ─── FILE ────────────────────────────────────────────────────────────────────

func SaveFile(file *models.File) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_, err := db.Collection("files").InsertOne(ctx, file)
	return err
}

func GetFileByToken(token string) (*models.File, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	var f models.File
	err := db.Collection("files").FindOne(ctx, bson.M{"unique_token": token}).Decode(&f)
	if err != nil {
		return nil, err
	}
	return &f, nil
}

func GetFileByMsgID(msgID int) (*models.File, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	var f models.File
	err := db.Collection("files").FindOne(ctx, bson.M{"message_id": msgID}).Decode(&f)
	if err != nil {
		return nil, err
	}
	return &f, nil
}

// ─── BATCH ───────────────────────────────────────────────────────────────────

func SaveBatch(batch *models.BatchLink) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_, err := db.Collection("batch_links").InsertOne(ctx, batch)
	return err
}

func GetBatchByToken(token string) (*models.BatchLink, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	var b models.BatchLink
	err := db.Collection("batch_links").FindOne(ctx, bson.M{"unique_token": token}).Decode(&b)
	if err != nil {
		return nil, err
	}
	return &b, nil
}

// ─── ADMIN ───────────────────────────────────────────────────────────────────

func AddAdmin(admin *models.Admin) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_, err := db.Collection("admins").InsertOne(ctx, admin)
	return err
}

func RemoveAdmin(userID int64) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_, err := db.Collection("admins").DeleteOne(ctx, bson.M{"user_id": userID})
	return err
}

func GetAllAdmins() ([]models.Admin, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	cursor, err := db.Collection("admins").Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)
	var admins []models.Admin
	_ = cursor.All(ctx, &admins)
	return admins, nil
}

func IsAdmin(userID int64) (bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	count, err := db.Collection("admins").CountDocuments(ctx, bson.M{"user_id": userID})
	return count > 0, err
}

// ─── FSUB ────────────────────────────────────────────────────────────────────

func AddFSubChannel(ch *models.FSubChannel) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	filter := bson.M{"channel_id": ch.ChannelID}
	_, err := db.Collection("fsub_channels").ReplaceOne(ctx, filter, ch, options.Replace().SetUpsert(true))
	return err
}

func RemoveFSubChannel(channelID int64) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_, err := db.Collection("fsub_channels").DeleteOne(ctx, bson.M{"channel_id": channelID})
	return err
}

func GetAllFSubChannels() ([]models.FSubChannel, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	cursor, err := db.Collection("fsub_channels").Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)
	var channels []models.FSubChannel
	_ = cursor.All(ctx, &channels)
	return channels, nil
}

func UpdateFSubMode(channelID int64, mode string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_, err := db.Collection("fsub_channels").UpdateOne(ctx, bson.M{"channel_id": channelID}, bson.M{"$set": bson.M{"mode": mode}})
	return err
}

// ─── SETTINGS ────────────────────────────────────────────────────────────────

func SetSetting(key, value string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	filter := bson.M{"key": key}
	update := bson.M{"$set": bson.M{"key": key, "value": value, "updated_at": time.Now()}}
	_, err := db.Collection("settings").UpdateOne(ctx, filter, update, options.Update().SetUpsert(true))
	return err
}

func GetSetting(key string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	var s models.Settings
	err := db.Collection("settings").FindOne(ctx, bson.M{"key": key}).Decode(&s)
	if err != nil {
		return "", err
	}
	return s.Value, nil
}

func GetBotSettings() models.BotSettings {
	settings := models.BotSettings{
		AutoDeleteSeconds: 0,
		FSubExpireMinutes: 5,
	}

	if v, err := GetSetting("auto_delete_seconds"); err == nil {
		if n, err := strconv.Atoi(v); err == nil {
			settings.AutoDeleteSeconds = n
		}
	}
	if v, err := GetSetting("fsub_expire_minutes"); err == nil {
		if n, err := strconv.Atoi(v); err == nil {
			settings.FSubExpireMinutes = n
		}
	}
	return settings
}

// ─── HELPERS ─────────────────────────────────────────────────────────────────

func NewObjectID() primitive.ObjectID {
	return primitive.NewObjectID()
}

func LogError(context string, err error) {
	if err != nil {
		log.Printf("❌ [%s] %v", context, err)
	}
}

func FormatSize(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}
