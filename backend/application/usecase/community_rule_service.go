package usecase

import (
	"errors"
	"strings"
	"unicode/utf8"

	communityrulemodel "github.com/sunibgram/domain/model/community_rule"
	"github.com/sunibgram/domain/repository"
	"github.com/sunibgram/domain/service"
)

const (
	maxCommunityRules      = 10
	maxCommunityRuleLength = 500
)

type CommunityRuleService struct {
	communityRepo repository.ICommunityRepository
	ruleRepo      repository.ICommunityRuleRepository
}

func NewCommunityRuleService(
	communityRepo repository.ICommunityRepository,
	ruleRepo repository.ICommunityRuleRepository,
) service.ICommunityRuleService {
	return &CommunityRuleService{communityRepo: communityRepo, ruleRepo: ruleRepo}
}

func (s *CommunityRuleService) ListByCommunity(communityID uint) ([]communityrulemodel.CommunityRule, error) {
	if _, err := s.communityRepo.FindByID(communityID); err != nil {
		return nil, errors.New("community not found")
	}
	return s.ruleRepo.ListByCommunity(communityID)
}

func dedupeRuleIDs(ids []uint) []uint {
	seen := make(map[uint]struct{}, len(ids))
	out := make([]uint, 0, len(ids))
	for _, id := range ids {
		if id == 0 {
			continue
		}
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		out = append(out, id)
	}
	return out
}

func buildCommunityRules(communityID uint, texts []string) []communityrulemodel.CommunityRule {
	rules := make([]communityrulemodel.CommunityRule, len(texts))
	for i, t := range texts {
		rules[i] = communityrulemodel.CommunityRule{
			CommunityID: communityID,
			OrderNum:    i + 1,
			Text:        t,
		}
	}
	return rules
}

func sanitizeCommunityRules(raw []string) ([]string, error) {
	cleaned := make([]string, 0, len(raw))
	for _, r := range raw {
		t := strings.TrimSpace(r)
		if t == "" {
			continue
		}
		if utf8.RuneCountInString(t) > maxCommunityRuleLength {
			return nil, errors.New("each rule must be 500 characters or less")
		}
		cleaned = append(cleaned, t)
	}
	if len(cleaned) > maxCommunityRules {
		return nil, errors.New("maximum 10 rules allowed")
	}
	return cleaned, nil
}
