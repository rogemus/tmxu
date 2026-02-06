package cli

import (
	"fmt"
	"time"
)

var oneMin int = int(time.Minute.Seconds())
var oneHour int = int(time.Hour.Seconds())
var oneDay int = int(24 * oneHour)
var oneMonth int = int(4 * 7 * oneDay)
var oneYear int = int(12 * oneMonth)

func TimeSince(d time.Duration) string {
	seconds := int(d.Seconds())

	switch {
	case seconds < 60:
		return "just now"
	case seconds < oneMin:
		minutes := seconds / oneMin

		if minutes <= 1 {
			return "1min ago"
		}

		return fmt.Sprintf("%dmin ago", minutes)
	case seconds < oneHour:
		hours := seconds / oneHour

		if hours <= 1 {
			return "1h ago"
		}

		return fmt.Sprintf("%dh ago", hours)
	case seconds < oneDay:
		days := seconds / oneDay

		if days <= 1 {
			return "1d ago"
		}
		return fmt.Sprintf("%dd ago", days)
	case seconds < oneMonth:
		months := seconds / oneMonth

		if months <= 1 {
			return "1m ago"
		}

		return fmt.Sprintf("%dm ago", months)
	default:
		years := seconds / oneYear

		if years <= 1 {
			return "1y ago"
		}

		return fmt.Sprintf("%dy ago", years)
	}
}
