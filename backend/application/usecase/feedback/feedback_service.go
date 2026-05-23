package feedbackuc

import (
	"errors"

	"github.com/google/uuid"

	feedbackmodel "github.com/sunibgram/domain/model/feedback"
	feedbackrepoiface "github.com/sunibgram/domain/repository/feedback"
	feedbacksvciface "github.com/sunibgram/domain/service/feedback"
)

const roleAdmin = "admin"

type FeedbackService struct {
	repo feedbackrepoiface.IFeedbackRepository
}

func NewFeedbackService(repo feedbackrepoiface.IFeedbackRepository) feedbacksvciface.IFeedbackService {
	return &FeedbackService{repo: repo}
}

func (s *FeedbackService) Create(userID uint, fType, title, body string) error {
	if userID == 0 {
		return errors.New("unauthorized")
	}
	if err := feedbackmodel.ValidateType(fType); err != nil {
		return err
	}
	cleanTitle, err := feedbackmodel.ValidateTitle(title)
	if err != nil {
		return err
	}
	cleanBody, err := feedbackmodel.ValidateBody(body)
	if err != nil {
		return err
	}
	f := &feedbackmodel.Feedback{
		UserID: userID,
		Type:   fType,
		Title:  cleanTitle,
		Body:   cleanBody,
		Status: feedbackmodel.StatusOpen,
	}
	return s.repo.Save(f)
}

func (s *FeedbackService) GetByID(viewerID uint, id uuid.UUID) (*feedbacksvciface.FeedbackView, error) {
	f, err := s.repo.FindByID(id)
	if err != nil {
		return nil, err
	}
	counts, err := s.repo.GetVoteCounts(id)
	if err != nil {
		return nil, err
	}
	userVote := 0
	if viewerID != 0 {
		v, _, err := s.repo.GetUserVote(id, viewerID)
		if err != nil {
			return nil, err
		}
		userVote = v
	}
	return &feedbacksvciface.FeedbackView{Feedback: *f, Counts: counts, UserVote: userVote}, nil
}

func (s *FeedbackService) List(viewerID uint, fType, status string, page, limit int) ([]feedbacksvciface.FeedbackView, int64, error) {
	if fType != "" {
		if err := feedbackmodel.ValidateType(fType); err != nil {
			return nil, 0, err
		}
	}
	if status != "" {
		if err := feedbackmodel.ValidateStatus(status); err != nil {
			return nil, 0, err
		}
	}
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}

	items, total, err := s.repo.FindAll(feedbackrepoiface.ListFilter{
		Type: fType, Status: status, Page: page, Limit: limit,
	})
	if err != nil {
		return nil, 0, err
	}
	if len(items) == 0 {
		return []feedbacksvciface.FeedbackView{}, total, nil
	}

	ids := make([]uuid.UUID, len(items))
	for i := range items {
		ids[i] = items[i].ID
	}
	countsMap, err := s.repo.GetVoteCountsBulk(ids)
	if err != nil {
		return nil, 0, err
	}
	var votesMap map[uuid.UUID]int
	if viewerID != 0 {
		votesMap, err = s.repo.GetUserVotes(ids, viewerID)
		if err != nil {
			return nil, 0, err
		}
	}

	views := make([]feedbacksvciface.FeedbackView, len(items))
	for i := range items {
		views[i] = feedbacksvciface.FeedbackView{
			Feedback: items[i],
			Counts:   countsMap[items[i].ID],
			UserVote: votesMap[items[i].ID],
		}
	}
	return views, total, nil
}

func (s *FeedbackService) UpdateStatus(adminID uint, id uuid.UUID, status string) error {
	if adminID == 0 {
		return errors.New("unauthorized")
	}
	if err := feedbackmodel.ValidateStatus(status); err != nil {
		return err
	}
	if _, err := s.repo.FindByID(id); err != nil {
		return err
	}
	return s.repo.UpdateStatus(id, status)
}

func (s *FeedbackService) Delete(userID uint, roleName string, id uuid.UUID) error {
	if userID == 0 {
		return errors.New("unauthorized")
	}
	f, err := s.repo.FindByID(id)
	if err != nil {
		return err
	}
	if f.UserID != userID && roleName != roleAdmin {
		return errors.New("forbidden")
	}
	return s.repo.Delete(id)
}

func (s *FeedbackService) Vote(userID uint, id uuid.UUID, value int) error {
	if userID == 0 {
		return errors.New("unauthorized")
	}
	if err := feedbackmodel.ValidateVoteValue(value); err != nil {
		return err
	}
	if _, err := s.repo.FindByID(id); err != nil {
		return err
	}
	existing, found, err := s.repo.GetUserVote(id, userID)
	if err != nil {
		return err
	}
	if !found {
		return s.repo.InsertVote(id, userID, value)
	}
	if existing == value {
		if value == 1 {
			return errors.New("you already liked this feedback")
		}
		return errors.New("you already disliked this feedback")
	}
	return s.repo.UpdateVote(id, userID, value)
}
