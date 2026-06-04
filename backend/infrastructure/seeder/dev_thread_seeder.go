package seeder

import (
	"log"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	bookmarkmodel "github.com/sunibgram/domain/model/bookmark"
	hashtagmodel "github.com/sunibgram/domain/model/hashtag"
	likemodel "github.com/sunibgram/domain/model/like"
	repostmodel "github.com/sunibgram/domain/model/repost"
	threadmodel "github.com/sunibgram/domain/model/thread"
	threadmediamodel "github.com/sunibgram/domain/model/thread_media"
	threadpollmodel "github.com/sunibgram/domain/model/thread_poll"
	usermodel "github.com/sunibgram/domain/model/user"
)

type demoPost struct {
	Author      string   // demo username
	Community   string   // community name, or "" for global feed
	Content     string
	Hashtags    []string
	MediaCount  int      // number of placeholder media items
	Replies     []demoReply
}

type demoReply struct {
	Author  string
	Content string
}

var demoPosts = []demoPost{
	{
		Author: "Octa.hi", Community: "BNCC",
		Content: "Reminder: BNCC web dev bootcamp starts this Saturday 9AM at Anggrek. Bring your laptop and a lot of caffeine ☕ #BNCC #webdev",
		Hashtags: []string{"BNCC", "webdev"}, MediaCount: 1,
		Replies: []demoReply{
			{"devin.hi", "Finally!! been waiting for this 🔥"},
			{"irfan.hi", "Will there be a backend track too?"},
			{"Octa.hi", "@irfan backend track next month, stay tuned"},
		},
	},
	{
		Author: "devin.hi", Community: "",
		Content: "spent 3 hours debugging only to find out it was a missing semicolon. peak software engineering moment 😭",
		Replies: []demoReply{
			{"irfan.hi", "the semicolon always wins"},
			{"aldi.hi", "skill issue (affectionate)"},
		},
	},
	{
		Author: "caesar.hi", Community: "Binus Basketball Club",
		Content: "Inter-faculty tournament bracket is OUT. We're up against FoE first round. Practice tonight 7PM, no excuses 🏀 #BinusBasketball",
		Hashtags: []string{"BinusBasketball"},
		Replies: []demoReply{
			{"eko.hi", "FoE going down 💪"},
			{"gilang.hi", "can someone bring the speaker?"},
		},
	},
	{
		Author: "dina.hi", Community: "Klik! Photography Club",
		Content: "Golden hour at Kemanggisan hits different. This month's theme: 'campus after dark'. Drop your shots below 📷",
		Hashtags: []string{"photography"}, MediaCount: 2,
		Replies: []demoReply{
			{"gilang.hi", "the lighting in the second one is insane"},
			{"chermiese", "okay now I need a new lens 🥲"},
		},
	},
	{
		Author: "bella.hi", Community: "",
		Content: "PSA: never join 5 committees in one semester. it is not a personality trait, it is a cry for help. ask me how I know.",
		Replies: []demoReply{
			{"jihan.hi", "calling me out at 2am huh"},
			{"hana.hi", "this is so real it hurts"},
		},
	},
	{
		Author: "fira.hi", Community: "Binus English Club",
		Content: "Debate night this Friday! Motion: 'This house would ban group projects.' Pro side already winning in my heart 😌 #debate",
		Hashtags: []string{"debate"},
		Replies: []demoReply{
			{"bella.hi", "as a serial group-project carrier, I am PRO"},
		},
	},
	{
		Author: "eko.hi", Community: "Binus Robotics",
		Content: "line-follower bot finally took the corner without flying off the table. small wins. competition in 3 weeks 🤖 #robotics",
		Hashtags: []string{"robotics"}, MediaCount: 1,
		Replies: []demoReply{
			{"Octa.hi", "the table edge: final boss"},
			{"irfan.hi", "what's your PID tuning lol"},
		},
	},
	{
		Author: "hana.hi", Community: "",
		Content: "redesigned the campus event app mockup in figma. accessibility-first this time. feedback welcome 🙏 #uiux #figma",
		Hashtags: []string{"uiux", "figma"}, MediaCount: 1,
		Replies: []demoReply{
			{"chermiese", "the spacing is so clean omg"},
			{"devin.hi", "please ship this, the real one is painful"},
		},
	},
	{
		Author: "jihan.hi", Community: "Binus Entrepreneurship Club",
		Content: "selling matcha lattes near Syahdan tomorrow ☕💚 100% student-run, profits fund our pitch trip. support local founders!",
		Replies: []demoReply{
			{"bella.hi", "reserve one for me!!"},
			{"caesar.hi", "matcha supremacy"},
		},
	},
	{
		Author: "aldi.hi", Community: "Binus Gaming & Esports",
		Content: "scrims tonight 8PM. need one more for the Valorant roster. immortal+ only (jk, vibes > rank) #esports",
		Hashtags: []string{"esports"},
		Replies: []demoReply{
			{"eko.hi", "I'm iron but I have heart"},
			{"gilang.hi", "put me in coach"},
		},
	},
	{
		Author: "chermiese", Community: "Binus Dance Club",
		Content: "new K-pop cover choreo dropping next week 💃 practice was brutal but worth it. stage performance at campus festival soon!",
		Hashtags: []string{}, MediaCount: 1,
		Replies: []demoReply{
			{"dina.hi", "front row reserved for me"},
		},
	},
	{
		Author: "irfan.hi", Community: "HIMTI",
		Content: "study group for algorithms midterm this weekend. bringing snacks and existential dread. all welcome #HIMTI #algorithms",
		Hashtags: []string{"HIMTI", "algorithms"},
		Replies: []demoReply{
			{"Octa.hi", "the snacks are the only reason I'm coming"},
			{"devin.hi", "is dynamic programming on it 😨"},
		},
	},
	{
		Author: "gilang.hi", Community: "",
		Content: "took 400 photos at the music club jam session. 6 are usable. that's photography baby 📸 #music #photography",
		Hashtags: []string{"music", "photography"}, MediaCount: 2,
	},
	{
		Author: "Octa.hi", Community: "",
		Content: "hot take: tabs vs spaces is a solved problem and the answer is 'whatever your linter says, stop arguing about it'",
		Replies: []demoReply{
			{"irfan.hi", "spaces. this is not up for debate."},
			{"devin.hi", "tabs. fight me."},
			{"aldi.hi", "I use both to watch the world burn"},
		},
	},
	{
		Author: "fira.hi", Community: "",
		Content: "humanities student here to remind the CS majors that 'just google it' is also a research methodology, you didn't invent it 😌",
		Replies: []demoReply{
			{"Octa.hi", "ok this is fair"},
		},
	},
	{
		Author: "dina.hi", Community: "Binus Music Community",
		Content: "acoustic night Thursday at the student lounge. open mic slots still open. come sing badly with us, no judgment 🎸",
		Replies: []demoReply{
			{"gilang.hi", "I'll bring the cajon"},
			{"caesar.hi", "badly is my specialty"},
		},
	},
	{
		Author: "hana.hi", Community: "Nippon Club",
		Content: "JLPT N3 study session this weekend + we're planning a bunkasai (cultural festival) for next month! 文化祭 hype 🎌",
		Replies: []demoReply{
			{"irfan.hi", "finally putting my anime hours to use"},
		},
	},
	{
		Author: "bella.hi", Community: "Teach For Indonesia",
		Content: "weekend teaching trip to the partner school was so worth it. the kids made us cards 🥺 sign up for next month's batch!",
		Hashtags: []string{}, MediaCount: 1,
		Replies: []demoReply{
			{"fira.hi", "this made my whole week"},
			{"jihan.hi", "count me in next round"},
		},
	},
}

// demoPoll is attached to one standalone post by author.
type demoPollSpec struct {
	Author   string
	Question string
	Options  []string
	Votes    []int // vote count per option
}

var demoPolls = []demoPollSpec{
	{
		Author:   "Octa.hi",
		Question: "Best place to grind code on campus?",
		Options:  []string{"Library 5th floor", "Anggrek cafeteria", "Empty classroom", "Home (cope)"},
		Votes:    []int{12, 7, 4, 9},
	},
	{
		Author:   "bella.hi",
		Question: "How many committees is too many?",
		Options:  []string{"One is plenty", "Two, manageable", "Three, send help"},
		Votes:    []int{5, 11, 8},
	},
}

// seedDevThreads creates posts, replies, reposts, media, polls and hashtags.
// Returns the IDs of all root POST threads for the engagement seeder.
func seedDevThreads(db *gorm.DB, users map[string]usermodel.User, communities map[string]uint) []uint {
	// Idempotency: if demo threads already exist (matched by a sentinel content),
	// skip the whole section.
	var sentinel int64
	db.Model(&threadmodel.Thread{}).Where("content LIKE ?", "%BNCC web dev bootcamp starts this Saturday%").Count(&sentinel)
	if sentinel > 0 {
		log.Println("dev seeder: demo threads already present, skipping thread seeding")
		return existingDemoPostIDs(db)
	}

	hashtagCache := make(map[string]hashtagmodel.Hashtag)
	postIDs := make([]uint, 0, len(demoPosts))
	now := time.Now()

	for i, p := range demoPosts {
		author, ok := users[p.Author]
		if !ok {
			continue
		}
		var commID *uint
		if p.Community != "" {
			if id, ok := communities[p.Community]; ok {
				commID = &id
			}
		}
		createdAt := now.Add(-time.Duration(len(demoPosts)-i) * 90 * time.Minute)

		post := threadmodel.Thread{
			UserID:          author.UserID,
			CommunityID:     commID,
			Type:            threadmodel.ThreadTypePost,
			Content:         p.Content,
			ReplyPermission: threadmodel.ReplyPermissionEveryone,
			CommentCount:    len(p.Replies),
			CreatedAt:       createdAt,
		}
		if err := db.Create(&post).Error; err != nil {
			log.Printf("dev seeder: create post failed: %v", err)
			continue
		}
		post.RootID = &post.ThreadID
		db.Model(&post).Update("root_id", post.ThreadID)
		postIDs = append(postIDs, post.ThreadID)

		for m := 0; m < p.MediaCount; m++ {
			media := threadmediamodel.ThreadMedia{
				ThreadID:   post.ThreadID,
				ObjectName: "post/demo-" + uuid.NewString() + ".jpg",
				CreatedAt:  createdAt,
			}
			db.Create(&media)
		}

		for _, tag := range p.Hashtags {
			ht := resolveDevHashtag(db, hashtagCache, tag)
			link := hashtagmodel.ThreadHashtag{ThreadID: post.ThreadID, HashtagID: ht.ID, CreatedAt: createdAt}
			if err := db.Create(&link).Error; err == nil {
				db.Model(&hashtagmodel.Hashtag{}).Where("id = ?", ht.ID).
					Update("thread_count", gorm.Expr("thread_count + 1"))
			}
		}

		// Replies (REPLY threads chained to the post).
		for r, reply := range p.Replies {
			ra, ok := users[reply.Author]
			if !ok {
				continue
			}
			parentID := post.ThreadID
			replyThread := threadmodel.Thread{
				UserID:          ra.UserID,
				Type:            threadmodel.ThreadTypeReply,
				Content:         reply.Content,
				ReplyPermission: threadmodel.ReplyPermissionEveryone,
				ParentID:        &parentID,
				RootID:          &post.ThreadID,
				CreatedAt:       createdAt.Add(time.Duration(r+1) * 15 * time.Minute),
			}
			db.Create(&replyThread)
		}
	}

	seedDevPolls(db, users)
	seedDevReposts(db, users, postIDs)

	log.Printf("dev seeder: created %d demo posts", len(postIDs))
	return postIDs
}

func existingDemoPostIDs(db *gorm.DB) []uint {
	var ids []uint
	db.Model(&threadmodel.Thread{}).Where("type = ?", threadmodel.ThreadTypePost).
		Order("thread_id").Limit(200).Pluck("thread_id", &ids)
	return ids
}

func resolveDevHashtag(db *gorm.DB, cache map[string]hashtagmodel.Hashtag, name string) hashtagmodel.Hashtag {
	if h, ok := cache[name]; ok {
		return h
	}
	var ht hashtagmodel.Hashtag
	if err := db.Where("name = ?", name).FirstOrCreate(&ht, hashtagmodel.Hashtag{Name: name}).Error; err != nil {
		log.Printf("dev seeder: hashtag %q failed: %v", name, err)
	}
	cache[name] = ht
	return ht
}

func seedDevPolls(db *gorm.DB, users map[string]usermodel.User) {
	for _, spec := range demoPolls {
		author, ok := users[spec.Author]
		if !ok {
			continue
		}
		post := threadmodel.Thread{
			UserID:          author.UserID,
			Type:            threadmodel.ThreadTypePost,
			Content:         spec.Question,
			ReplyPermission: threadmodel.ReplyPermissionEveryone,
			CreatedAt:       time.Now().Add(-6 * time.Hour),
		}
		if err := db.Create(&post).Error; err != nil {
			continue
		}
		db.Model(&post).Update("root_id", post.ThreadID)

		total := 0
		for _, v := range spec.Votes {
			total += v
		}
		poll := threadpollmodel.ThreadPoll{
			ThreadID:   post.ThreadID,
			Question:   spec.Question,
			EndsAt:     time.Now().Add(48 * time.Hour),
			TotalVotes: total,
		}
		if err := db.Create(&poll).Error; err != nil {
			continue
		}
		for i, label := range spec.Options {
			vc := 0
			if i < len(spec.Votes) {
				vc = spec.Votes[i]
			}
			opt := threadpollmodel.ThreadPollOption{
				PollID:    poll.PollID,
				Label:     label,
				Position:  int16(i),
				VoteCount: vc,
			}
			db.Create(&opt)
		}
	}
}

func seedDevReposts(db *gorm.DB, users map[string]usermodel.User, postIDs []uint) {
	if len(postIDs) < 4 {
		return
	}
	type repostSpec struct {
		Author    string
		PostIndex int
	}
	specs := []repostSpec{
		{"hana.hi", 0}, {"caesar.hi", 3}, {"devin.hi", 7},
	}
	for _, s := range specs {
		author, ok := users[s.Author]
		if !ok || s.PostIndex >= len(postIDs) {
			continue
		}
		orig := postIDs[s.PostIndex]
		repost := threadmodel.Thread{
			UserID:           author.UserID,
			Type:             threadmodel.ThreadTypeRepost,
			Content:          "",
			ReplyPermission:  threadmodel.ReplyPermissionEveryone,
			OriginalThreadID: &orig,
			CreatedAt:        time.Now().Add(-3 * time.Hour),
		}
		if err := db.Create(&repost).Error; err == nil {
			db.Model(&threadmodel.Thread{}).Where("thread_id = ?", orig).
				Update("repost_count", gorm.Expr("repost_count + 1"))
			// Track the repost row so un-repost logic stays consistent.
			db.Create(&repostmodel.Repost{UserID: author.UserID, ThreadID: orig, CreatedAt: repost.CreatedAt})
		}
	}
}

// seedDevEngagement spreads likes and bookmarks across demo posts and syncs the
// denormalized counters to match the seeded rows.
func seedDevEngagement(db *gorm.DB, users map[string]usermodel.User, postIDs []uint) {
	if len(postIDs) == 0 {
		return
	}
	all := demoUserSlice(users)
	if len(all) == 0 {
		return
	}

	for idx, tid := range postIDs {
		// Deterministic, varied like spread: each post liked by a rotating subset.
		likeCount := 0
		for j, u := range all {
			if (j+idx)%3 != 0 { // ~1/3 of users like each post
				continue
			}
			like := likemodel.Like{UserID: u.UserID, ThreadID: tid, CreatedAt: time.Now().Add(-time.Duration(idx) * time.Hour)}
			var existing likemodel.Like
			if err := db.Where("user_id = ? AND thread_id = ?", u.UserID, tid).First(&existing).Error; err != nil {
				if err := db.Create(&like).Error; err == nil {
					likeCount++
				}
			}
		}

		// Bookmarks: every 4th post bookmarked by a couple users.
		bookmarkCount := 0
		if idx%4 == 0 {
			for j := 0; j < 2 && j < len(all); j++ {
				u := all[(idx+j)%len(all)]
				bm := bookmarkmodel.Bookmark{UserID: u.UserID, ThreadID: tid, CreatedAt: time.Now()}
				var existing bookmarkmodel.Bookmark
				if err := db.Where("user_id = ? AND thread_id = ?", u.UserID, tid).First(&existing).Error; err != nil {
					if err := db.Create(&bm).Error; err == nil {
						bookmarkCount++
					}
				}
			}
		}

		updates := map[string]any{}
		if likeCount > 0 {
			updates["like_count"] = gorm.Expr("like_count + ?", likeCount)
		}
		if bookmarkCount > 0 {
			updates["bookmark_count"] = gorm.Expr("bookmark_count + ?", bookmarkCount)
		}
		if len(updates) > 0 {
			db.Model(&threadmodel.Thread{}).Where("thread_id = ?", tid).Updates(updates)
		}
	}
	log.Printf("dev seeder: seeded engagement across %d posts", len(postIDs))
}
