package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/sunibgram/domain/model"
	domainrepo "github.com/sunibgram/domain/repository"
	"github.com/sunibgram/infrastructure/cache"
	"gorm.io/gorm"
)

const categoryTTL = 24 * time.Hour

type categoryRepository struct {
	writeDB *gorm.DB
	readDB  *gorm.DB
	cache   *cache.Cache
}

func NewCategoryRepository(writeDB, readDB *gorm.DB, c *cache.Cache) domainrepo.ICategoryRepository {
	return &categoryRepository{writeDB: writeDB, readDB: readDB, cache: c}
}

func categoryAllKey() string {
	return "category:all"
}

func categoryIDKey(id uuid.UUID) string {
	return fmt.Sprintf("category:id:%s", id.String())
}

func (r *categoryRepository) Save(category *model.Category) error {
	if err := r.writeDB.Create(category).Error; err != nil {
		return err
	}
	r.invalidateCategoryList()
	return nil
}

func (r *categoryRepository) FindAll() ([]model.Category, error) {
	ctx := context.Background()
	var categories []model.Category
	if r.cache != nil {
		if err := r.cache.Get(ctx, categoryAllKey(), &categories); err == nil {
			return categories, nil
		}
	}

	if err := r.readDB.Order("name ASC").Find(&categories).Error; err != nil {
		return nil, err
	}
	if r.cache != nil {
		_ = r.cache.Set(ctx, categoryAllKey(), categories, categoryTTL)
		for _, category := range categories {
			_ = r.cache.Set(ctx, categoryIDKey(category.ID), category, categoryTTL)
		}
	}
	return categories, nil
}

func (r *categoryRepository) FindByID(id uuid.UUID) (*model.Category, error) {
	ctx := context.Background()
	var category model.Category
	if r.cache != nil {
		if err := r.cache.Get(ctx, categoryIDKey(id), &category); err == nil {
			return &category, nil
		}
	}

	categories, err := r.FindAll()
	if err != nil {
		return nil, err
	}
	for i := range categories {
		if categories[i].ID == id {
			return &categories[i], nil
		}
	}
	return nil, gorm.ErrRecordNotFound
}

func (r *categoryRepository) Update(category *model.Category) error {
	result := r.writeDB.Model(&model.Category{}).
		Where("id = ?", category.ID).
		Updates(map[string]any{
			"name": category.Name,
		})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("category not found")
	}
	r.invalidateCategoryUpdate(category.ID)
	return nil
}

func (r *categoryRepository) Delete(id uuid.UUID) error {
	err := r.writeDB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&model.Community{}).
			Where("category_id = ?", id).
			UpdateColumns(map[string]any{"category_id": nil}).Error; err != nil {
			return err
		}
		result := tx.Delete(&model.Category{}, "id = ?", id)
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			return errors.New("category not found")
		}
		return nil
	})
	if err != nil {
		return err
	}
	r.invalidateCategoryUpdate(id)
	return nil
}

func (r *categoryRepository) invalidateCategoryList() {
	if r.cache == nil {
		return
	}
	_ = r.cache.Delete(context.Background(), categoryAllKey())
}

func (r *categoryRepository) invalidateCategoryUpdate(id uuid.UUID) {
	if r.cache == nil {
		return
	}
	ctx := context.Background()
	_ = r.cache.Delete(ctx, categoryAllKey(), categoryIDKey(id))
	_ = r.cache.Delete(ctx, "discover:communities")
	_ = r.cache.InvalidateTags(ctx,
		communityIDGlobalTag(),
		communityListsGlobalTag(),
	)
	invalidateExploreCache(ctx, r.cache)
}
