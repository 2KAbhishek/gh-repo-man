package cmd_test

import (
	"testing"

	"github.com/2KAbhishek/gh-repo-man/cmd"
)

func TestGetIcon(t *testing.T) {
	tests := []struct {
		name     string
		key      string
		expected string
	}{
		{
			name:     "existing general icon - star",
			key:      "star",
			expected: "\uf41e ",
		},
		{
			name:     "existing general icon - fork",
			key:      "fork",
			expected: "\uf402 ",
		},
		{
			name:     "existing general icon - info",
			key:      "info",
			expected: "\uf449 ",
		},
		{
			name:     "existing general icon - link",
			key:      "link",
			expected: "\uf465 ",
		},
		{
			name:     "existing general icon - calendar",
			key:      "calendar",
			expected: "\uf455 ",
		},
		{
			name:     "existing general icon - clock",
			key:      "clock",
			expected: "\uf43a ",
		},
		{
			name:     "existing general icon - disk",
			key:      "disk",
			expected: "\uf473 ",
		},
		{
			name:     "existing general icon - home",
			key:      "home",
			expected: "\uf46d ",
		},
		{
			name:     "existing general icon - tag",
			key:      "tag",
			expected: "\uf412 ",
		},
		{
			name:     "existing general icon - owner",
			key:      "owner",
			expected: "\uf415 ",
		},
		{
			name:     "existing general icon - watch",
			key:      "watch",
			expected: "\uf441 ",
		},
		{
			name:     "existing general icon - issue",
			key:      "issue",
			expected: "\uf41b ",
		},
		{
			name:     "non-existing icon",
			key:      "nonexistent",
			expected: "?",
		},
		{
			name:     "empty key",
			key:      "",
			expected: "?",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := cmd.GetIcon(tt.key)
			if result != tt.expected {
				t.Errorf("GetIcon(%q) = %q, want %q", tt.key, result, tt.expected)
			}
		})
	}
}

func TestGetLanguageIcon(t *testing.T) {
	tests := []struct {
		name     string
		language string
		expected string
	}{
		{
			name:     "Go language",
			language: "Go",
			expected: "\ue627 ",
		},
		{
			name:     "go lowercase",
			language: "go",
			expected: "\ue627 ",
		},
		{
			name:     "gO mixed case",
			language: "gO",
			expected: "\ue627 ",
		},
		{
			name:     "Python language",
			language: "Python",
			expected: "\ue606 ",
		},
		{
			name:     "python lowercase",
			language: "python",
			expected: "\ue606 ",
		},
		{
			name:     "JavaScript language",
			language: "JavaScript",
			expected: "\ue60c ",
		},
		{
			name:     "javascript lowercase",
			language: "javascript",
			expected: "\ue60c ",
		},
		{
			name:     "TypeScript language",
			language: "TypeScript",
			expected: "\ue628 ",
		},
		{
			name:     "typescript lowercase",
			language: "typescript",
			expected: "\ue628 ",
		},
		{
			name:     "Rust language",
			language: "Rust",
			expected: "\ue68b ",
		},
		{
			name:     "rust lowercase",
			language: "rust",
			expected: "\ue68b ",
		},
		{
			name:     "Java language",
			language: "Java",
			expected: "\ue738 ",
		},
		{
			name:     "C language",
			language: "C",
			expected: "\ue61e ",
		},
		{
			name:     "C++ language",
			language: "C++",
			expected: "\ue646 ",
		},
		{
			name:     "Shell language",
			language: "Shell",
			expected: "\ue760 ",
		},
		{
			name:     "HTML language",
			language: "HTML",
			expected: "\ue60e ",
		},
		{
			name:     "CSS language",
			language: "CSS",
			expected: "\ue749 ",
		},
		{
			name:     "unknown language defaults to markdown",
			language: "lolcat",
			expected: "\ue609 ",
		},
		{
			name:     "empty language defaults to markdown",
			language: "",
			expected: "\ue609 ",
		},
		{
			name:     "case insensitive test",
			language: "PYTHON",
			expected: "\ue606 ",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := cmd.GetLanguageIcon(tt.language)
			if result != tt.expected {
				t.Errorf("GetLanguageIcon(%q) = %q, want %q", tt.language, result, tt.expected)
			}
		})
	}
}

func TestIconMapsExist(t *testing.T) {
	if len(cmd.GeneralIcons) == 0 {
		t.Error("GeneralIcons map should not be empty")
	}

	if len(cmd.LanguageIcons) == 0 {
		t.Error("LanguageIcons map should not be empty")
	}

	requiredGeneralIcons := []string{"star", "fork", "info", "link", "calendar", "clock", "disk", "home", "tag", "owner", "watch", "issue"}
	for _, icon := range requiredGeneralIcons {
		if _, exists := cmd.GeneralIcons[icon]; !exists {
			t.Errorf("GeneralIcons should contain '%s'", icon)
		}
	}

	requiredLanguageIcons := []string{"go", "python", "javascript", "typescript", "rust", "java", "c", "c++", "shell", "html", "css", "markdown"}
	for _, lang := range requiredLanguageIcons {
		if _, exists := cmd.LanguageIcons[lang]; !exists {
			t.Errorf("LanguageIcons should contain '%s'", lang)
		}
	}
}
