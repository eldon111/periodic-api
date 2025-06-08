package utils

import (
	"testing"
	"time"
)

func TestCalculateNextExecution(t *testing.T) {
	now := time.Now()
	future := now.Add(24 * time.Hour)
	past := now.Add(-24 * time.Hour)

	tests := []struct {
		name           string
		startsAt       time.Time
		repeats        bool
		cronExpression *string
		expiration     *time.Time
		expectNil      bool
		description    string
	}{
		{
			name:        "Non-repeating future item",
			startsAt:    future,
			repeats:     false,
			expectNil:   false,
			description: "Should return startsAt for future non-repeating items",
		},
		{
			name:        "Non-repeating past item",
			startsAt:    past,
			repeats:     false,
			expectNil:   true,
			description: "Should return nil for past non-repeating items",
		},
		{
			name:           "Repeating item without cron",
			startsAt:       past,
			repeats:        true,
			cronExpression: nil,
			expectNil:      true,
			description:    "Should return nil for repeating items without cron expression",
		},
		{
			name:           "Repeating item with invalid cron",
			startsAt:       past,
			repeats:        true,
			cronExpression: stringPtr("invalid cron"),
			expectNil:      true,
			description:    "Should return nil for invalid cron expressions",
		},
		{
			name:           "Repeating item with valid cron",
			startsAt:       past,
			repeats:        true,
			cronExpression: stringPtr("0 9 * * MON-FRI"), // 9 AM weekdays
			expectNil:      false,
			description:    "Should calculate next execution for valid cron",
		},
		{
			name:           "Repeating item expired",
			startsAt:       past,
			repeats:        true,
			cronExpression: stringPtr("0 9 * * MON-FRI"),
			expiration:     &past,
			expectNil:      true,
			description:    "Should return nil if next execution would be after expiration",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CalculateNextExecution(tt.startsAt, tt.repeats, tt.cronExpression, tt.expiration)

			if tt.expectNil {
				if result != nil {
					t.Errorf("Expected nil but got %v. %s", result, tt.description)
				}
			} else {
				if result == nil {
					t.Errorf("Expected non-nil result but got nil. %s", tt.description)
				} else {
					// For non-repeating future items, should return startsAt
					if !tt.repeats && result.Unix() != tt.startsAt.Unix() {
						t.Errorf("Expected %v but got %v", tt.startsAt, *result)
					}
					// For repeating items, should return a future time
					if tt.repeats && !result.After(now) {
						t.Errorf("Expected future time but got %v", *result)
					}
				}
			}
		})
	}
}

func TestValidateCronExpression(t *testing.T) {
	tests := []struct {
		name       string
		expression string
		expectErr  bool
	}{
		{"Valid weekday expression", "0 9 * * MON-FRI", false},
		{"Valid daily expression", "0 0 * * *", false},
		{"Valid hourly expression", "0 * * * *", false},
		{"Invalid expression", "invalid", true},
		{"Too many fields", "0 0 0 * * * *", true},
		{"Too few fields", "0 0", true},
		{"Empty expression", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateCronExpression(tt.expression)

			if tt.expectErr {
				if err == nil {
					t.Errorf("Expected error for expression '%s' but got nil", tt.expression)
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error for expression '%s' but got: %v", tt.expression, err)
				}
			}
		})
	}
}

// Helper function to create string pointers
func stringPtr(s string) *string {
	return &s
}
