package utils

import (
	"strings"
	"time"
)

type FormattedTime struct {
	time.Time
}

const layout = "2006-01-02 15:04"

var wib = time.FixedZone("WIB", 7*60*60)

func (ft FormattedTime) MarshalJSON() ([]byte, error) {
	return []byte(`"` + ft.Time.In(wib).Format(layout) + `"`), nil
}

func (ft *FormattedTime) UnmarshalJSON(data []byte) error {
	s := strings.Trim(string(data), `"`)
	if s == "null" || s == "" {
		ft.Time = time.Time{}
		return nil
	}
	t, err := time.Parse(layout, s)
	if err != nil {
		t, err = time.Parse(time.RFC3339Nano, s)
		if err != nil {
			return err
		}
	}
	ft.Time = t
	return nil
}

func Wrap(t time.Time) FormattedTime {
	return FormattedTime{t}
}
