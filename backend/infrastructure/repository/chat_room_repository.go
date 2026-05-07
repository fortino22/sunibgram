package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/sunibgram/domain/model"
	domainrepo "github.com/sunibgram/domain/repository"
	"github.com/sunibgram/infrastructure/cache"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type ChatRoomRepository struct {
	writeDB *gorm.DB
	readDB  *gorm.DB
	cache   *cache.Cache
}

func NewChatRoomRepository(writeDB, readDB *gorm.DB, c *cache.Cache) domainrepo.IChatRoomRepository {
	return &ChatRoomRepository{writeDB: writeDB, readDB: readDB, cache: c}
}

func (r *ChatRoomRepository) Save(room *model.ChatRoom) error {
	return r.writeDB.Create(room).Error
}

func (r *ChatRoomRepository) FindByID(id uint) (*model.ChatRoom, error) {
	var room model.ChatRoom
	ctx := context.Background()
	key := fmt.Sprintf("chat:room:%d", id)
	if err := r.cache.Get(ctx, key, &room); err == nil {
		return &room, nil
	}

	err := r.readDB.First(&room, id).Error
	if err == nil {
		_ = r.cache.Set(ctx, key, room, cache.DefaultTTL)
	}
	return &room, err
}

func (r *ChatRoomRepository) FindByUserID(userID uint, page, limit int) ([]model.ChatRoom, int64, error) {
	var total int64
	if err := r.readDB.Model(&model.ChatRoomMember{}).Where("user_id = ?", userID).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var rooms []model.ChatRoom
	offset := (page - 1) * limit
	err := r.readDB.Model(&model.ChatRoom{}).
		Joins("JOIN chat_room_members ON chat_room_members.room_id = chat_rooms.room_id").
		Where("chat_room_members.user_id = ?", userID).
		Order("chat_rooms.updated_at DESC").
		Offset(offset).Limit(limit).
		Find(&rooms).Error
	return rooms, total, err
}

func (r *ChatRoomRepository) IsMember(roomID, userID uint) bool {
	ctx := context.Background()
	key := fmt.Sprintf("chat:member:%d:%d", roomID, userID)
	var cached bool
	if err := r.cache.Get(ctx, key, &cached); err == nil {
		return cached
	}

	var count int64
	if err := r.readDB.Model(&model.ChatRoomMember{}).
		Where("room_id = ? AND user_id = ?", roomID, userID).
		Count(&count).Error; err != nil {
		return false
	}
	isMember := count > 0
	_ = r.cache.Set(ctx, key, isMember, cache.DefaultTTL)
	return isMember
}

func (r *ChatRoomRepository) CountMembers(roomID uint) int64 {
	var count int64
	if err := r.readDB.Table("chat_room_members").
		Where("room_id = ?", roomID).
		Count(&count).Error; err != nil {
		return 0
	}
	return count
}

func (r *ChatRoomRepository) AddMember(roomID, userID uint) error {
	member := model.ChatRoomMember{RoomID: roomID, UserID: userID, JoinedAt: time.Now()}
	result := r.writeDB.Clauses(clause.OnConflict{DoNothing: true}).Create(&member)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected > 0 {
		r.writeDB.Model(&model.ChatRoom{}).Where("room_id = ?", roomID).UpdateColumn("member_count", gorm.Expr("member_count + 1"))
		_ = r.cache.Delete(context.Background(),
			fmt.Sprintf("chat:member:%d:%d", roomID, userID),
			fmt.Sprintf("chat:room:%d", roomID),
		)
	}
	return nil
}

func (r *ChatRoomRepository) FindPrivateRoom(userA, userB uint) (*model.ChatRoom, error) {
	var room model.ChatRoom
	err := r.readDB.Table("chat_rooms").
		Select("chat_rooms.*").
		Joins("JOIN chat_room_members m1 ON m1.room_id = chat_rooms.room_id AND m1.user_id = ?", userA).
		Joins("JOIN chat_room_members m2 ON m2.room_id = chat_rooms.room_id AND m2.user_id = ?", userB).
		Where("chat_rooms.type = ?", model.ChatRoomTypePrivate).
		First(&room).Error
	return &room, err
}
