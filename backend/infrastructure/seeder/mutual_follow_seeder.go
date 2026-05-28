package seeder

import (
	"log"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	usermodel "github.com/sunibgram/domain/model/user"
	userrelationshipmodel "github.com/sunibgram/domain/model/user_relationship"
	userstatmodel "github.com/sunibgram/domain/model/user_stat"
)

// SeedMutualFollows wires every non-admin user as mutual followers with the
// next N users in user_id order (wrap-around). Lets the Leaderboard "Your
// Stats" expand modal find mutual friends to compare against in dev without
// manually following accounts. Idempotent: skips if any user_relationships
// rows already exist.
func SeedMutualFollows(db *gorm.DB) {
	const mutualsPerUser = 5

	var existing int64
	if err := db.Model(&userrelationshipmodel.UserRelationship{}).Count(&existing).Error; err != nil {
		log.Printf("mutual follow seeder: count failed: %v", err)
		return
	}
	if existing > 0 {
		return
	}

	var users []usermodel.User
	if err := db.
		Joins("JOIN roles ON roles.role_id = users.role_id").
		Where("roles.name <> ?", "admin").
		Select("users.user_id").
		Order("users.user_id ASC").
		Find(&users).Error; err != nil {
		log.Printf("mutual follow seeder: list users failed: %v", err)
		return
	}
	n := len(users)
	if n < 2 {
		return
	}

	now := time.Now().UTC()
	rels := make([]userrelationshipmodel.UserRelationship, 0, n*mutualsPerUser*2)
	for i := 0; i < n; i++ {
		for k := 1; k <= mutualsPerUser; k++ {
			j := (i + k) % n
			if i == j {
				continue
			}
			a := users[i].UserID
			b := users[j].UserID
			rels = append(rels,
				userrelationshipmodel.UserRelationship{FollowerID: a, FollowingID: b, CreatedAt: now},
				userrelationshipmodel.UserRelationship{FollowerID: b, FollowingID: a, CreatedAt: now},
			)
		}
	}

	if err := db.Clauses(clause.OnConflict{DoNothing: true}).
		CreateInBatches(rels, 500).Error; err != nil {
		log.Printf("mutual follow seeder: insert failed: %v", err)
		return
	}

	// Recompute user_stats follower/following counts from the new rows.
	if err := db.Exec(`
		INSERT INTO user_stats (user_id, follower_count, following_count)
		SELECT u.user_id,
		       (SELECT COUNT(*) FROM user_relationships WHERE following_id = u.user_id),
		       (SELECT COUNT(*) FROM user_relationships WHERE follower_id  = u.user_id)
		FROM users u
		ON CONFLICT (user_id) DO UPDATE SET
		    follower_count  = EXCLUDED.follower_count,
		    following_count = EXCLUDED.following_count
	`).Error; err != nil {
		log.Printf("mutual follow seeder: stats recompute failed: %v", err)
	}

	// touch ref so import isn't unused if compiler check tightens
	_ = userstatmodel.UserStat{}

	log.Printf("mutual follow seeder: inserted %d relationships across %d users", len(rels), n)
}
