package utils

import (
	"fmt"
	"strings"
)

// BuildFolderTree renders a list of file paths into an ASCII tree string.
func BuildFolderTree(paths []string) string {
	if len(paths) == 0 {
		return ""
	}

	var sb strings.Builder
	sb.WriteString("```\n")
	for _, p := range paths {
		depth := strings.Count(p, "/")
		name := p
		if idx := strings.LastIndex(p, "/"); idx != -1 {
			name = p[idx+1:]
		}
		indent := strings.Repeat("  ", depth)
		sb.WriteString(fmt.Sprintf("%s├── %s\n", indent, name))
	}
	sb.WriteString("```")
	return sb.String()
}

// FormatLanguages formats the languages percentage map into a readable string.
func FormatLanguages(langs map[string]float64) string {
	if len(langs) == 0 {
		return "Not detected"
	}
	parts := make([]string, 0, len(langs))
	for lang, pct := range langs {
		parts = append(parts, fmt.Sprintf("%s (%.1f%%)", lang, pct))
	}
	return strings.Join(parts, ", ")
}

// TruncateString trims a string to maxLen runes, appending "..." if needed.
func TruncateString(s string, maxLen int) string {
	runes := []rune(s)
	if len(runes) <= maxLen {
		return s
	}
	return string(runes[:maxLen-3]) + "..."
}

// JoinNonEmpty joins only non-empty strings with a separator.
func JoinNonEmpty(sep string, parts ...string) string {
	var filtered []string
	for _, p := range parts {
		if strings.TrimSpace(p) != "" {
			filtered = append(filtered, p)
		}
	}
	return strings.Join(filtered, sep)
}
