package seeder

import (
	"log"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	categorymodel "github.com/sunibgram/domain/model/category"
	communitymodel "github.com/sunibgram/domain/model/community"
	communitymembermodel "github.com/sunibgram/domain/model/community_member"
	communityrulemodel "github.com/sunibgram/domain/model/community_rule"
	usermodel "github.com/sunibgram/domain/model/user"
)

type demoCommunity struct {
	Name        string
	Slug        string
	Description string
	Category    string   // must match a seeded category name
	Creator     string   // demo username
	Members     []string // demo usernames (excluding creator)
	Rules       []string
}

// realBinusUKM lists actual Binus student clubs (UKM) so the discover/browse
// surface feels authentic.
var realBinusUKM = []demoCommunity{
	{
		Name: "BNCC", Slug: "bncc",
		Description: "Bina Nusantara Computer Club — the largest IT student organization at Binus. Bootcamps, competitive programming, and tech talks.",
		Category:    "Technology", Creator: "Octa.hi",
		Members: []string{"devin.hi", "irfan.hi", "aldi.hi", "hana.hi"},
		Rules:   []string{"Be respectful to fellow members.", "No spam or self-promotion.", "Share knowledge, help juniors."},
	},
	{
		Name: "Teach For Indonesia", Slug: "tfi",
		Description: "Binus social movement empowering communities through education. Volunteer, teach, and make an impact.",
		Category:    "Volunteering", Creator: "fira.hi",
		Members: []string{"bella.hi", "jihan.hi", "dina.hi"},
		Rules:   []string{"Commitment matters — show up for the kids.", "Kindness first."},
	},
	{
		Name: "Binus Basketball Club", Slug: "basketball",
		Description: "Hoops, drills, and inter-faculty tournaments. All skill levels welcome.",
		Category:    "Sports", Creator: "caesar.hi",
		Members: []string{"eko.hi", "gilang.hi", "aldi.hi"},
		Rules:   []string{"Respect practice schedules.", "Bring your own water bottle."},
	},
	{
		Name: "HIMTI", Slug: "himti",
		Description: "Himpunan Mahasiswa Teknik Informatika — the Informatics student association. Academic events, study groups, and networking.",
		Category:    "Academic", Creator: "irfan.hi",
		Members: []string{"Octa.hi", "devin.hi", "hana.hi"},
		Rules:   []string{"Academic integrity always.", "Help organize, don't just attend."},
	},
	{
		Name: "Binus English Club", Slug: "english-club",
		Description: "Practice English through debates, public speaking, and casual conversation nights.",
		Category:    "Student Life", Creator: "fira.hi",
		Members: []string{"bella.hi", "jihan.hi", "chermiese"},
		Rules:   []string{"English only during sessions.", "No judgment — everyone is learning."},
	},
	{
		Name: "Binus Gaming & Esports", Slug: "esports",
		Description: "Competitive and casual gaming. Valorant, MLBB, and tournament squads forming every semester.",
		Category:    "Gaming", Creator: "aldi.hi",
		Members: []string{"eko.hi", "gilang.hi", "devin.hi", "irfan.hi"},
		Rules:   []string{"No toxicity in voice chat.", "Schedule scrims in advance."},
	},
	{
		Name: "Binus Entrepreneurship Club", Slug: "entrepreneurship",
		Description: "Build, pitch, and grow. Workshops with founders and a campus startup community.",
		Category:    "Career", Creator: "jihan.hi",
		Members: []string{"bella.hi", "caesar.hi"},
		Rules:   []string{"Idea-friendly zone — no mocking pitches.", "Give before you take."},
	},
	{
		Name: "Klik! Photography Club", Slug: "photography",
		Description: "Hunting golden hour, lighting workshops, and monthly photo challenges.",
		Category:    "Arts", Creator: "gilang.hi",
		Members: []string{"dina.hi", "chermiese", "hana.hi"},
		Rules:   []string{"Credit photographers.", "Constructive critique only."},
	},
	{
		Name: "Binus Dance Club", Slug: "dance",
		Description: "From K-pop covers to contemporary. Weekly practice and stage performances.",
		Category:    "Arts", Creator: "dina.hi",
		Members: []string{"chermiese", "bella.hi", "fira.hi"},
		Rules:   []string{"Respect everyone's pace.", "Practice attendance counts."},
	},
	{
		Name: "Nippon Club", Slug: "nippon",
		Description: "Japanese language and culture — anime nights, JLPT study, and bunkasai events.",
		Category:    "Social", Creator: "aldi.hi",
		Members: []string{"irfan.hi", "hana.hi", "eko.hi"},
		Rules:   []string{"日本語 practice encouraged.", "Be welcoming to beginners."},
	},
	{
		Name: "Binus Robotics", Slug: "robotics",
		Description: "Build robots, join competitions, and break stuff (productively). Arduino to autonomous.",
		Category:    "Technology", Creator: "eko.hi",
		Members: []string{"Octa.hi", "irfan.hi", "devin.hi"},
		Rules:   []string{"Lab safety first.", "Document your builds."},
	},
	{
		Name: "Binus Music Community", Slug: "music",
		Description: "Jam sessions, acoustic nights, and a band-matching board for aspiring musicians.",
		Category:    "Arts", Creator: "gilang.hi",
		Members: []string{"caesar.hi", "dina.hi", "chermiese"},
		Rules:   []string{"Share the practice room fairly.", "Encourage, don't gatekeep."},
	},
}

// seedDevCommunities creates the UKM communities with members and rules.
// Returns community name → CommunityID for downstream seeders.
func seedDevCommunities(db *gorm.DB, users map[string]usermodel.User) map[string]uint {
	out := make(map[string]uint)
	catCache := make(map[string]uuid.UUID)

	for _, c := range realBinusUKM {
		creator, ok := users[c.Creator]
		if !ok {
			log.Printf("dev seeder: community %s creator %s missing, skipping", c.Name, c.Creator)
			continue
		}

		// Idempotency: skip if a community with this name already exists.
		var existing communitymodel.Community
		if err := db.Where("name = ?", c.Name).First(&existing).Error; err == nil {
			out[c.Name] = existing.CommunityID
			continue
		}

		catID := resolveDevCategory(db, catCache, c.Category)

		// Members = creator (ADMIN) + listed members (MEMBER).
		memberUsers := make([]usermodel.User, 0, len(c.Members)+1)
		memberUsers = append(memberUsers, creator)
		for _, m := range c.Members {
			if mu, ok := users[m]; ok && mu.UserID != creator.UserID {
				memberUsers = append(memberUsers, mu)
			}
		}

		community := communitymodel.Community{
			UserID:           creator.UserID,
			Name:             c.Name,
			Description:      c.Description,
			Visibility:       communitymodel.CommunityVisibilityPublic,
			BannerObjectName: "community-banner/demo-" + c.Slug + ".jpg",
			MemberCount:      len(memberUsers),
			IsActive:         true,
		}
		if catID != uuid.Nil {
			community.CategoryID = &catID
		}
		if err := db.Create(&community).Error; err != nil {
			log.Printf("dev seeder: create community %s failed: %v", c.Name, err)
			continue
		}
		out[c.Name] = community.CommunityID

		for i, mu := range memberUsers {
			role := communitymembermodel.CommunityMemberRoleMember
			if mu.UserID == creator.UserID {
				role = communitymembermodel.CommunityMemberRoleAdmin
			}
			member := communitymembermodel.CommunityMember{
				CommunityID: community.CommunityID,
				UserID:      mu.UserID,
				Role:        role,
				JoinedAt:    time.Now().Add(-time.Duration(72-i*3) * time.Hour),
			}
			if err := db.Create(&member).Error; err != nil {
				log.Printf("dev seeder: add member %d to %s failed: %v", mu.UserID, c.Name, err)
			}
		}

		for i, rule := range c.Rules {
			r := communityrulemodel.CommunityRule{
				CommunityID: community.CommunityID,
				OrderNum:    i + 1,
				Text:        rule,
			}
			if err := db.Create(&r).Error; err != nil {
				log.Printf("dev seeder: add rule to %s failed: %v", c.Name, err)
			}
		}

		log.Printf("dev seeder: community %q created (%d members)", c.Name, len(memberUsers))
	}

	return out
}

func resolveDevCategory(db *gorm.DB, cache map[string]uuid.UUID, name string) uuid.UUID {
	if id, ok := cache[name]; ok {
		return id
	}
	var cat categorymodel.Category
	if err := db.Where("name = ?", name).First(&cat).Error; err != nil {
		log.Printf("dev seeder: category %q not found", name)
		cache[name] = uuid.Nil
		return uuid.Nil
	}
	cache[name] = cat.ID
	return cat.ID
}
