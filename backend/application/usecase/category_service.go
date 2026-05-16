package usecase

import (
	"errors"
	"strings"

	"github.com/google/uuid"
	"github.com/sunibgram/domain/model"
	"github.com/sunibgram/domain/repository"
	"github.com/sunibgram/domain/service"
)

type CategoryService struct {
	categoryRepo repository.ICategoryRepository
}

func NewCategoryService(categoryRepo repository.ICategoryRepository) service.ICategoryService {
	return &CategoryService{categoryRepo: categoryRepo}
}

func (s *CategoryService) Create(name string) error {
	name = strings.TrimSpace(name)
	if name == "" {
		return errors.New("category name is required")
	}
	return s.categoryRepo.Save(&model.Category{Name: name})
}

func (s *CategoryService) GetAll() ([]model.Category, error) {
	return s.categoryRepo.FindAll()
}

func (s *CategoryService) Update(id uuid.UUID, name string) error {
	name = strings.TrimSpace(name)
	if name == "" {
		return errors.New("category name is required")
	}

	category, err := s.categoryRepo.FindByID(id)
	if err != nil {
		return errors.New("category not found")
	}
	category.Name = name
	return s.categoryRepo.Update(category)
}

func (s *CategoryService) Delete(id uuid.UUID) error {
	if _, err := s.categoryRepo.FindByID(id); err != nil {
		return errors.New("category not found")
	}
	return s.categoryRepo.Delete(id)
}
