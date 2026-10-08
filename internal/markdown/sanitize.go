package markdown

import (
	"html"
	"regexp"
	"strings"
)

var (
	htmlTagPattern = regexp.MustCompile(`(?i)</?([a-zA-Z][a-zA-Z0-9]*)\b([^>]*)>`)
	attrPattern    = regexp.MustCompile(`(?i)([a-zA-Z][a-zA-Z0-9_-]*)(?:\s*=\s*(?:"([^"]*)"|'([^']*)'|([^\s"'=<>` + "`" + `]+)))?`)

	allowedTags = map[string]bool{
		"a":      true,
		"b":      true,
		"blockquote": true,
		"br":     true,
		"code":   true,
		"del":    true,
		"em":     true,
		"i":      true,
		"img":    true,
		"kbd":    true,
		"mark":   true,
		"s":      true,
		"span":   true,
		"strong": true,
		"sub":    true,
		"sup":    true,
	}

	allowedAttributes = map[string]bool{
		"alt":    true,
		"class":  true,
		"height": true,
		"href":   true,
		"id":     true,
		"rel":    true,
		"src":    true,
		"style":  true,
		"target": true,
		"title":  true,
		"width":  true,
	}
)

// sanitizeHTML preserves whitelisted safe HTML tags and attributes while escaping
// dangerous tags (e.g. script, iframe) and dangerous attributes/schemes (e.g. onclick, javascript:).
func sanitizeHTML(input string) string {
	var result strings.Builder
	lastIndex := 0

	matches := htmlTagPattern.FindAllStringSubmatchIndex(input, -1)
	for _, m := range matches {
		matchStart := m[0]
		matchEnd := m[1]
		tagStart := m[2]
		tagEnd := m[3]
		attrStart := m[4]
		attrEnd := m[5]

		// Escape raw text preceding this tag.
		result.WriteString(html.EscapeString(input[lastIndex:matchStart]))
		lastIndex = matchEnd

		fullTag := input[matchStart:matchEnd]
		tagName := strings.ToLower(input[tagStart:tagEnd])
		rawAttrs := input[attrStart:attrEnd]
		isClosing := strings.HasPrefix(fullTag, "</")
		isSelfClosing := strings.HasSuffix(strings.TrimSpace(rawAttrs), "/")

		if !allowedTags[tagName] {
			// Disallowed tag: escape it completely.
			result.WriteString(html.EscapeString(fullTag))
			continue
		}

		if isClosing {
			result.WriteString("</" + tagName + ">")
			continue
		}

		// Parse and filter attributes for allowed tag.
		safeAttrs := filterAttributes(rawAttrs)
		result.WriteString("<" + tagName)
		if safeAttrs != "" {
			result.WriteString(" " + safeAttrs)
		}
		if isSelfClosing {
			result.WriteString(" />")
		} else {
			result.WriteString(">")
		}
	}

	// Escape any remaining trailing text.
	result.WriteString(html.EscapeString(input[lastIndex:]))
	return result.String()
}

func filterAttributes(rawAttrs string) string {
	rawAttrs = strings.TrimSuffix(strings.TrimSpace(rawAttrs), "/")
	if strings.TrimSpace(rawAttrs) == "" {
		return ""
	}

	attrMatches := attrPattern.FindAllStringSubmatch(rawAttrs, -1)
	var safe []string

	for _, match := range attrMatches {
		attrName := strings.ToLower(match[1])

		// Block any event handler (e.g. onclick, onerror, onload).
		if strings.HasPrefix(attrName, "on") {
			continue
		}

		if !allowedAttributes[attrName] {
			continue
		}

		attrVal := ""
		if match[2] != "" {
			attrVal = match[2]
		} else if match[3] != "" {
			attrVal = match[3]
		} else if match[4] != "" {
			attrVal = match[4]
		}

		valLower := strings.ToLower(strings.TrimSpace(attrVal))

		// Disallow dangerous URI schemes in href or src.
		if (attrName == "href" || attrName == "src") && (strings.HasPrefix(valLower, "javascript:") || strings.HasPrefix(valLower, "data:text/html")) {
			continue
		}

		// Disallow dangerous script execution in style.
		if attrName == "style" && (strings.Contains(valLower, "javascript:") || strings.Contains(valLower, "expression(") || strings.Contains(valLower, "behavior:")) {
			continue
		}

		safe = append(safe, attrName+`="`+html.EscapeString(attrVal)+`"`)
	}

	return strings.Join(safe, " ")
}
