package threadpollmodel

import (
	"errors"
	"fmt"
	"mime/multipart"
	"strings"
	"time"

	mediamodel "github.com/sunibgram/domain/model/media"
	threadmodel "github.com/sunibgram/domain/model/thread"
)

const (
	MinPollOptions       = 2
	MaxPollOptions       = 4
	MaxPollLabelLength   = 25
	MinPollDurationSecs  = 300
	MaxPollDurationSecs  = 604800
	MaxPollImageSize     = 5 * 1024 * 1024
	MaxPollQuestionRunes = 280
)

type ThreadPoll struct {
	PollID     uint               `gorm:"primaryKey;autoIncrement" json:"poll_id"`
	ThreadID   uint               `gorm:"not null;uniqueIndex" json:"thread_id"`
	Question   string             `gorm:"type:text" json:"question"`
	EndsAt     time.Time          `gorm:"not null;index" json:"ends_at"`
	TotalVotes int                `gorm:"not null;default:0" json:"total_votes"`
	CreatedAt  time.Time          `json:"created_at"`
	UpdatedAt  time.Time          `json:"updated_at"`
	Options    []ThreadPollOption `gorm:"foreignKey:PollID;references:PollID" json:"options"`
}

type ThreadPollOption struct {
	OptionID   uint      `gorm:"primaryKey;autoIncrement" json:"option_id"`
	PollID     uint      `gorm:"not null;index" json:"poll_id"`
	Label      string    `gorm:"type:varchar(25);not null" json:"label"`
	ObjectName string    `gorm:"type:varchar(500)" json:"object_name,omitempty"`
	Bucket     string    `gorm:"type:varchar(100)" json:"bucket,omitempty"`
	Position   int16     `gorm:"not null" json:"position"`
	VoteCount  int       `gorm:"not null;default:0" json:"vote_count"`
	CreatedAt  time.Time `json:"created_at"`
}

type ThreadPollVote struct {
	VoteID    uint      `gorm:"primaryKey;autoIncrement" json:"vote_id"`
	PollID    uint      `gorm:"not null;index;uniqueIndex:uq_poll_user" json:"poll_id"`
	OptionID  uint      `gorm:"not null;index" json:"option_id"`
	UserID    uint      `gorm:"not null;index;uniqueIndex:uq_poll_user" json:"user_id"`
	CreatedAt time.Time `json:"created_at"`
}

func (p *ThreadPoll) IsExpired() bool {
	return !time.Now().Before(p.EndsAt)
}

func ValidatePollOptionsCount(n int) error {
	if n < MinPollOptions {
		return fmt.Errorf("poll must have at least %d options", MinPollOptions)
	}
	if n > MaxPollOptions {
		return fmt.Errorf("poll can have at most %d options", MaxPollOptions)
	}
	return nil
}

func ValidatePollLabel(label string) error {
	label = strings.TrimSpace(label)
	if label == "" {
		return errors.New("poll option label is required")
	}
	if len([]rune(label)) > MaxPollLabelLength {
		return fmt.Errorf("poll option label must be %d characters or less", MaxPollLabelLength)
	}
	return nil
}

func ValidatePollDuration(seconds int) error {
	if seconds < MinPollDurationSecs {
		return fmt.Errorf("poll duration must be at least %d seconds", MinPollDurationSecs)
	}
	if seconds > MaxPollDurationSecs {
		return fmt.Errorf("poll duration must be at most %d seconds", MaxPollDurationSecs)
	}
	return nil
}

func ValidatePollQuestion(question string) error {
	if len([]rune(question)) > MaxPollQuestionRunes {
		return fmt.Errorf("poll question must be %d characters or less", MaxPollQuestionRunes)
	}
	return nil
}

func ValidatePollImage(header *multipart.FileHeader, file multipart.File) error {
	if err := mediamodel.ValidateUploadSize(header, MaxPollImageSize); err != nil {
		return err
	}
	if err := mediamodel.ValidateImageExt(header.Filename); err != nil {
		return err
	}
	return mediamodel.ValidateImageFile(file)
}

func ValidatePollAllowed(threadType threadmodel.ThreadType) error {
	if threadType != threadmodel.ThreadTypePost {
		return errors.New("polls are only allowed on post-type threads")
	}
	return nil
}
