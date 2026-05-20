package usecase

import (
	"context"
	"errors"
	"strings"
	"unicode/utf8"

	"github.com/sunibgram/domain/model"
	"github.com/sunibgram/domain/repository"
	"github.com/sunibgram/domain/service"
)

const (
	maxCommunityRules      = 10
	maxCommunityRuleLength = 500
)

type CommunityRuleService struct {
	communityRepo repository.ICommunityRepository
	memberRepo    repository.ICommunityMemberRepository
	ruleRepo      repository.ICommunityRuleRepository
}

func NewCommunityRuleService(
	communityRepo repository.ICommunityRepository,
	memberRepo repository.ICommunityMemberRepository,
	ruleRepo repository.ICommunityRuleRepository,
) service.ICommunityRuleService {
	return &CommunityRuleService{communityRepo: communityRepo, memberRepo: memberRepo, ruleRepo: ruleRepo}
}

func (s *CommunityRuleService) ListByCommunity(communityID uint) ([]model.CommunityRule, error) {
	if _, err := s.communityRepo.FindByID(communityID); err != nil {
		return nil, errors.New("community not found")
	}
	return s.ruleRepo.ListByCommunity(communityID)
}

func (s *CommunityRuleService) Create(_ context.Context, communityID, actorUserID uint, orderNum int, text string) (*model.CommunityRule, error) {
	if err := s.ensureCanMutate(communityID, actorUserID); err != nil {
		return nil, err
	}
	if err := validateCommunityRule(orderNum, text); err != nil {
		return nil, err
	}
	count, err := s.ruleRepo.CountByCommunity(communityID)
	if err != nil {
		return nil, err
	}
	if count >= maxCommunityRules {
		return nil, errors.New("maximum 10 rules allowed")
	}
	rule := model.CommunityRule{
		CommunityID: communityID,
		OrderNum:    orderNum,
		Text:        strings.TrimSpace(text),
	}
	return s.ruleRepo.Create(rule)
}

func (s *CommunityRuleService) Update(_ context.Context, communityID, actorUserID, ruleID uint, orderNum int, text string) (*model.CommunityRule, error) {
	if err := s.ensureCanMutate(communityID, actorUserID); err != nil {
		return nil, err
	}
	if err := validateCommunityRule(orderNum, text); err != nil {
		return nil, err
	}
	existing, err := s.ruleRepo.FindByID(ruleID)
	if err != nil || existing.CommunityID != communityID {
		return nil, errors.New("rule not found")
	}
	existing.OrderNum = orderNum
	existing.Text = strings.TrimSpace(text)
	return s.ruleRepo.Update(*existing)
}

func (s *CommunityRuleService) Delete(_ context.Context, communityID, actorUserID, ruleID uint) error {
	if err := s.ensureCanMutate(communityID, actorUserID); err != nil {
		return err
	}
	existing, err := s.ruleRepo.FindByID(ruleID)
	if err != nil || existing.CommunityID != communityID {
		return errors.New("rule not found")
	}
	return s.ruleRepo.Delete(ruleID)
}

func (s *CommunityRuleService) ensureCanMutate(communityID, actorUserID uint) error {
	community, err := s.communityRepo.FindByID(communityID)
	if err != nil {
		return errors.New("community not found")
	}
	if !community.IsWritable() {
		return errors.New("community is inactive")
	}
	actor, err := s.memberRepo.FindMember(communityID, actorUserID)
	if err != nil || !isModeratorOrAbove(actor.Role) {
		return errors.New("forbidden")
	}
	return nil
}

func validateCommunityRule(orderNum int, text string) error {
	text = strings.TrimSpace(text)
	if text == "" {
		return errors.New("text is required")
	}
	if utf8.RuneCountInString(text) > maxCommunityRuleLength {
		return errors.New("text must be 500 characters or less")
	}
	if orderNum < 1 {
		return errors.New("order_num must be at least 1")
	}
	return nil
}
