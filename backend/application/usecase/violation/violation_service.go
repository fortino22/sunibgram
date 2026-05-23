package violationuc

import (
	"errors"
	"strings"
	"unicode/utf8"

	"github.com/google/uuid"

	violationmodel "github.com/sunibgram/domain/model/violation"
	violationrepoiface "github.com/sunibgram/domain/repository/violation"
	violationsvciface "github.com/sunibgram/domain/service/violation"
)

type ViolationService struct {
	violationRepo violationrepoiface.IViolationRepository
}

func NewViolationService(violationRepo violationrepoiface.IViolationRepository) violationsvciface.IViolationService {
	return &ViolationService{violationRepo: violationRepo}
}

func (s *ViolationService) Create(name, description string) error {
	name = strings.TrimSpace(name)
	description = strings.TrimSpace(description)
	if name == "" {
		return errors.New("violation name is required")
	}
	if utf8.RuneCountInString(name) > 100 {
		return errors.New("name must be 100 characters or less")
	}
	if description == "" {
		return errors.New("violation description is required")
	}
	if utf8.RuneCountInString(description) > 2000 {
		return errors.New("description must be 2000 characters or less")
	}
	return s.violationRepo.Save(&violationmodel.Violation{Name: name, Description: description})
}

func (s *ViolationService) GetAll() ([]violationmodel.Violation, error) {
	return s.violationRepo.FindAll()
}

func (s *ViolationService) GetByID(id uuid.UUID) (*violationmodel.Violation, error) {
	violation, err := s.violationRepo.FindByID(id)
	if err != nil {
		return nil, errors.New("violation not found")
	}
	return violation, nil
}

func (s *ViolationService) Update(id uuid.UUID, name, description string) error {
	name = strings.TrimSpace(name)
	description = strings.TrimSpace(description)
	if name == "" {
		return errors.New("violation name is required")
	}
	if utf8.RuneCountInString(name) > 100 {
		return errors.New("name must be 100 characters or less")
	}
	if description == "" {
		return errors.New("violation description is required")
	}
	if utf8.RuneCountInString(description) > 2000 {
		return errors.New("description must be 2000 characters or less")
	}
	violation, err := s.violationRepo.FindByID(id)
	if err != nil {
		return errors.New("violation not found")
	}
	violation.Name = name
	violation.Description = description
	return s.violationRepo.Update(violation)
}

func (s *ViolationService) Delete(id uuid.UUID) error {
	if _, err := s.violationRepo.FindByID(id); err != nil {
		return errors.New("violation not found")
	}
	return s.violationRepo.Delete(id)
}
