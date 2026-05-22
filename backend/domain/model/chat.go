package model

import (
	"errors"
	"net/url"
	"strings"
	"time"
)

func ValidateGifURL(raw string) error {
	u, err := url.ParseRequestURI(raw)
	if err != nil || u.Scheme != "https" {
		return errors.New("gif_url must be a valid HTTPS URL")
	}
	host := strings.ToLower(u.Hostname())
	if host == "giphy.com" || strings.HasSuffix(host, ".giphy.com") {
		return nil
	}
	return errors.New("gif_url must be from a supported provider (Giphy)")
}

type Chat struct {
	ChatID    uint     `gorm:"primaryKey;autoIncrement"`
	RoomID    uint     `gorm:"not null;index;index:idx_chat_room_deleted_created,priority:1;index:idx_chat_room_deleted,priority:1"`
	Room      ChatRoom `gorm:"foreignKey:RoomID;references:RoomID;constraint:OnUpdate:NO ACTION,OnDelete:NO ACTION"`
	SenderID  uint     `gorm:"not null"`
	Sender    User     `gorm:"foreignKey:SenderID;references:UserID;constraint:OnUpdate:NO ACTION,OnDelete:NO ACTION"`
	ReplyToID *uint    `gorm:"index"`
	ReplyTo   *Chat    `gorm:"foreignKey:ReplyToID;references:ChatID"`
	Message   *string  `gorm:"type:text"`
	GifURL    *string  `gorm:"type:text"`
	IsRead    bool     `gorm:"not null;default:false"`
	IsEdited  bool     `gorm:"not null;default:false"`
	EditedAt  *time.Time
	DeletedAt *time.Time `gorm:"index;index:idx_chat_room_deleted_created,priority:2;index:idx_chat_room_deleted,priority:2"`
	CreatedAt time.Time  `gorm:"index:idx_chat_room_deleted_created,priority:3,sort:desc;index:idx_chat_room_deleted,priority:3,sort:desc"`
}
