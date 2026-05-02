package utils

import (
	"fmt"
	"time"
)

func TimeAgo(t time.Time) string {
	d := time.Since(t)
	switch {
	case d < time.Minute:
		s := int(d.Seconds())
		if s <= 1 {
			return "1 second ago"
		}
		return fmt.Sprintf("%d seconds ago", s)
	case d < time.Hour:
		m := int(d.Minutes())
		if m == 1 {
			return "1 minute ago"
		}
		return fmt.Sprintf("%d minutes ago", m)
	case d < 24*time.Hour:
		h := int(d.Hours())
		if h == 1 {
			return "1 hour ago"
		}
		return fmt.Sprintf("%d hours ago", h)
	case d < 7*24*time.Hour:
		day := int(d.Hours() / 24)
		if day == 1 {
			return "1 day ago"
		}
		return fmt.Sprintf("%d days ago", day)
	case d < 30*24*time.Hour:
		week := int(d.Hours() / (24 * 7))
		if week == 1 {
			return "1 week ago"
		}
		return fmt.Sprintf("%d weeks ago", week)
	case d < 365*24*time.Hour:
		month := int(d.Hours() / (24 * 30))
		if month == 1 {
			return "1 month ago"
		}
		return fmt.Sprintf("%d months ago", month)
	default:
		year := int(d.Hours() / (24 * 365))
		if year == 1 {
			return "1 year ago"
		}
		return fmt.Sprintf("%d years ago", year)
	}
}
