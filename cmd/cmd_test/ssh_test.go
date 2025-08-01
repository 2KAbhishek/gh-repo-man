package cmd_test

import (
	"testing"

	"github.com/2KAbhishek/gh-repo-man/cmd"
)

func TestConvertToSSHURL(t *testing.T) {
	tests := []struct {
		name     string
		httpsURL string
		expected string
	}{
		{
			name:     "converts GitHub HTTPS URL to SSH",
			httpsURL: "https://github.com/owner/repo",
			expected: "git@github.com:owner/repo.git",
		},
		{
			name:     "handles URL that already has .git suffix",
			httpsURL: "https://github.com/owner/repo.git",
			expected: "git@github.com:owner/repo.git",
		},
		{
			name:     "returns non-GitHub URL unchanged",
			httpsURL: "https://gitlab.com/owner/repo",
			expected: "https://gitlab.com/owner/repo",
		},
		{
			name:     "returns SSH URL unchanged",
			httpsURL: "git@github.com:owner/repo.git",
			expected: "git@github.com:owner/repo.git",
		},
		{
			name:     "handles complex repository names",
			httpsURL: "https://github.com/2KAbhishek/gh-repo-man",
			expected: "git@github.com:2KAbhishek/gh-repo-man.git",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := cmd.ConvertToSSHURL(tt.httpsURL)
			if result != tt.expected {
				t.Errorf("ConvertToSSHURL() = %v, want %v", result, tt.expected)
			}
		})
	}
}
