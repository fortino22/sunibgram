package seeder

import (
	"log"
	"os"
	"strings"
	"time"

	"gorm.io/gorm"

	userrelationshipmodel "github.com/sunibgram/domain/model/user_relationship"
	usermodel "github.com/sunibgram/domain/model/user"
	userprofilemodel "github.com/sunibgram/domain/model/user_profile"
)

// SeedDev populates a rich, browseable demo dataset (users, UKM communities,
// threads with engagement, events, chats) so a developer running locally sees a
// live-feeling app. It is gated on SEED_DEV=true and is a no-op otherwise — it
// must never run in production. Every section is idempotent (exists-check before
// create) so re-running migration does not duplicate rows.
func SeedDev(db *gorm.DB) {
	if strings.TrimSpace(strings.ToLower(os.Getenv("SEED_DEV"))) != "true" {
		return
	}
	log.Println("dev seeder: SEED_DEV=true → seeding demo data")

	// Reuse the working dummy-user creation (profiles + settings + faculty).
	SeedDummy(db)

	users := loadDemoUsers(db)
	if len(users) == 0 {
		log.Println("dev seeder: no demo users found, aborting")
		return
	}

	enrichDemoProfiles(db, users)
	seedDevFollows(db, users)
	communities := seedDevCommunities(db, users)
	threadIDs := seedDevThreads(db, users, communities)
	seedDevEngagement(db, users, threadIDs)
	seedDevEvents(db, users, communities)
	seedDevChats(db, users)

	log.Println("dev seeder: demo data seeding complete")
}

// loadDemoUsers returns the seeded dummy users keyed by username, so downstream
// seeders can reference people by a stable handle without hardcoding IDs.
func loadDemoUsers(db *gorm.DB) map[string]usermodel.User {
	out := make(map[string]usermodel.User)
	for _, acc := range dummyAccounts {
		var u usermodel.User
		if err := db.Where("email = ?", strings.TrimSpace(acc.Email)).First(&u).Error; err != nil {
			continue
		}
		out[u.UserName] = u
	}
	return out
}

// demoUsernames is the stable ordering of demo handles, used where we need a
// deterministic slice of users (round-robin authorship, membership, etc.).
var demoUsernames = []string{
	"Octa.hi", "devin.hi", "chermiese", "aldi.hi", "bella.hi", "caesar.hi",
	"dina.hi", "eko.hi", "fira.hi", "gilang.hi", "hana.hi", "irfan.hi", "jihan.hi",
}

// demoUserSlice returns the demo users in demoUsernames order, skipping any that
// failed to load.
func demoUserSlice(users map[string]usermodel.User) []usermodel.User {
	out := make([]usermodel.User, 0, len(demoUsernames))
	for _, name := range demoUsernames {
		if u, ok := users[name]; ok {
			out = append(out, u)
		}
	}
	return out
}

type demoProfileBio struct {
	Username string
	Bio      string
}

var demoProfileBios = []demoProfileBio{
	{"Octa.hi", "CS '23 · BNCC core team · ngoding sambil ngopi ☕"},
	{"devin.hi", "Sistem Informasi · suka bikin web random tengah malam"},
	{"chermiese", "Design student · doodling > deadlines 🎨"},
	{"aldi.hi", "Competitive programmer wannabe · halal anime enjoyer"},
	{"bella.hi", "SI '23 · org freak · ketua panitia 5 acara sekaligus"},
	{"caesar.hi", "Accounting but my heart is in basketball 🏀"},
	{"dina.hi", "Design · motion & illustration · commission open"},
	{"eko.hi", "Engineering · robot tinkerer · solder fumes addict"},
	{"fira.hi", "Humanities · debate club · will argue about anything"},
	{"gilang.hi", "Creative arts · photographer · golden hour chaser 📷"},
	{"hana.hi", "SI · UI/UX · figma is my second home"},
	{"irfan.hi", "CS · backend gremlin · Go enthusiast 🐹"},
	{"jihan.hi", "BBS · entrepreneurship · selling matcha on campus"},
}

// enrichDemoProfiles adds bios and placeholder avatar/banner object names so the
// demo profiles look populated when browsed.
func enrichDemoProfiles(db *gorm.DB, users map[string]usermodel.User) {
	for _, b := range demoProfileBios {
		u, ok := users[b.Username]
		if !ok {
			continue
		}
		var profile userprofilemodel.UserProfile
		if err := db.Where("user_id = ?", u.UserID).First(&profile).Error; err != nil {
			continue
		}
		if profile.Bio != "" {
			continue // already enriched, keep idempotent
		}
		slug := strings.ToLower(strings.ReplaceAll(b.Username, ".", "-"))
		db.Model(&profile).Updates(map[string]any{
			"bio":                b.Bio,
			"avatar_object_name": "avatar/demo-" + slug + ".jpg",
			"banner_object_name": "banner/demo-" + slug + ".jpg",
		})
	}
}

// seedDevFollows builds a small follow graph: a couple of mutual follows plus a
// few one-way follows so feeds and follower counts are non-trivial.
func seedDevFollows(db *gorm.DB, users map[string]usermodel.User) {
	type edge struct{ follower, following string }
	edges := []edge{
		{"devin.hi", "Octa.hi"}, {"Octa.hi", "devin.hi"}, // mutual
		{"bella.hi", "hana.hi"}, {"hana.hi", "bella.hi"}, // mutual
		{"chermiese", "dina.hi"}, {"dina.hi", "chermiese"}, // mutual
		{"aldi.hi", "Octa.hi"}, {"irfan.hi", "Octa.hi"}, {"eko.hi", "Octa.hi"},
		{"caesar.hi", "bella.hi"}, {"fira.hi", "bella.hi"},
		{"gilang.hi", "dina.hi"}, {"jihan.hi", "bella.hi"},
		{"devin.hi", "irfan.hi"}, {"hana.hi", "chermiese"},
		{"Octa.hi", "irfan.hi"}, {"bella.hi", "jihan.hi"},
	}
	for _, e := range edges {
		f, ok1 := users[e.follower]
		g, ok2 := users[e.following]
		if !ok1 || !ok2 || f.UserID == g.UserID {
			continue
		}
		var existing userrelationshipmodel.UserRelationship
		if err := db.Where("follower_id = ? AND following_id = ?", f.UserID, g.UserID).First(&existing).Error; err == nil {
			continue
		}
		rel := userrelationshipmodel.UserRelationship{
			FollowerID:  f.UserID,
			FollowingID: g.UserID,
			CreatedAt:   time.Now().Add(-time.Duration(48) * time.Hour),
		}
		if err := db.Create(&rel).Error; err != nil {
			log.Printf("dev seeder: follow %s→%s failed: %v", e.follower, e.following, err)
		}
	}
}
