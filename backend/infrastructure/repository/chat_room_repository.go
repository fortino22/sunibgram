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

func (r *ChatRoomRepository) UpdateGroup(roomID uint, name string, description *string) error {
	err := r.writeDB.Model(&model.ChatRoom{}).
		Where("room_id = ?", roomID).
		Updates(map[string]any{"name": name, "description": description}).Error
	if err != nil {
		return err
	}
	return r.cache.Delete(context.Background(), fmt.Sprintf("chat:room:%d", roomID))
}

func (r *ChatRoomRepository) DeleteGroup(roomID uint) error {
	ctx := context.Background()
	var userIDs []uint
	if err := r.readDB.Model(&model.ChatRoomMember{}).
		Where("room_id = ?", roomID).
		Pluck("user_id", &userIDs).Error; err != nil {
		return err
	}

	err := r.writeDB.Transaction(func(tx *gorm.DB) error {
		chatIDs := tx.Model(&model.Chat{}).Select("chat_id").Where("room_id = ?", roomID)
		if err := tx.Where("chat_id IN (?)", chatIDs).Delete(&model.ChatRead{}).Error; err != nil {
			return err
		}
		if err := tx.Where("chat_id IN (?)", chatIDs).Delete(&model.ChatMedia{}).Error; err != nil {
			return err
		}
		if err := tx.Where("room_id = ?", roomID).Delete(&model.Chat{}).Error; err != nil {
			return err
		}
		if err := tx.Where("room_id = ?", roomID).Delete(&model.ChatRoomMember{}).Error; err != nil {
			return err
		}
		return tx.Delete(&model.ChatRoom{}, "room_id = ? AND type = ?", roomID, model.ChatRoomTypeGroup).Error
	})
	if err != nil {
		return err
	}

	keys := []string{fmt.Sprintf("chat:room:%d", roomID)}
	for _, userID := range userIDs {
		keys = append(keys, fmt.Sprintf("chat:member:%d:%d", roomID, userID))
	}
	return r.cache.Delete(ctx, keys...)
}

func (r *ChatRoomRepository) FindOtherMemberInfo(roomIDs []uint, excludeUserID uint) (map[uint]domainrepo.ChatRoomMemberInfo, error) {
	if len(roomIDs) == 0 {
		return map[uint]domainrepo.ChatRoomMemberInfo{}, nil
	}
	type row struct {
		RoomID      uint
		UserID      uint
		Username    string
		Name        string
		FacultyName string
	}
	var rows []row
	err := r.readDB.Raw(`
		SELECT crm.room_id, u.user_id, u.user_name AS username, u.name,
		       COALESCE(f.faculty_name, '') AS faculty_name
		FROM chat_room_members crm
		JOIN users u ON u.user_id = crm.user_id
		LEFT JOIN user_faculties uf ON uf.user_id = u.user_id
		LEFT JOIN faculties f ON f.faculty_id = uf.faculty_id
		WHERE crm.room_id IN ? AND crm.user_id != ?`,
		roomIDs, excludeUserID,
	).Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	result := make(map[uint]domainrepo.ChatRoomMemberInfo, len(rows))
	for _, r := range rows {
		result[r.RoomID] = domainrepo.ChatRoomMemberInfo{
			UserID:      r.UserID,
			Username:    r.Username,
			Name:        r.Name,
			FacultyName: r.FacultyName,
		}
	}
	return result, nil
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
