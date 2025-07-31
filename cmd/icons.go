package cmd

import "strings"

var GeneralIcons = map[string]string{
	"star":     " ",
	"fork":     " ",
	"watch":    " ",
	"issue":    " ",
	"owner":    " ",
	"calendar": " ",
	"clock":    " ",
	"disk":     " ",
	"home":     " ",
	"tag":      " ",
	"link":     " ",
	"forked":   " ",
	"archived": " ",
	"private":  " ",
	"template": " ",
	"success":  " ",
	"error":    " ",
	"info":     " ",
	"cloning":  " ",
	"done":     " ",
	"editor":   " ",
}

var LanguageIcons = map[string]string{
	"go":         " ",
	"python":     " ",
	"javascript": " ",
	"typescript": " ",
	"java":       " ",
	"c":          " ",
	"c++":        " ",
	"ruby":       " ",
	"php":        " ",
	"rust":       " ",
	"swift":      " ",
	"kotlin":     " ",
	"shell":      " ",
	"html":       " ",
	"css":        " ",
	"lua":        " ",
	"dockerfile": " ",
	"yaml":       " ",
	"json":       " ",
	"markdown":   " ",
}

// GetIcon returns the icon for a given key
func GetIcon(key string) string {
	if icon, exists := config.UI.Icons.General[key]; exists && icon != "" {
		return icon
	}

	if icon, exists := GeneralIcons[key]; exists {
		return icon
	}

	return "?"
}

// GetLanguageIcon returns the icon for a programming language
func GetLanguageIcon(language string) string {
	normalizedLang := strings.ToLower(language)
	if icon, exists := config.UI.Icons.Languages[normalizedLang]; exists && icon != "" {
		return icon
	}

	if icon, exists := LanguageIcons[normalizedLang]; exists {
		return icon
	}

	return LanguageIcons["markdown"]
}
