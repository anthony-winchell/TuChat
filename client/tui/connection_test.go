package tui

import (
	"testing"
	"time"
)

func TestReconnectDelay(t *testing.T) {
	tests := []struct {
		attempt int
		want    time.Duration
	}{
		{1, 1 * time.Second},
		{2, 2 * time.Second},
		{3, 4 * time.Second},
		{4, 8 * time.Second},
		{5, 16 * time.Second},
		{6, 30 * time.Second},
		{7, 30 * time.Second},
		{10, 30 * time.Second},
	}

	for _, tt := range tests {
		if got := reconnectDelay(tt.attempt); got != tt.want {
			t.Fatalf("reconnectDelay(%d) = %v, want %v", tt.attempt, got, tt.want)
		}
	}
}
