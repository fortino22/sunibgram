package seeder

import (
	"log"
	"strings"
	"time"

	"gorm.io/gorm"

	chatmodel "github.com/sunibgram/domain/model/chat"
	chatroommodel "github.com/sunibgram/domain/model/chat_room"
	chatroommembermodel "github.com/sunibgram/domain/model/chat_room_member"
	usermodel "github.com/sunibgram/domain/model/user"
)

type demoChatMessage struct {
	Sender string
	Text   string
}

type demoChatRoom struct {
	Name     string // "" for private (1:1)
	Type     chatroommodel.ChatRoomType
	Members  []string
	Messages []demoChatMessage
}

var demoChatRooms = []demoChatRoom{
	{
		Type: chatroommodel.ChatRoomTypePrivate,
		Members: []string{"devin.hi", "Octa.hi"},
		Messages: []demoChatMessage{
			{"devin.hi", "yo are you going to the bootcamp saturday?"},
			{"Octa.hi", "I'm literally running it lol"},
			{"devin.hi", "oh. right. see you there then 😭"},
			{"Octa.hi", "save me a seat in the front"},
		},
	},
	{
		Name: "BNCC Core Team", Type: chatroommodel.ChatRoomTypeGroup,
		Members: []string{"Octa.hi", "irfan.hi", "hana.hi", "devin.hi"},
		Messages: []demoChatMessage{
			{"Octa.hi", "ok team, bootcamp prep checklist in the doc"},
			{"hana.hi", "slides are 80% done, will finish tonight"},
			{"irfan.hi", "backend demo repo is ready 👍"},
			{"devin.hi", "I'll handle registration desk"},
			{"Octa.hi", "legends. let's gooo"},
		},
	},
	{
		Name: "Photo Walk Squad", Type: chatroommodel.ChatRoomTypeGroup,
		Members: []string{"gilang.hi", "dina.hi", "chermiese", "hana.hi"},
		Messages: []demoChatMessage{
			{"gilang.hi", "meet at anggrek lobby 6pm sharp"},
			{"dina.hi", "bringing my tripod finally"},
			{"chermiese", "can someone lend me a 50mm 🥺"},
			{"hana.hi", "I got you chermiese"},
		},
	},
}

func seedDevChats(db *gorm.DB, users map[string]usermodel.User) {
	for _, room := range demoChatRooms {
		members := make([]usermodel.User, 0, len(room.Members))
		for _, m := range room.Members {
			if u, ok := users[m]; ok {
				members = append(members, u)
			}
		}
		if len(members) < 2 {
			continue
		}

		// Idempotency: for groups match by name; for private match by member pair
		// already sharing a private room.
		if room.Type == chatroommodel.ChatRoomTypeGroup {
			var existing chatroommodel.ChatRoom
			if err := db.Where("name = ? AND type = ?", room.Name, room.Type).First(&existing).Error; err == nil {
				continue
			}
		} else if devPrivateRoomExists(db, members[0].UserID, members[1].UserID) {
			continue
		}

		cr := chatroommodel.ChatRoom{
			Type:        room.Type,
			MemberCount: len(members),
			CreatedAt:   time.Now().Add(-24 * time.Hour),
		}
		if room.Name != "" {
			name := room.Name
			cr.Name = &name
		}
		if err := db.Create(&cr).Error; err != nil {
			log.Printf("dev seeder: chat room failed: %v", err)
			continue
		}

		for _, m := range members {
			db.Create(&chatroommembermodel.ChatRoomMember{
				RoomID:   cr.RoomID,
				UserID:   m.UserID,
				JoinedAt: time.Now().Add(-24 * time.Hour),
			})
		}

		base := time.Now().Add(-3 * time.Hour)
		for i, msg := range room.Messages {
			sender, ok := users[msg.Sender]
			if !ok {
				continue
			}
			text := msg.Text
			chat := chatmodel.Chat{
				RoomID:    cr.RoomID,
				SenderID:  sender.UserID,
				Message:   &text,
				IsRead:    true,
				CreatedAt: base.Add(time.Duration(i) * 2 * time.Minute),
			}
			db.Create(&chat)
		}
		label := room.Name
		if label == "" {
			label = "private"
		}
		log.Printf("dev seeder: chat room %q created (%d members, %d messages)", label, len(members), len(room.Messages))
	}
}

// devPrivateRoomExists reports whether a private room already contains both users.
func devPrivateRoomExists(db *gorm.DB, a, b uint) bool {
	var count int64
	db.Model(&chatroommembermodel.ChatRoomMember{}).
		Joins("JOIN chat_rooms ON chat_rooms.room_id = chat_room_members.room_id").
		Where("chat_rooms.type = ?", chatroommodel.ChatRoomTypePrivate).
		Where("chat_room_members.room_id IN (?)",
			db.Model(&chatroommembermodel.ChatRoomMember{}).
				Select("room_id").Where("user_id = ?", a)).
		Where("chat_room_members.user_id = ?", b).
		Count(&count)
	return count > 0
}

// uuidSlug builds a short, deterministic, filename-safe slug from arbitrary text
// for placeholder object names.
func uuidSlug(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	var b strings.Builder
	for _, r := range s {
		switch {
		case r >= 'a' && r <= 'z', r >= '0' && r <= '9':
			b.WriteRune(r)
		case r == ' ' || r == '-' || r == '_':
			b.WriteRune('-')
		}
	}
	out := b.String()
	if len(out) > 40 {
		out = out[:40]
	}
	return strings.Trim(out, "-")
}
