package cli

import (
	"fmt"
	"time"
)

const day = 24 * time.Hour
const month = 28 * day
const year = 12 * month

func TimeSince(d time.Duration) string {
	if d >= year {
		years := d / year

		if years <= 1 {
			return "1y ago"
		}

		return fmt.Sprintf("%dy ago", years)
	}

	if d >= month && d < year {
		months := d / month

		if months <= 1 {
			return "1m ago"
		}

		return fmt.Sprintf("%dm ago", months)
	}

	if d >= day && d < month {
		days := d / day
		return fmt.Sprintf("%dd ago", days)
	}

	if d >= time.Hour && d < day {
		hours := d / time.Hour
		return fmt.Sprintf("%dh ago", hours)
	}

	if d >= time.Minute && d < time.Hour {
		minutes := d / time.Minute
		return fmt.Sprintf("%dmin ago", minutes)
	}

	return "just now"
}
