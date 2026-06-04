package seeder

import (
	"log"
	"time"

	"gorm.io/gorm"

	communityeventmodel "github.com/sunibgram/domain/model/community_event"
	eventparticipantmodel "github.com/sunibgram/domain/model/event_participant"
	threadmodel "github.com/sunibgram/domain/model/thread"
	usermodel "github.com/sunibgram/domain/model/user"
)

type demoCommunityEvent struct {
	Community   string
	Creator     string
	Title       string
	Description string
	Location    string
	DaysAhead   int
	Invitees    []string
	Comments    []demoEventComment
}

type demoEventComment struct {
	Author  string
	Content string
}

var demoCommunityEvents = []demoCommunityEvent{
	{
		Community: "BNCC", Creator: "Octa.hi",
		Title:       "BNCC Web Dev Bootcamp — Day 1",
		Description: "Hands-on intro to modern web development. Bring your laptop. Snacks provided 🍕",
		Location:    "Anggrek Campus, Lab 7", DaysAhead: 4,
		Invitees: []string{"devin.hi", "irfan.hi", "aldi.hi", "hana.hi"},
		Comments: []demoEventComment{
			{"devin.hi", "Do we need to install anything beforehand?"},
			{"Octa.hi", "Just have Node and VS Code ready, we'll cover the rest!"},
		},
	},
	{
		Community: "Binus Basketball Club", Creator: "caesar.hi",
		Title:       "Inter-Faculty Tournament — Round 1",
		Description: "First round vs Faculty of Engineering. Come support the team! 🏀",
		Location:    "Binus Sports Hall", DaysAhead: 6,
		Invitees: []string{"eko.hi", "gilang.hi", "aldi.hi"},
		Comments: []demoEventComment{
			{"eko.hi", "Bringing the whole squad to cheer 📣"},
		},
	},
	{
		Community: "Klik! Photography Club", Creator: "gilang.hi",
		Title:       "Campus After Dark — Photo Walk",
		Description: "Evening photo walk around Kemanggisan. Tripods recommended. Theme: long exposure 📷",
		Location:    "Meet at Anggrek lobby, 6PM", DaysAhead: 9,
		Invitees: []string{"dina.hi", "chermiese", "hana.hi"},
		Comments: []demoEventComment{
			{"dina.hi", "finally an excuse to use my tripod"},
		},
	},
}

func seedDevEvents(db *gorm.DB, users map[string]usermodel.User, communities map[string]uint) {
	// Community events
	for _, e := range demoCommunityEvents {
		commID, ok := communities[e.Community]
		if !ok {
			continue
		}
		creator, ok := users[e.Creator]
		if !ok {
			continue
		}

		// Idempotency: skip if same title already exists for this community.
		var existing communityeventmodel.CommunityEvent
		if err := db.Where("community_id = ? AND title = ?", commID, e.Title).First(&existing).Error; err == nil {
			continue
		}

		start := time.Now().Add(time.Duration(e.DaysAhead) * 24 * time.Hour)
		ce := communityeventmodel.CommunityEvent{
			CommunityID:      commID,
			CreatedByUserID:  creator.UserID,
			Title:            e.Title,
			Description:      e.Description,
			Location:         e.Location,
			StartTime:        start,
			EndTime:          start.Add(3 * time.Hour),
			BannerObjectName: "community-event-banner/demo-" + uuidSlug(e.Title) + ".jpg",
			CommentCount:     len(e.Comments),
		}
		if err := db.Create(&ce).Error; err != nil {
			log.Printf("dev seeder: community event %q failed: %v", e.Title, err)
			continue
		}

		for _, inv := range e.Invitees {
			iu, ok := users[inv]
			if !ok {
				continue
			}
			invitee := communityeventmodel.CommunityEventInvitee{
				CommunityEventID: ce.ID,
				UserID:           iu.UserID,
				CreatedAt:        time.Now(),
			}
			db.Create(&invitee)
		}

		for _, c := range e.Comments {
			cu, ok := users[c.Author]
			if !ok {
				continue
			}
			comment := communityeventmodel.CommunityEventComment{
				CommunityEventID: ce.ID,
				UserID:           cu.UserID,
				Content:          c.Content,
			}
			db.Create(&comment)
		}
		log.Printf("dev seeder: community event %q created", e.Title)
	}

	// One global (campus-wide) event, approved.
	seedDevGlobalEvent(db, users)
}

func seedDevGlobalEvent(db *gorm.DB, users map[string]usermodel.User) {
	creator, ok := users["bella.hi"]
	if !ok {
		return
	}
	title := "Binus Tech & Arts Festival 2026"
	var existing threadmodel.Event
	if err := db.Where("title = ?", title).First(&existing).Error; err == nil {
		return
	}
	start := time.Now().Add(14 * 24 * time.Hour)
	ev := threadmodel.Event{
		UserID:           creator.UserID,
		Title:            title,
		Description:      "A campus-wide celebration of student tech projects, art exhibitions, music, and food stalls. Open to all Binusians!",
		Location:         "Anggrek Campus Atrium",
		Type:             "official",
		Status:           "approved",
		StartTime:        start,
		EndTime:          start.Add(8 * time.Hour),
		BannerObjectName: "event-banner/demo-techarts-festival.jpg",
		ParticipantCount: 0,
	}
	if err := db.Create(&ev).Error; err != nil {
		log.Printf("dev seeder: global event failed: %v", err)
		return
	}

	participants := []string{"Octa.hi", "devin.hi", "hana.hi", "gilang.hi", "dina.hi", "irfan.hi"}
	count := 0
	for _, p := range participants {
		pu, ok := users[p]
		if !ok {
			continue
		}
		ep := eventparticipantmodel.EventParticipant{
			EventID:  ev.ID,
			UserID:   pu.UserID,
			JoinedAt: time.Now(),
		}
		if err := db.Create(&ep).Error; err == nil {
			count++
		}
	}
	if count > 0 {
		db.Model(&threadmodel.Event{}).Where("id = ?", ev.ID).Update("participant_count", count)
	}
	log.Printf("dev seeder: global event %q created (%d participants)", title, count)
}
