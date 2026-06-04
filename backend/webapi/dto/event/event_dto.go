package eventdto

import (
	"fmt"
	"strings"
	"time"

	utils "github.com/sunibgram/webapi/dto/_shared"
	eventreviewdto "github.com/sunibgram/webapi/dto/event_review"
	threaddto "github.com/sunibgram/webapi/dto/thread"
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
	Type        string    `json:"type" form:"type"`
	StartTime   EventTime `json:"start_time"`
	EndTime     EventTime `json:"end_time"`
}

// FacultyLocationMini is a campus summary in an event's region audience.
type FacultyLocationMini struct {
	ID           uint   `json:"id"`
	Name         string `json:"name"`
	Region       string `json:"region"`
	IsAllRegions bool   `json:"is_all_regions"`
}

type EventResponse struct {
	ID            string                `json:"id"`
	UserID        uint                  `json:"user_id"`
	Title         string                `json:"title"`
	Description   string                `json:"description"`
	Location      string                `json:"location"`
	Type          string                `json:"type"`
	Status        string                `json:"status"`
	StartTime     EventTime             `json:"start_time"`
	EndTime       EventTime             `json:"end_time"`
	BannerURL     string                `json:"banner_url,omitempty"`
	AttachmentURL string                `json:"attachment_url,omitempty"`
	DocURLs       []string              `json:"doc_urls,omitempty"`
	Audience      []FacultyLocationMini `json:"audience"`
	CreatedAt     utils.FormattedTime   `json:"created_at"`
	UpdatedAt     utils.FormattedTime   `json:"updated_at"`
}

type EventListResponse struct {
	ID          string                `json:"id"`
	UserID      uint                  `json:"user_id"`
	Title       string                `json:"title"`
	Description string                `json:"description"`
	Location    string                `json:"location"`
	Type        string                `json:"type"`
	StartTime   EventTime             `json:"start_time"`
	EndTime     EventTime             `json:"end_time"`
	BannerURL   string                `json:"banner_url,omitempty"`
	Audience    []FacultyLocationMini `json:"audience"`
	CreatedAt   utils.FormattedTime   `json:"created_at"`
	UpdatedAt   utils.FormattedTime   `json:"updated_at"`
}

type EventWithReviewHistoryResponse struct {
	EventResponse
	Notes []eventreviewdto.EventReviewNoteResponse `json:"notes"`
}

type AdminEventWithReviewHistoryResponse struct {
	EventResponse
	Notes []eventreviewdto.EventAdminReviewNoteResponse `json:"notes"`
}

type EventGroupsResponse struct {
	Trending []EventListResponse `json:"trending"`
	Running  []EventListResponse `json:"running"`
	Future   []EventListResponse `json:"future"`
	Past     []EventListResponse `json:"past"`
}

type PaginatedEventResponse struct {
	Data []EventResponse    `json:"data"`
	Meta threaddto.FeedMeta `json:"meta"`
}

type PaginatedEventWithReviewHistoryResponse struct {
	Data []AdminEventWithReviewHistoryResponse `json:"data"`
	Meta threaddto.FeedMeta                    `json:"meta"`
}

type PaginatedMyEventResponse struct {
	Data []EventWithReviewHistoryResponse `json:"data"`
	Meta threaddto.FeedMeta               `json:"meta"`
}

type EventMiniResponse struct {
	ID        string `json:"id"`
	Title     string `json:"title"`
	BannerURL string `json:"banner_url,omitempty"`
}
