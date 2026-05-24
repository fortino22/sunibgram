package userfacultyrepo

import (
	"context"
	"errors"
	"fmt"

	"gorm.io/gorm"

	usermodel "github.com/sunibgram/domain/model/user"
	userfacultyrepoiface "github.com/sunibgram/domain/repository/user_faculty"
	"github.com/sunibgram/infrastructure/cache"
	infrarepoShared "github.com/sunibgram/infrastructure/repository/_shared"
	notificationrepo "github.com/sunibgram/infrastructure/repository/notification"
	userrepo "github.com/sunibgram/infrastructure/repository/user"
)

type userFacultyRepository struct {
	writeDB *gorm.DB
	readDB  *gorm.DB
	cache   *cache.Cache
}

func NewUserFacultyRepository(writeDB, readDB *gorm.DB, c *cache.Cache) userfacultyrepoiface.IUserFacultyRepository {
	return &userFacultyRepository{writeDB: writeDB, readDB: readDB, cache: c}
}

func userFacultyKey(userID uint) string {
	return fmt.Sprintf("userfaculty:user:%d", userID)
}

func (r *userFacultyRepository) Save(uf *usermodel.UserFaculty) error {
	if err := r.writeDB.Create(uf).Error; err != nil {
		return err
	}
	r.invalidateUser(uf.UserID)
	r.invalidateProfileMeta(uf.UserID)
	notificationrepo.InvalidateNotificationSenderCache(context.Background(), r.cache, uf.UserID)
	r.invalidateThreadCaches()
	return nil
}

func (r *userFacultyRepository) FindByUserID(userID uint) (*usermodel.UserFaculty, error) {
	ctx := context.Background()
	key := userFacultyKey(userID)
	var uf usermodel.UserFaculty
	if err := r.cache.Get(ctx, key, &uf); err == nil {
		return &uf, nil
	}
	if err := r.readDB.Preload("Faculty").Where("user_id = ?", userID).First(&uf).Error; err != nil {
		return nil, err
	}
	r.cache.Set(ctx, key, uf, cache.FacultyTTL)
	return &uf, nil
}

func (r *userFacultyRepository) FindByUserIDs(ids []uint) ([]usermodel.UserFaculty, error) {
	if len(ids) == 0 {
		return []usermodel.UserFaculty{}, nil
	}
	ctx := context.Background()
	rows := make([]usermodel.UserFaculty, 0, len(ids))
	missing := make([]uint, 0, len(ids))
	if r.cache != nil {
		for _, id := range ids {
			var uf usermodel.UserFaculty
			if err := r.cache.Get(ctx, userFacultyKey(id), &uf); err == nil {
				rows = append(rows, uf)
				continue
			}
			missing = append(missing, id)
		}
	} else {
		missing = ids
	}
	if len(missing) == 0 {
		return rows, nil
	}
	var fetched []usermodel.UserFaculty
	if err := r.readDB.Preload("Faculty").Where("user_id IN ?", missing).Find(&fetched).Error; err != nil {
		return nil, err
	}
	if r.cache != nil {
		for i := range fetched {
			r.cache.Set(ctx, userFacultyKey(fetched[i].UserID), fetched[i], cache.FacultyTTL)
		}
	}
	rows = append(rows, fetched...)
	return rows, nil
}

func (r *userFacultyRepository) FindByFacultyID(facultyID uint, page, limit int) ([]usermodel.UserFaculty, int64, error) {
	var rows []usermodel.UserFaculty
	var total int64
	r.readDB.Model(&usermodel.UserFaculty{}).Where("faculty_id = ?", facultyID).Count(&total)
	offset := (page - 1) * limit
	if err := r.readDB.Preload("User").
		Where("faculty_id = ?", facultyID).
		Offset(offset).Limit(limit).
		Find(&rows).Error; err != nil {
		return nil, 0, err
	}
	return rows, total, nil
}

func (r *userFacultyRepository) Delete(userID, facultyID uint) error {
	result := r.writeDB.Delete(&usermodel.UserFaculty{}, "user_id = ? AND faculty_id = ?", userID, facultyID)
	if result.RowsAffected == 0 {
		return errors.New("assignment not found")
	}
	if result.Error != nil {
		return result.Error
	}
	r.invalidateUser(userID)
	r.invalidateProfileMeta(userID)
	notificationrepo.InvalidateNotificationSenderCache(context.Background(), r.cache, userID)
	r.invalidateThreadCaches()
	return nil
}

func (r *userFacultyRepository) invalidateUser(userID uint) {
	if r.cache == nil {
		return
	}
	r.cache.Delete(context.Background(), userFacultyKey(userID))
}

func (r *userFacultyRepository) invalidateProfileMeta(userID uint) {
	if r.cache == nil {
		return
	}
	r.cache.Delete(context.Background(), userrepo.ProfileMetaKey(userID))
}

func (r *userFacultyRepository) invalidateThreadCaches() {
	if r.cache == nil {
		return
	}
	ctx := context.Background()
	_ = r.cache.Delete(ctx, "feed:trending:global", "reports:all")
	_ = r.cache.InvalidateTags(ctx,
		infrarepoShared.ThreadRootsGlobalTag(),
		infrarepoShared.ThreadCommunityGlobalTag(),
		infrarepoShared.ThreadAllCommunitiesGlobalTag(),
		infrarepoShared.ThreadMyCommunitiesGlobalTag(),
		infrarepoShared.ThreadFollowingGlobalTag(),
		infrarepoShared.CommunityThreadsGlobalTag(),
		infrarepoShared.EventThreadsGlobalTag(),
	)
	// thread:userfeed/usermedia:*, feed:foryou:*, reports:id:* still SCAN —
	// rare faculty assignment path, no global tag for those families.
	_ = r.cache.DeleteByPatterns(ctx,
		"thread:userfeed:*",
		"thread:usermedia:*",
		"feed:foryou:*",
		"reports:id:*",
	)
}
