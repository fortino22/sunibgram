package dto

import (
	"fmt"
	"strings"
	"time"

	"github.com/sunibgram/dto/utils"
)

type EventTime struct {
	time.Time
}

const eventTimeLayout = "2006-01-02 15:04:05"

func (et EventTime) MarshalJSON() ([]byte, error) {
	if et.Time.IsZero() {
		return []byte("null"), nil
	}
	return []byte(`"` + et.Time.UTC().Format(eventTimeLayout) + `"`), nil
}

func (et *EventTime) UnmarshalJSON(data []byte) error {
	s := strings.Trim(string(data), `"`)
	if s == "null" || s == "" {
		et.Time = time.Time{}
		return nil
	}

	layouts := []string{
		eventTimeLayout,
		"2006-01-02 15:04",
		time.RFC3339Nano,
	}
	for _, layout := range layouts {
		t, err := time.Parse(layout, s)
		if err == nil {
			et.Time = t.UTC()
			return nil
		}
	}
	return fmt.Errorf("invalid event time %q", s)
}

type CreateEventRequest struct {
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Location    string    `json:"location"`
	Type        string    `json:"type" form:"type"`
	Status      string    `json:"status"`
	StartTime   EventTime `json:"start_time"`
	EndTime     EventTime `json:"end_time"`
}

type UpdateEventRequest struct {
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Location    string    `json:"location"`
	StartTime   EventTime `json:"start_time"`
	EndTime     EventTime `json:"end_time"`
}

type EventResponse struct {
	ID               string              `json:"id"`
	Title            string              `json:"title"`
	Description      string              `json:"description"`
	Location         string              `json:"location"`
	Type             string              `json:"type"`
	Status           string              `json:"status"`
	StartTime        EventTime           `json:"start_time"`
	EndTime          EventTime           `json:"end_time"`
	BannerURL        string              `json:"banner_url,omitempty"`
	AttachmentURL    string              `json:"attachment_url,omitempty"`
	ParticipantCount int                 `json:"participant_count"`
	ThreadCount      int                 `json:"thread_count"`
	LikeCount        int                 `json:"like_count"`
	RepostCount      int                 `json:"repost_count"`
	WeekThreadCount  int                 `json:"week_thread_count"`
	WeekLikeCount    int                 `json:"week_like_count"`
	WeekRepostCount  int                 `json:"week_repost_count"`
	CreatedAt        utils.FormattedTime `json:"created_at"`
	UpdatedAt        utils.FormattedTime `json:"updated_at"`
}

type EventGroupsResponse struct {
	Trending []EventResponse `json:"trending"`
	Running  []EventResponse `json:"running"`
	Future   []EventResponse `json:"future"`
}
