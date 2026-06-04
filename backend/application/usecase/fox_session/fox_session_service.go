package foxsessionuc

import (
	"context"

	helper "github.com/sunibgram/application/helper"
	foxsessionrepoiface "github.com/sunibgram/domain/repository/fox_session"
	foxsessionsvciface "github.com/sunibgram/domain/service/fox_session"
)

type FoxSessionService struct {
	repo foxsessionrepoiface.IFoxSessionRepository
}

func NewFoxSessionService(repo foxsessionrepoiface.IFoxSessionRepository) foxsessionsvciface.IFoxSessionService {
	return &FoxSessionService{repo: repo}
}

func (s *FoxSessionService) ListSessions(ctx context.Context, userID uint, page, limit int, q string) ([]foxsessionrepoiface.SessionPreview, int64, error) {
	if s.repo == nil {
		return nil, 0, nil
	}
	page, limit = helper.NormalizePageLimit(page, limit, 20)
	offset := (page - 1) * limit
	return s.repo.ListSessions(userID, offset, limit, q)
}

func (s *FoxSessionService) GetSessionTurns(ctx context.Context, userID, sessionID uint) ([]foxsessionrepoiface.ChatTurn, error) {
	if s.repo == nil {
		return nil, nil
	}
	return s.repo.GetSessionTurns(userID, sessionID)
}
