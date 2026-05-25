package bookmarkcollectiondto

import (
	bookmarkcollectionmodel "github.com/sunibgram/domain/model/bookmark_collection"
	utils "github.com/sunibgram/webapi/dto/_shared"
	threaddto "github.com/sunibgram/webapi/dto/thread"
)

type CreateCollectionRequest struct {
	Name        string  `json:"name"`
	Description *string `json:"description,omitempty"`
}

type UpdateCollectionRequest struct {
	Name        string  `json:"name"`
	Description *string `json:"description,omitempty"`
}

type AddThreadRequest struct {
	ThreadID uint `json:"thread_id"`
}

type CollectionResponse struct {
	ID          uint                `json:"id"`
	Name        string              `json:"name"`
	Description *string             `json:"description,omitempty"`
	ItemCount   int                 `json:"item_count"`
	CreatedAt   utils.FormattedTime `json:"created_at"`
	UpdatedAt   utils.FormattedTime `json:"updated_at"`
}

func ToCollectionResponse(c *bookmarkcollectionmodel.BookmarkCollection) CollectionResponse {
	return CollectionResponse{
		ID:          c.ID,
		Name:        c.Name,
		Description: c.Description,
		ItemCount:   c.ItemCount,
		CreatedAt:   utils.Wrap(c.CreatedAt),
		UpdatedAt:   utils.Wrap(c.UpdatedAt),
	}
}

type CollectionListResponse struct {
	Data []CollectionResponse `json:"data"`
}

type CollectionThreadsResponse struct {
	Data []threaddto.FeedItemResponse `json:"data"`
	Meta threaddto.FeedMeta           `json:"meta"`
}

type CollectionIDsResponse struct {
	CollectionIDs []uint `json:"collection_ids"`
}
