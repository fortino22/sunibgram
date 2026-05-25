package bookmarkcollectionmodel

import (
	"time"

	usermodel "github.com/sunibgram/domain/model/user"
)

type BookmarkCollection struct {
	ID          uint           `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID      uint           `gorm:"not null;index;uniqueIndex:idx_bookmark_collections_user_name,priority:1" json:"user_id"`
	User        usermodel.User `gorm:"foreignKey:UserID;references:UserID;constraint:OnUpdate:NO ACTION,OnDelete:CASCADE" json:"-"`
	Name        string         `gorm:"type:varchar(50);not null;uniqueIndex:idx_bookmark_collections_user_name,priority:2" json:"name"`
	Description *string        `gorm:"type:text" json:"description,omitempty"`
	ItemCount   int            `gorm:"not null;default:0" json:"item_count"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
}

type BookmarkCollectionItem struct {
	CollectionID uint               `gorm:"primaryKey;autoIncrement:false;index:idx_bci_collection_thread,priority:1" json:"collection_id"`
	Collection   BookmarkCollection `gorm:"foreignKey:CollectionID;references:ID;constraint:OnUpdate:NO ACTION,OnDelete:CASCADE" json:"-"`
	UserID       uint               `gorm:"primaryKey;autoIncrement:false;index:idx_bci_user_thread,priority:1" json:"user_id"`
	ThreadID     uint               `gorm:"primaryKey;autoIncrement:false;index:idx_bci_collection_thread,priority:2;index:idx_bci_user_thread,priority:2" json:"thread_id"`
	CreatedAt    time.Time          `json:"created_at"`
}

const MaxCollectionsPerUser = 50
const MaxCollectionNameLen = 50
