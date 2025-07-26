package cmd

// Unicode icons and symbols used throughout the application
const (
	// Repository info icons
	IconStar     = "⭐"
	IconFork     = "🍴"
	IconWatch    = "👁"
	IconIssue    = "🐛"
	IconOwner    = "👤"
	IconCalendar = "📅"
	IconClock    = "⏰"
	IconDisk     = "💾"
	IconHome     = "🏠"
	IconTag      = "🏷"
	IconLink     = "🔗"

	// Status indicators
	IconForked   = "🍴"
	IconArchived = "📦"
	IconPrivate  = "🔒"
	IconTemplate = "📋"
	IconSuccess  = "✅"
	IconError    = "❌"
	IconInfo     = "ℹ️"

	// Progress indicators
	IconCloning = "📥"
	IconDone    = "✓"
)

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

// GetLanguageIcon returns the appropriate icon for a programming language
func GetLanguageIcon(language string) string {
	if icon, exists := LanguageIcons[language]; exists {
		return icon
	}
	return "📁" // default folder icon
}
