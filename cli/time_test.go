package cli

import (
	"testing"
	"time"
)

func TestTimeSince(t *testing.T) {
	const day = 24 * time.Hour
	const month = 28 * day
	const year = 12 * month

	tests := []struct {
		name  string
		since time.Duration
		want  string
	}{
		{
			name:  "zero duration returns just now",
			since: 0,
			want:  "just now",
		},
		{
			name:  "negative duration returns just now",
			since: -5 * time.Second,
			want:  "just now",
		},
		{
			name:  "10 seconds returns just now",
			since: 10 * time.Second,
			want:  "just now",
		},
		{
			name:  "59 seconds returns just now",
			since: 59 * time.Second,
			want:  "just now",
		},
		{
			name:  "exactly 60 seconds returns 1 min ago",
			since: 60 * time.Second,
			want:  "1min ago",
		},
		{
			name:  "65 seconds returns 1 min ago",
			since: 65 * time.Second,
			want:  "1min ago",
		},
		{
			name:  "119 seconds still returns 1 min ago",
			since: 119 * time.Second,
			want:  "1min ago",
		},
		{
			name:  "130 seconds returns 2 min ago",
			since: 130 * time.Second,
			want:  "2min ago",
		},
		{
			name:  "59 minutes returns 59 min ago",
			since: 59 * time.Minute,
			want:  "59min ago",
		},
		{
			name:  "exactly 1 hour returns 1h ago",
			since: time.Hour,
			want:  "1h ago",
		},
		{
			name:  "90 minutes still returns 1h ago",
			since: 90 * time.Minute,
			want:  "1h ago",
		},
		{
			name:  "5 hours returns 5h ago",
			since: 5 * time.Hour,
			want:  "5h ago",
		},
		{
			name:  "23 hours returns 23h ago",
			since: 23 * time.Hour,
			want:  "23h ago",
		},
		{
			name:  "exactly 1 day returns 1d ago",
			since: day,
			want:  "1d ago",
		},
		{
			name:  "3 days returns 3d ago",
			since: 3 * day,
			want:  "3d ago",
		},
		{
			name:  "27 days returns 27d ago",
			since: 27 * day,
			want:  "27d ago",
		},
		{
			name:  "exactly 1 month returns 1m ago",
			since: month,
			want:  "1m ago",
		},
		{
			name:  "3 months returns 3m ago",
			since: 3 * month,
			want:  "3m ago",
		},
		{
			name:  "exactly 1 year returns 1y ago",
			since: year,
			want:  "1y ago",
		},
		{
			name:  "3 years returns 3y ago",
			since: 3 * year,
			want:  "3y ago",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := TimeSince(test.since)

			if got != test.want {
				t.Errorf("TimeSince(%v) = %s, want %s", test.since, got, test.want)
			}
		})
	}
}
