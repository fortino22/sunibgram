package repository

import (
	"context"
	"fmt"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	chatmodel "github.com/sunibgram/domain/model/chat"
	chatmediamodel "github.com/sunibgram/domain/model/chat_media"
	chatreadmodel "github.com/sunibgram/domain/model/chat_read"
	chatroommodel "github.com/sunibgram/domain/model/chat_room"
	chatroommembermodel "github.com/sunibgram/domain/model/chat_room_member"
	chatroomrepoiface "github.com/sunibgram/domain/repository/chat_room"
	"github.com/sunibgram/infrastructure/cache"
)

type ChatRoomRepository struct {
	writeDB *gorm.DB
	readDB  *gorm.DB
	cache   *cache.Cache
}

func NewChatRoomRepository(writeDB, readDB *gorm.DB, c *cache.Cache) chatroomrepoiface.IChatRoomRepository {
	return &ChatRoomRepository{writeDB: writeDB, readDB: readDB, cache: c}
}

func (r *ChatRoomRepository) Save(room *chatroommodel.ChatRoom) error {
	return r.writeDB.Create(room).Error
}

func (r *ChatRoomRepository) FindByID(id uint) (*chatroommodel.ChatRoom, error) {
	var room chatroommodel.ChatRoom
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

func (r *ChatRoomRepository) FindByUserID(userID uint, page, limit int) ([]chatroommodel.ChatRoom, int64, error) {
	var total int64
	if err := r.readDB.Model(&chatroommembermodel.ChatRoomMember{}).Where("user_id = ?", userID).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var rooms []chatroommodel.ChatRoom
	offset := (page - 1) * limit
	err := r.readDB.Model(&chatroommodel.ChatRoom{}).
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
	if err := r.readDB.Model(&chatroommembermodel.ChatRoomMember{}).
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
	member := chatroommembermodel.ChatRoomMember{RoomID: roomID, UserID: userID, JoinedAt: time.Now()}
	result := r.writeDB.Clauses(clause.OnConflict{DoNothing: true}).Create(&member)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected > 0 {
		r.writeDB.Model(&chatroommodel.ChatRoom{}).Where("room_id = ?", roomID).UpdateColumn("member_count", gorm.Expr("member_count + 1"))
		_ = r.cache.Delete(context.Background(),
			fmt.Sprintf("chat:member:%d:%d", roomID, userID),
			fmt.Sprintf("chat:room:%d", roomID),
		)
	}
	return nil
}

func (r *ChatRoomRepository) UpdateEncryptedKey(roomID uint, encryptedKey string) error {
	err := r.writeDB.Model(&chatroommodel.ChatRoom{}).
		Where("room_id = ?", roomID).
		Update("encrypted_key", encryptedKey).Error
	if err != nil {
		return err
	}
	return r.cache.Delete(context.Background(), fmt.Sprintf("chat:room:%d", roomID))
}

func (r *ChatRoomRepository) RemoveMember(roomID, userID uint) error {
	ctx := context.Background()
	result := r.writeDB.Where("room_id = ? AND user_id = ?", roomID, userID).Delete(&chatroommembermodel.ChatRoomMember{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected > 0 {
		r.writeDB.Model(&chatroommodel.ChatRoom{}).Where("room_id = ?", roomID).
			UpdateColumn("member_count", gorm.Expr("GREATEST(member_count - 1, 0)"))
		_ = r.cache.Delete(ctx,
			fmt.Sprintf("chat:member:%d:%d", roomID, userID),
			fmt.Sprintf("chat:room:%d", roomID),
		)
	}
	return nil
}

func (r *ChatRoomRepository) UpdateGroup(roomID uint, name string, description *string, avatarURL *string) error {
	updates := map[string]any{
		"name":        name,
		"description": description,
	}
	if avatarURL != nil {
		updates["avatar_url"] = *avatarURL
	}
	err := r.writeDB.Model(&chatroommodel.ChatRoom{}).
		Where("room_id = ?", roomID).
		Updates(updates).Error
	if err != nil {
		return err
	}
	return r.cache.Delete(context.Background(), fmt.Sprintf("chat:room:%d", roomID))
}

func (r *ChatRoomRepository) DeleteGroup(roomID uint) error {
	ctx := context.Background()
	var userIDs []uint
	if err := r.readDB.Model(&chatroommembermodel.ChatRoomMember{}).
		Where("room_id = ?", roomID).
		Pluck("user_id", &userIDs).Error; err != nil {
		return err
	}

	err := r.writeDB.Transaction(func(tx *gorm.DB) error {
		chatIDs := tx.Model(&chatmodel.Chat{}).Select("chat_id").Where("room_id = ?", roomID)
		if err := tx.Where("chat_id IN (?)", chatIDs).Delete(&chatreadmodel.ChatRead{}).Error; err != nil {
			return err
		}
		if err := tx.Where("chat_id IN (?)", chatIDs).Delete(&chatmediamodel.ChatMedia{}).Error; err != nil {
			return err
		}
		if err := tx.Where("room_id = ?", roomID).Delete(&chatmodel.Chat{}).Error; err != nil {
			return err
		}
		if err := tx.Where("room_id = ?", roomID).Delete(&chatroommembermodel.ChatRoomMember{}).Error; err != nil {
			return err
		}
		return tx.Delete(&chatroommodel.ChatRoom{}, "room_id = ? AND type = ?", roomID, chatroommodel.ChatRoomTypeGroup).Error
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

func (r *ChatRoomRepository) FindMembers(roomID uint) ([]chatroomrepoiface.ChatRoomMemberInfo, error) {
	type row struct {
		UserID           uint
		Username         string
		Name             string
		FacultyName      string
		AvatarObjectName string
		AvatarBucket     string
	}
	var rows []row
	err := r.readDB.Raw(`
		SELECT u.user_id, u.user_name AS username, u.name,
		       COALESCE(f.faculty_name, '') AS faculty_name,
		       COALESCE(up.avatar_object_name, '') AS avatar_object_name,
		       COALESCE(up.avatar_bucket, '') AS avatar_bucket
		FROM chat_room_members crm
		JOIN users u ON u.user_id = crm.user_id
		LEFT JOIN user_profiles up ON up.user_id = u.user_id
		LEFT JOIN user_faculties uf ON uf.user_id = u.user_id
		LEFT JOIN faculties f ON f.faculty_id = uf.faculty_id
		WHERE crm.room_id = ?
		ORDER BY crm.joined_at ASC`,
		roomID,
	).Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	result := make([]chatroomrepoiface.ChatRoomMemberInfo, len(rows))
	for i, row := range rows {
		result[i] = chatroomrepoiface.ChatRoomMemberInfo{
			UserID:           row.UserID,
			Username:         row.Username,
			Name:             row.Name,
			FacultyName:      row.FacultyName,
			AvatarObjectName: row.AvatarObjectName,
			AvatarBucket:     row.AvatarBucket,
		}
	}
	return result, nil
}

func (r *ChatRoomRepository) FindOtherMemberInfo(roomIDs []uint, excludeUserID uint) (map[uint]chatroomrepoiface.ChatRoomMemberInfo, error) {
	if len(roomIDs) == 0 {
		return map[uint]chatroomrepoiface.ChatRoomMemberInfo{}, nil
	}
	type row struct {
		RoomID           uint
		UserID           uint
		Username         string
		Name             string
		FacultyName      string
		AvatarObjectName string
		AvatarBucket     string
	}
	var rows []row
	err := r.readDB.Raw(`
		SELECT crm.room_id, u.user_id, u.user_name AS username, u.name,
		       COALESCE(f.faculty_name, '') AS faculty_name,
		       COALESCE(up.avatar_object_name, '') AS avatar_object_name,
		       COALESCE(up.avatar_bucket, '') AS avatar_bucket
		FROM chat_room_members crm
		JOIN users u ON u.user_id = crm.user_id
		LEFT JOIN user_faculties uf ON uf.user_id = u.user_id
		LEFT JOIN faculties f ON f.faculty_id = uf.faculty_id
		LEFT JOIN user_profiles up ON up.user_id = u.user_id
		WHERE crm.room_id IN ? AND crm.user_id != ?`,
		roomIDs, excludeUserID,
	).Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	result := make(map[uint]chatroomrepoiface.ChatRoomMemberInfo, len(rows))
	for _, r := range rows {
		result[r.RoomID] = chatroomrepoiface.ChatRoomMemberInfo{
			UserID:           r.UserID,
			Username:         r.Username,
			Name:             r.Name,
			FacultyName:      r.FacultyName,
			AvatarObjectName: r.AvatarObjectName,
			AvatarBucket:     r.AvatarBucket,
		}
	}
	return result, nil
}

func (r *ChatRoomRepository) FindLastMessages(roomIDs []uint) (map[uint]chatroomrepoiface.LastMessageInfo, error) {
	if len(roomIDs) == 0 {
		return map[uint]chatroomrepoiface.LastMessageInfo{}, nil
	}
	type row struct {
		RoomID    uint
		ChatID    uint
		SenderID  uint
		Message   *string
		GifURL    *string
		HasMedia  bool
		CreatedAt time.Time
	}
	var rows []row
	err := r.readDB.Raw(`
		SELECT c.room_id, c.chat_id, c.sender_id, c.message, c.gif_url, c.created_at,
		       EXISTS(SELECT 1 FROM chat_media cm WHERE cm.chat_id = c.chat_id) AS has_media
		FROM chats c
		WHERE c.room_id IN ?
		  AND c.deleted_at IS NULL
		  AND c.chat_id = (
		      SELECT MAX(c2.chat_id) FROM chats c2
		      WHERE c2.room_id = c.room_id AND c2.deleted_at IS NULL
		  )`,
		roomIDs,
	).Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	result := make(map[uint]chatroomrepoiface.LastMessageInfo, len(rows))
	for _, row := range rows {
		result[row.RoomID] = chatroomrepoiface.LastMessageInfo{
			RoomID:    row.RoomID,
			ChatID:    row.ChatID,
			SenderID:  row.SenderID,
			Message:   row.Message,
			GifURL:    row.GifURL,
			HasMedia:  row.HasMedia,
			CreatedAt: row.CreatedAt,
		}
	}
	return result, nil
}

func (r *ChatRoomRepository) FindUnreadCounts(userID uint, roomIDs []uint) (map[uint]int, error) {
	if len(roomIDs) == 0 {
		return map[uint]int{}, nil
	}
	type row struct {
		RoomID      uint
		UnreadCount int
	}
	var rows []row
	err := r.readDB.Raw(`
		SELECT c.room_id, COUNT(*) AS unread_count
		FROM chats c
		JOIN chat_rooms r ON r.room_id = c.room_id
		WHERE c.room_id IN ?
		  AND c.sender_id != ?
		  AND c.deleted_at IS NULL
		  AND (
		      (r.type = 'private' AND c.is_read = false)
		      OR
		      (r.type = 'group' AND NOT EXISTS (
		          SELECT 1 FROM chat_reads cr
		          WHERE cr.chat_id = c.chat_id AND cr.user_id = ?
		      ))
		  )
		GROUP BY c.room_id`,
		roomIDs, userID, userID,
	).Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	result := make(map[uint]int, len(rows))
	for _, row := range rows {
		result[row.RoomID] = row.UnreadCount
	}
	return result, nil
}

func (r *ChatRoomRepository) GetUserRoomIDs(userID uint) ([]uint, error) {
	var ids []uint
	err := r.readDB.Model(&chatroommembermodel.ChatRoomMember{}).
		Where("user_id = ?", userID).
		Pluck("room_id", &ids).Error
	return ids, err
}

func (r *ChatRoomRepository) ToggleMute(roomID, userID uint) (bool, error) {
	var member chatroommembermodel.ChatRoomMember
	if err := r.readDB.Where("room_id = ? AND user_id = ?", roomID, userID).First(&member).Error; err != nil {
		return false, err
	}
	newState := !member.IsMuted
	err := r.writeDB.Model(&chatroommembermodel.ChatRoomMember{}).
		Where("room_id = ? AND user_id = ?", roomID, userID).
		UpdateColumn("is_muted", newState).Error
	return newState, err
}

func (r *ChatRoomRepository) GetMuteStatuses(userID uint, roomIDs []uint) (map[uint]bool, error) {
	if len(roomIDs) == 0 {
		return map[uint]bool{}, nil
	}
	type row struct {
		RoomID  uint
		IsMuted bool
	}
	var rows []row
	err := r.readDB.Model(&chatroommembermodel.ChatRoomMember{}).
		Where("user_id = ? AND room_id IN ?", userID, roomIDs).
		Select("room_id, is_muted").
		Find(&rows).Error
	if err != nil {
		return nil, err
	}
	result := make(map[uint]bool, len(rows))
	for _, row := range rows {
		result[row.RoomID] = row.IsMuted
	}
	return result, nil
}

func (r *ChatRoomRepository) TouchUpdatedAt(roomID uint) error {
	return r.writeDB.Model(&chatroommodel.ChatRoom{}).
		Where("room_id = ?", roomID).
		UpdateColumn("updated_at", gorm.Expr("NOW()")).Error
}

func (r *ChatRoomRepository) FindPrivateRoom(userA, userB uint) (*chatroommodel.ChatRoom, error) {
	var room chatroommodel.ChatRoom
	err := r.readDB.Table("chat_rooms").
		Select("chat_rooms.*").
		Joins("JOIN chat_room_members m1 ON m1.room_id = chat_rooms.room_id AND m1.user_id = ?", userA).
		Joins("JOIN chat_room_members m2 ON m2.room_id = chat_rooms.room_id AND m2.user_id = ?", userB).
		Where("chat_rooms.type = ?", chatroommodel.ChatRoomTypePrivate).
		First(&room).Error
	return &room, err
}
