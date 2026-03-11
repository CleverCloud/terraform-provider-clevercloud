package retry

import (
	"testing"
	"time"
)

func TestShouldRetry(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
		want       bool
	}{
		{"503 should retry", 503, true},
		{"200 should not retry", 200, false},
		{"400 should not retry", 400, false},
		{"401 should not retry", 401, false},
		{"404 should not retry", 404, false},
		{"500 should not retry", 500, false},
		{"502 should not retry", 502, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := shouldRetry(tt.statusCode); got != tt.want {
				t.Errorf("shouldRetry(%d) = %v, want %v", tt.statusCode, got, tt.want)
			}
		})
	}
}

func TestCalculateBackoff(t *testing.T) {
	config := Config{
		InitialDelay: 1 * time.Second,
		MaxDelay:     30 * time.Second,
		Multiplier:   2.0,
	}

	tests := []struct {
		attempt int
		want    time.Duration
	}{
		{0, 1 * time.Second},  // 1 * 2^0 = 1s
		{1, 2 * time.Second},  // 1 * 2^1 = 2s
		{2, 4 * time.Second},  // 1 * 2^2 = 4s
		{3, 8 * time.Second},  // 1 * 2^3 = 8s
		{4, 16 * time.Second}, // 1 * 2^4 = 16s
		{5, 30 * time.Second}, // 1 * 2^5 = 32s, capped at 30s
		{6, 30 * time.Second}, // 1 * 2^6 = 64s, capped at 30s
	}

	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			got := calculateBackoff(tt.attempt, config)
			if got != tt.want {
				t.Errorf("calculateBackoff(%d) = %v, want %v", tt.attempt, got, tt.want)
			}
		})
	}
}

func TestDefaultConfig(t *testing.T) {
	config := DefaultConfig()

	if config.MaxAttempts != 5 {
		t.Errorf("DefaultConfig().MaxAttempts = %d, want 5", config.MaxAttempts)
	}
	if config.InitialDelay != 1*time.Second {
		t.Errorf("DefaultConfig().InitialDelay = %v, want 1s", config.InitialDelay)
	}
	if config.MaxDelay != 30*time.Second {
		t.Errorf("DefaultConfig().MaxDelay = %v, want 30s", config.MaxDelay)
	}
	if config.Multiplier != 2.0 {
		t.Errorf("DefaultConfig().Multiplier = %f, want 2.0", config.Multiplier)
	}
}
