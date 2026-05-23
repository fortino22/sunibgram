package threadpolldto

import (
	"math"

	threadpollmodel "github.com/sunibgram/domain/model/thread_poll"
	utils "github.com/sunibgram/webapi/dto/_shared"
)

type PollOptionResponse struct {
	OptionID  uint    `json:"option_id"`
	Label     string  `json:"label"`
	ImageURL  string  `json:"image_url,omitempty"`
	VoteCount int     `json:"vote_count"`
	Percent   float64 `json:"percent"`
}

type PollResponse struct {
	PollID            uint                 `json:"poll_id"`
	Question          string               `json:"question,omitempty"`
	EndsAt            utils.FormattedTime  `json:"ends_at"`
	TotalVotes        int                  `json:"total_votes"`
	Expired           bool                 `json:"expired"`
	UserVotedOptionID *uint                `json:"user_voted_option_id,omitempty"`
	Options           []PollOptionResponse `json:"options"`
}

type VotePollRequest struct {
	OptionID uint `json:"option_id"`
}

type CreatePollDTO struct {
	Question        string
	Labels          []string
	DurationSeconds int
}

func BuildPollResponse(poll *threadpollmodel.ThreadPoll, votedOptionID *uint, urlFn func(bucket, object string) string) *PollResponse {
	if poll == nil {
		return nil
	}
	options := make([]PollOptionResponse, len(poll.Options))
	total := poll.TotalVotes
	for i, opt := range poll.Options {
		percent := 0.0
		if total > 0 {
			percent = math.Round(float64(opt.VoteCount)/float64(total)*1000) / 10
		}
		var imageURL string
		if urlFn != nil {
			imageURL = urlFn(opt.Bucket, opt.ObjectName)
		}
		options[i] = PollOptionResponse{
			OptionID:  opt.OptionID,
			Label:     opt.Label,
			ImageURL:  imageURL,
			VoteCount: opt.VoteCount,
			Percent:   percent,
		}
	}
	return &PollResponse{
		PollID:            poll.PollID,
		Question:          poll.Question,
		EndsAt:            utils.Wrap(poll.EndsAt),
		TotalVotes:        poll.TotalVotes,
		Expired:           poll.IsExpired(),
		UserVotedOptionID: votedOptionID,
		Options:           options,
	}
}
