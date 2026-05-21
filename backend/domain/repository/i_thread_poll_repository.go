package repository

import "github.com/sunibgram/domain/model"

type IThreadPollRepository interface {
	Create(poll *model.ThreadPoll, options []model.ThreadPollOption) error
	FindByThreadID(threadID uint) (*model.ThreadPoll, error)
	FindByThreadIDs(threadIDs []uint) (map[uint]*model.ThreadPoll, error)
	FindUserVotes(userID uint, pollIDs []uint) (map[uint]uint, error)
	FindOptionByID(optionID uint) (*model.ThreadPollOption, error)
	CastVote(pollID, optionID, userID uint) error
	DeleteByThreadID(threadID uint) ([]model.ThreadPollOption, error)
}
