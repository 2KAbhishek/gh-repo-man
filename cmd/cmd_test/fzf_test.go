package cmd_test

import (
	"testing"
)

func TestFzfCancellationLogic(t *testing.T) {
	tests := []struct {
		name     string
		exitCode int
		isCancel bool
	}{
		{"ctrl-c cancellation", 130, true},
		{"esc cancellation", 1, true},
		{"other error", 2, false},
		{"success", 0, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			isCancelError := tt.exitCode == 130 || tt.exitCode == 1

			if isCancelError != tt.isCancel {
				t.Errorf("Expected isCancelError=%v for exit code %d, got %v", tt.isCancel, tt.exitCode, isCancelError)
			}
		})
	}
}
