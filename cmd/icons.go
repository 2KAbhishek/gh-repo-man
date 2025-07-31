package cmd

import "strings"

var GeneralIcons = map[string]string{
	"archived": " ",
	"calendar": " ",
	"clock":    " ",
	"cloning":  " ",
	"disk":     " ",
	"done":     " ",
	"editor":   " ",
	"error":    " ",
	"fork":     " ",
	"forked":   " ",
	"home":     " ",
	"info":     " ",
	"issue":    " ",
	"link":     " ",
	"owner":    " ",
	"private":  " ",
	"star":     " ",
	"success":  " ",
	"tag":      " ",
	"template": " ",
	"watch":    " ",
}

var LanguageIcons = map[string]string{
	"c":          " ",
	"d":          " ",
	"r":          " ",
	"c++":        " ",
	"css":        " ",
	"dockerfile": " ",
	"go":         " ",
	"html":       " ",
	"java":       " ",
	"javascript": " ",
	"json":       " ",
	"kotlin":     " ",
	"lua":        " ",
	"markdown":   " ",
	"php":        " ",
	"python":     " ",
	"ruby":       " ",
	"rust":       " ",
	"shell":      " ",
	"swift":      " ",
	"typescript": " ",
	"yaml":       " ",
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
