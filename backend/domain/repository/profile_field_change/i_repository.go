package profilefieldchangerepoiface

import (
	"time"

	profilefieldchangemodel "github.com/sunibgram/domain/model/profile_field_change"
)

type IProfileFieldChangeRepository interface {
	// Record appends one change-log row for the given field.
	Record(userID uint, field profilefieldchangemodel.Field, changedAt time.Time, oldValue, newValue string) error
	// LatestByUser returns the most recent change time per field for a user.
	// Fields never changed are absent from the map.
	LatestByUser(userID uint) (map[profilefieldchangemodel.Field]time.Time, error)
	// HistoryByUserField returns all change rows for one field, newest first.
	HistoryByUserField(userID uint, field profilefieldchangemodel.Field) ([]profilefieldchangemodel.ProfileFieldChange, error)
}
