package cmd

var GeneralIcons = map[string]string{
	"star":     "⭐",
	"fork":     "🍴",
	"watch":    "👁",
	"issue":    "🐛",
	"owner":    "👤",
	"calendar": "📅",
	"clock":    "⏰",
	"disk":     "💾",
	"home":     "🏠",
	"tag":      "🏷",
	"link":     "🔗",
	"forked":   "🍴",
	"archived": "📦",
	"private":  "🔒",
	"template": "📋",
	"success":  "✅",
	"error":    "❌",
	"info":     "ℹ️",
	"cloning":  "📥",
	"done":     "✓",
}

func GetIcon(key string) string {
	if config.UI.Icons.General != nil {
		if icon, exists := config.UI.Icons.General[key]; exists && icon != "" {
			return icon
		}
	}

	if icon, exists := GeneralIcons[key]; exists {
		return icon
	}

	return "?"
}

// Language icons mapping
var LanguageIcons = map[string]string{
	"Go":         "🐹",
	"Python":     "🐍",
	"JavaScript": "📜",
	"TypeScript": "📘",
	"Java":       "☕",
	"C":          "⚙️",
	"C++":        "⚙️",
	"Ruby":       "💎",
	"PHP":        "🐘",
	"Rust":       "🦀",
	"Swift":      "🍎",
	"Kotlin":     "🎯",
	"Shell":      "🐚",
	"HTML":       "🌐",
	"CSS":        "🎨",
	"Lua":        "🌙",
	"Vim Script": "✏️",
	"Dockerfile": "🐳",
	"YAML":       "📄",
	"JSON":       "📋",
	"Markdown":   "📝",
}

// GetLanguageIcon returns the icon for a programming language
func GetLanguageIcon(language string) string {
	if config.UI.Icons.Languages != nil {
		if icon, exists := config.UI.Icons.Languages[language]; exists && icon != "" {
			return icon
		}
	}

	if icon, exists := LanguageIcons[language]; exists {
		return icon
	}

	return "📁"
}
