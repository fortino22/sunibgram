package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/sunibgram/domain/model"
	domainrepo "github.com/sunibgram/domain/repository"
	"github.com/sunibgram/infrastructure/cache"

	"gorm.io/gorm"
)

type UserRepository struct {
	writeDB *gorm.DB
	readDB  *gorm.DB
	cache   *cache.Cache
	ttl     time.Duration
}

func userKey(userID uint) string {
	return fmt.Sprintf("user:id:%d", userID)
}

func NewUserRepository(writeDB *gorm.DB, readDB *gorm.DB, c ...*cache.Cache) domainrepo.IUserRepository {
	var redisCache *cache.Cache
	if len(c) > 0 {
		redisCache = c[0]
	}
	return &UserRepository{writeDB: writeDB, readDB: readDB, cache: redisCache, ttl: cache.DefaultTTL}
}

func (r *UserRepository) FindByID(id uint) (*model.User, error) {
	ctx := context.Background()
	var user model.User
	key := userKey(id)
	if r.cache != nil {
		if err := r.cache.Get(ctx, key, &user); err == nil {
			return &user, nil
		}
	}
	if err := r.readDB.Preload("Role").First(&user, id).Error; err != nil {
		return nil, err
	}
	if r.cache != nil {
		r.cache.Set(ctx, key, user, r.ttl)
	}
	return &user, nil
}

func (r *UserRepository) Save(user *model.User) error {
	if err := r.writeDB.Create(user).Error; err != nil {
		return err
	}
	invalidateNotificationSenderCache(context.Background(), r.cache, user.UserID)
	invalidateExploreCache(context.Background(), r.cache)
	return nil
}

func (r *UserRepository) SaveWithProfile(user *model.User, profile *model.UserProfile) error {
	if err := r.writeDB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(user).Error; err != nil {
			return err
		}
		profile.UserID = user.UserID
		return tx.Create(profile).Error
	}); err != nil {
		return err
	}
	invalidateNotificationSenderCache(context.Background(), r.cache, user.UserID)
	invalidateExploreCache(context.Background(), r.cache)
	return nil
}

func (r *UserRepository) FindByEmail(email string) (*model.User, error) {
	var user model.User
	err := r.readDB.Preload("Role").Where("email = ?", email).First(&user).Error
	return &user, err
}

func (r *UserRepository) FindByUserName(username string) (*model.User, error) {
	var user model.User
	err := r.readDB.Preload("Role").Where("user_name = ?", username).First(&user).Error
	return &user, err
}

func (r *UserRepository) ExistsByEmail(email string) (bool, error) {
	var count int64
	err := r.readDB.Model(&model.User{}).Where("email = ?", email).Count(&count).Error
	return count > 0, err
}

func (r *UserRepository) ExistsByUserName(username string) (bool, error) {
	var count int64
	err := r.readDB.Model(&model.User{}).Where("user_name = ?", username).Count(&count).Error
	return count > 0, err
}

func (r *UserRepository) UpdatePassword(userID uint, hashedPassword string) error {
	if err := r.writeDB.Model(&model.User{}).Where("user_id = ?", userID).Update("password", hashedPassword).Error; err != nil {
		return err
	}
	r.invalidateUser(userID)
	invalidateNotificationSenderCache(context.Background(), r.cache, userID)
	return nil
}

func (r *UserRepository) UpdateInfo(userID uint, name, username string) error {
	if err := r.writeDB.Model(&model.User{}).Where("user_id = ?", userID).
		Updates(map[string]any{"name": name, "user_name": username}).Error; err != nil {
		return err
	}
	r.invalidateUser(userID)
	invalidateNotificationSenderCache(context.Background(), r.cache, userID)
	invalidateExploreCache(context.Background(), r.cache)
	r.invalidateUserThreadCaches(userID)
	return nil
}

func (r *UserRepository) UpdatePremium(userID uint, premium bool) error {
	if err := r.writeDB.Model(&model.User{}).Where("user_id = ?", userID).Update("is_premium", premium).Error; err != nil {
		return err
	}
	r.invalidateUser(userID)
	invalidateNotificationSenderCache(context.Background(), r.cache, userID)
	invalidateExploreCache(context.Background(), r.cache)
	r.invalidateUserThreadCaches(userID)
	return nil
}

func (r *UserRepository) invalidateUserThreadCaches(userID uint) {
	if r.cache == nil {
		return
	}
	ctx := context.Background()
	_ = r.cache.DeleteByPattern(ctx, fmt.Sprintf("thread:userfeed:%d:*", userID))
	_ = r.cache.DeleteByPattern(ctx, fmt.Sprintf("thread:usermedia:%d:*", userID))
	_ = r.cache.DeleteByPattern(ctx, "thread:following:*")
	_ = r.cache.DeleteByPattern(ctx, "thread:roots:*")
}

func (r *UserRepository) FindAllPaginated(page, limit int) ([]model.User, int64, error) {
	var users []model.User
	var total int64

	offset := (page - 1) * limit

	if err := r.readDB.Model(&model.User{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := r.readDB.Preload("Role").Offset(offset).Limit(limit).Find(&users).Error
	return users, total, err
}

func (r *UserRepository) invalidateUser(userID uint) {
	if r.cache == nil {
		return
	}
	r.cache.Delete(context.Background(), userKey(userID))
}
