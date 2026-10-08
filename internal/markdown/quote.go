package markdown

import (
	"regexp"
	"strings"
)

var (
	calloutHeaderPattern = regexp.MustCompile(`^\[!(NOTE|TIP|IMPORTANT|WARNING|CAUTION)\]\s*$`)
	quoteLinePattern     = regexp.MustCompile(`^(>+)\s?(.*)$`)
)

func isQuoteLine(line string) bool {
	return quoteLinePattern.MatchString(strings.TrimSpace(line))
}

func parseQuoteBlock(block []string) string {
	if len(block) == 0 {
		return ""
	}

	// First, check if this is a Callout / Alert
	firstLineClean := strings.TrimSpace(quoteLinePattern.ReplaceAllString(strings.TrimSpace(block[0]), "$2"))
	if calloutMatch := calloutHeaderPattern.FindStringSubmatch(firstLineClean); calloutMatch != nil {
		alertType := strings.ToLower(calloutMatch[1])
		alertTitle := strings.ToUpper(alertType)

		bodyLines := make([]string, 0, len(block)-1)
		for _, line := range block[1:] {
			cleaned := quoteLinePattern.ReplaceAllString(strings.TrimSpace(line), "$2")
			bodyLines = append(bodyLines, strings.TrimSpace(cleaned))
		}

		bodyContent := renderInline(strings.Join(bodyLines, "<br />"))
		return `<div class="markdown-alert markdown-alert-` + alertType + `"><p class="markdown-alert-title">` + alertTitle + `</p><p>` + bodyContent + `</p></div>`
	}

	// Regular or nested blockquote
	var buf strings.Builder
	currentDepth := 0

	for _, line := range block {
		match := quoteLinePattern.FindStringSubmatch(strings.TrimSpace(line))
		if match == nil {
			continue
		}

		depth := len(match[1])
		text := strings.TrimSpace(match[2])

		for depth > currentDepth {
			buf.WriteString("<blockquote>")
			currentDepth++
		}
		for depth < currentDepth {
			buf.WriteString("</blockquote>")
			currentDepth--
		}

		if text != "" {
			buf.WriteString(renderInline(text))
			buf.WriteString("<br />")
		}
	}

	for currentDepth > 0 {
		buf.WriteString("</blockquote>")
		currentDepth--
	}

	// Clean up any trailing <br /> right before <blockquote> or </blockquote>
	result := buf.String()
	result = strings.ReplaceAll(result, "<br /><blockquote>", "<blockquote>")
	result = strings.ReplaceAll(result, "<br /></blockquote>", "</blockquote>")
	return result
}
