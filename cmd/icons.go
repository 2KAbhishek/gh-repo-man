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
	"astro":      " ",
	"ada":        " ",
	"asm":        " ",
	"babel":      " ",
	"batchfile":  " ",
	"bazel":      " ",
	"c":          " ",
	"c++":        " ",
	"c-sharp":    " ",
	"cake":       " ",
	"cake_php":   " ",
	"clojure":    " ",
	"cobol":      " ",
	"coffee":     " ",
	"coldfusion": " ",
	"crystal":    " ",
	"css":        " ",
	"cuda":       " ",
	"d":          " ",
	"dart":       " ",
	"dockerfile": " ",
	"elixir":     " ",
	"elm":        " ",
	"emacs lisp": " ",
	"erlang":     " ",
	"f-sharp":    " ",
	"fennel":     " ",
	"fortran":    " ",
	"go":         " ",
	"godot":      " ",
	"gradle":     " ",
	"groovy":     " ",
	"grunt":      " ",
	"hacklang":   " ",
	"haml":       " ",
	"haskell":    " ",
	"haxe":       " ",
	"html":       " ",
	"java":       " ",
	"javascript": " ",
	"jinja":      " ",
	"json":       " ",
	"julia":      " ",
	"karma":      " ",
	"kotlin":     " ",
	"less":       " ",
	"lisp":       " ",
	"livescript": " ",
	"lua":        " ",
	"makefile":   " ",
	"markdown":   " ",
	"nim":        " ",
	"nunjucks":   " ",
	"ocaml":      " ",
	"perl":       " ",
	"php":        " ",
	"powershell": " ",
	"prolog":     " ",
	"pug":        " ",
	"puppet":     " ",
	"purescript": " ",
	"python":     " ",
	"r":          " ",
	"rails":      " ",
	"react":      " ",
	"reasonml":   " ",
	"rescript":   " ",
	"rollup":     " ",
	"ruby":       " ",
	"rust":       " ",
	"sass":       " ",
	"sbt":        " ",
	"scala":      " ",
	"scheme":     " ",
	"shell":      " ",
	"slim":       " ",
	"spring":     " ",
	"stylus":     " ",
	"svelte":     " ",
	"swift":      " ",
	"tex":        " ",
	"toml":       " ",
	"twig":       " ",
	"typescript": " ",
	"v":          " ",
	"vala":       " ",
	"vim script": " ",
	"vue":        " ",
	"wasm":       " ",
	"webpack":    " ",
	"yaml":       " ",
	"zig":        " ",
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
