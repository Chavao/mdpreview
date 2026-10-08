package markdown

import (
	"fmt"
	"html"
	"regexp"
	"strings"
)

var (
	inlineCodePattern    = regexp.MustCompile("`([^`]+)`")
	inlineMathPattern    = regexp.MustCompile(`\$([^\$\s](?:[^\$\n]*?[^\$\s])?)\$`)
	autolinkPattern      = regexp.MustCompile(`<(https?://[^>]+)>`)
	imageWithTitlePattern = regexp.MustCompile(`!\[([^\]]*)\]\((https?://[^\s)"]+|/[^\s)"]+|[^\s)"]+)\s+"([^"]*)"\)`)
	imageSimplePattern   = regexp.MustCompile(`!\[([^\]]*)\]\((https?://[^\s)"]+|/[^\s)"]+|[^\s)"]+)\)`)
	linkWithTitlePattern  = regexp.MustCompile(`\[([^\]]+)\]\((https?://[^\s)"]+|/[^\s)"]+|[^\s)"]+)\s+"([^"]*)"\)`)
	linkSimplePattern    = regexp.MustCompile(`\[([^\]]+)\]\((https?://[^\s)"]+|/[^\s)"]+|[^\s)"]+)\)`)
	strikethroughPattern = regexp.MustCompile(`~~([^~]+)~~`)
	boldItalicAstPattern = regexp.MustCompile(`\*\*\*([^*]+)\*\*\*`)
	boldItalicUndPattern = regexp.MustCompile(`___([^_]+)___`)
	boldAstPattern       = regexp.MustCompile(`\*\*([^*]+)\*\*`)
	boldUndPattern       = regexp.MustCompile(`__([^_]+)__`)
	italicAstPattern     = regexp.MustCompile(`\*([^*]+)\*`)
	italicUndPattern     = regexp.MustCompile(`(?:^|[^\w])_([^_]+)_(?:[^\w]|$)`)
)

func escapeHTML(value string) string {
	return sanitizeHTML(value)
}

func renderInline(text string) string {
	// 1. Protect code spans from any formatting or tag manipulation.
	codePlaceholders := make([]string, 0, 4)
	text = inlineCodePattern.ReplaceAllStringFunc(text, func(match string) string {
		sub := inlineCodePattern.FindStringSubmatch(match)
		idx := len(codePlaceholders)
		codePlaceholders = append(codePlaceholders, `<code>`+html.EscapeString(sub[1])+`</code>`)
		return fmt.Sprintf("\x00INLINE_CODE_%d\x00", idx)
	})

	// 2. Protect inline math formulas from formatting.
	mathPlaceholders := make([]string, 0, 4)
	text = inlineMathPattern.ReplaceAllStringFunc(text, func(match string) string {
		sub := inlineMathPattern.FindStringSubmatch(match)
		idx := len(mathPlaceholders)
		mathPlaceholders = append(mathPlaceholders, `<span class="math-inline"><code>`+html.EscapeString(sub[1])+`</code></span>`)
		return fmt.Sprintf("\x00INLINE_MATH_%d\x00", idx)
	})

	// 3. Autolinks: <https://...>
	text = autolinkPattern.ReplaceAllString(text, `<a href="$1" target="_blank" rel="noreferrer">$1</a>`)

	// 4. Images (must precede link processing).
	text = imageWithTitlePattern.ReplaceAllString(text, `<img src="$2" alt="$1" title="$3" />`)
	text = imageSimplePattern.ReplaceAllString(text, `<img src="$2" alt="$1" />`)

	// 5. Footnotes reference: [^1]
	text = footnoteRefPattern.ReplaceAllString(text, `<sup class="footnote-ref" id="fnref:$1"><a href="#fn:$1">[$1]</a></sup>`)

	// 6. Links with or without titles.
	text = linkWithTitlePattern.ReplaceAllString(text, `<a href="$2" target="_blank" rel="noreferrer" title="$3">$1</a>`)
	text = linkSimplePattern.ReplaceAllString(text, `<a href="$2" target="_blank" rel="noreferrer">$1</a>`)

	// 6. Sanitize HTML, preserving safe tags/attributes and escaping dangerous ones.
	safe := sanitizeHTML(text)

	// 7. Strikethrough.
	safe = strikethroughPattern.ReplaceAllString(safe, `<del>$1</del>`)

	// 8. Emphasis (Bold and Italic combinations).
	safe = boldItalicAstPattern.ReplaceAllString(safe, `<strong><em>$1</em></strong>`)
	safe = boldItalicUndPattern.ReplaceAllString(safe, `<strong><em>$1</em></strong>`)
	safe = boldAstPattern.ReplaceAllString(safe, `<strong>$1</strong>`)
	safe = boldUndPattern.ReplaceAllString(safe, `<strong>$1</strong>`)
	safe = italicAstPattern.ReplaceAllString(safe, `<em>$1</em>`)

	// Single underscore italic matching word boundaries.
	safe = italicUndPattern.ReplaceAllStringFunc(safe, func(m string) string {
		prefix := ""
		suffix := ""
		trimmed := m
		if len(m) > 0 && m[0] != '_' {
			prefix = string(m[0])
			trimmed = m[1:]
		}
		if len(trimmed) > 0 && trimmed[len(trimmed)-1] != '_' {
			suffix = string(trimmed[len(trimmed)-1])
			trimmed = trimmed[:len(trimmed)-1]
		}
		if strings.HasPrefix(trimmed, "_") && strings.HasSuffix(trimmed, "_") && len(trimmed) > 2 {
			inner := trimmed[1 : len(trimmed)-1]
			return prefix + `<em>` + inner + `</em>` + suffix
		}
		return m
	})

	// 9. Restore math placeholders.
	for i, mathSnippet := range mathPlaceholders {
		placeholder := fmt.Sprintf("\x00INLINE_MATH_%d\x00", i)
		safe = strings.ReplaceAll(safe, placeholder, mathSnippet)
	}

	// 10. Restore code placeholders.
	for i, codeSnippet := range codePlaceholders {
		placeholder := fmt.Sprintf("\x00INLINE_CODE_%d\x00", i)
		safe = strings.ReplaceAll(safe, placeholder, codeSnippet)
	}

	return safe
}
