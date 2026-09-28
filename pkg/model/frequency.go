package model

import "time"

type Frequency string

const (
	FrequencyHourly  Frequency = "hourly"
	FrequencyDaily   Frequency = "daily"
	FrequencyWeekly  Frequency = "weekly"
	FrequencyMonthly Frequency = "monthly"
)

func (f Frequency) Duration() time.Duration {
	switch f {
	case FrequencyHourly:
		return time.Hour
	case FrequencyDaily:
		return 24 * time.Hour
	case FrequencyWeekly:
		return 7 * 24 * time.Hour
	case FrequencyMonthly:
		return 30 * 24 * time.Hour
	default:
		return 0
	}
}
