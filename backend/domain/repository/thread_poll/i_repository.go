package threadpollrepoiface

import threadpollmodel "github.com/sunibgram/domain/model/thread_poll"

type IThreadPollRepository interface {
	Create(poll *threadpollmodel.ThreadPoll, options []threadpollmodel.ThreadPollOption) error
	FindByThreadID(threadID uint) (*threadpollmodel.ThreadPoll, error)
	FindByThreadIDs(threadIDs []uint) (map[uint]*threadpollmodel.ThreadPoll, error)
	FindUserVotes(userID uint, pollIDs []uint) (map[uint]uint, error)
	FindOptionByID(optionID uint) (*threadpollmodel.ThreadPollOption, error)
	CastVote(pollID, threadID, optionID, userID uint) error
	DeleteByThreadID(threadID uint) ([]threadpollmodel.ThreadPollOption, error)
}
