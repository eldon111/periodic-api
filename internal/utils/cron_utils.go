package utils

import (
	"time"

	"github.com/robfig/cron/v3"
)

// CalculateNextExecution calculates the next execution time for a scheduled item
// Returns nil if the item should not execute again (expired or one-time item in the past)
func CalculateNextExecution(startsAt time.Time, repeats bool, cronExpression *string, expiration *time.Time) *time.Time {
	now := time.Now()

	// For non-repeating items
	if !repeats {
		// If starts in the future, return startsAt
		if startsAt.After(now) {
			return &startsAt
		}
		// If starts in the past, no next execution
		return nil
	}

	// For repeating items, we need a cron expression
	if cronExpression == nil || *cronExpression == "" {
		return nil
	}

	// Parse the cron expression
	parser := cron.NewParser(cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow)
	schedule, err := parser.Parse(*cronExpression)
	if err != nil {
		// Invalid cron expression
		return nil
	}

	// Calculate next execution time
	nextTime := schedule.Next(now)

	// Check if next execution is after expiration
	if expiration != nil && nextTime.After(*expiration) {
		return nil
	}

	// Ensure we don't schedule before the original start time
	if nextTime.Before(startsAt) {
		nextTime = schedule.Next(startsAt.Add(-time.Second))
		// Check expiration again after adjusting for start time
		if expiration != nil && nextTime.After(*expiration) {
			return nil
		}
	}

	return &nextTime
}

// ValidateCronExpression validates if a cron expression is valid
func ValidateCronExpression(cronExpression string) error {
	parser := cron.NewParser(cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow)
	_, err := parser.Parse(cronExpression)
	return err
}
