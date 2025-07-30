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
	"editor":   "📝",
}

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
	if icon, exists := config.UI.Icons.Languages[language]; exists && icon != "" {
		return icon
	}

	if icon, exists := LanguageIcons[language]; exists {
		return icon
	}

	return "📁"
}
