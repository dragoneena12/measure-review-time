package printer

import (
	"time"
)

func formatDuration(d time.Duration) int {
	return int(d.Minutes())
}

func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}
