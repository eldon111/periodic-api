package integration

import (
	"testing"
	"time"

	"periodic-api/internal/utils"
)

// TestCronCalculationDebug helps debug cron calculation issues
func TestCronCalculationDebug(t *testing.T) {
	now := time.Now()
	pastTime := now.Add(-time.Hour) // 1 hour ago
	cronExpr := "0 * * * *"         // Every hour at minute 0
	futureExpiration := now.Add(24 * time.Hour)

	t.Logf("Now: %v", now)
	t.Logf("Past time (starts at): %v", pastTime)
	t.Logf("Cron expression: %s", cronExpr)
	t.Logf("Expiration: %v", futureExpiration)

	// Test what CalculateNextExecution returns
	nextExec := utils.CalculateNextExecution(pastTime, true, &cronExpr, &futureExpiration)
	
	if nextExec == nil {
		t.Error("CalculateNextExecution returned nil - this might be the issue!")
		return
	}

	t.Logf("Calculated next execution: %v", *nextExec)
	t.Logf("Is next execution in the future? %v", nextExec.After(now))

	// Check if the next execution is reasonable
	if !nextExec.After(now) {
		t.Error("Next execution should be in the future")
	}

	// For "0 * * * *" (every hour at minute 0), the next execution should be at the next hour boundary
	expectedNextHour := now.Truncate(time.Hour).Add(time.Hour)
	t.Logf("Expected next hour boundary: %v", expectedNextHour)
	
	// The calculation might return the exact hour boundary or slightly different
	timeDiff := nextExec.Sub(expectedNextHour)
	if timeDiff < -time.Minute || timeDiff > time.Minute {
		t.Logf("Warning: Next execution time differs from expected by %v", timeDiff)
	}
}