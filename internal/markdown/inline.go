package markdown

import (
	"html"
	"regexp"
	"strconv"
	"strings"
)

var (
	codePattern     = regexp.MustCompile("`([^`]+)`")
	boldPattern     = regexp.MustCompile(`\*\*([^*]+)\*\*`)
	emphasisPattern = regexp.MustCompile(`\*([^*]+)\*`)
	linkPattern     = regexp.MustCompile(`\[([^\]]+)\]\((https?://[^)\s]+)\)`)
	tableSeparator  = regexp.MustCompile(`^\s*\|?(?:\s*:?-{3,}:?\s*\|)+\s*:?-{3,}:?\s*\|?\s*$`)
)

func escapeHTML(value string) string {
	return html.EscapeString(value)
}

func renderInline(text string) string {
	safe := escapeHTML(text)
	safe = codePattern.ReplaceAllString(safe, `<code>$1</code>`)
	safe = boldPattern.ReplaceAllString(safe, `<strong>$1</strong>`)
	safe = emphasisPattern.ReplaceAllString(safe, `<em>$1</em>`)
	safe = linkPattern.ReplaceAllString(safe, `<a href="$2" target="_blank" rel="noreferrer">$1</a>`)
	return safe
}

func parseTable(block []string) string {
	if len(block) < 2 {
		return ""
	}

	if !tableSeparator.MatchString(block[1]) {
		return ""
	}

	rows := make([][]string, 0, len(block))
	for _, line := range block {
		parts := strings.Split(strings.Trim(strings.TrimSpace(line), "|"), "|")
		cells := make([]string, 0, len(parts))
		for _, part := range parts {
			cells = append(cells, renderInline(strings.TrimSpace(part)))
		}
		rows = append(rows, cells)
	}

	headers := rows[0]
	bodyRows := rows[2:]

	headerCells := make([]string, 0, len(headers))
	for _, cell := range headers {
		headerCells = append(headerCells, `<th>`+cell+`</th>`)
	}

	body := make([]string, 0, len(bodyRows))
	for _, row := range bodyRows {
		cells := make([]string, 0, len(row))
		for _, cell := range row {
			cells = append(cells, `<td>`+cell+`</td>`)
		}
		body = append(body, `<tr>`+strings.Join(cells, "")+`</tr>`)
	}

	return `<table><thead><tr>` + strings.Join(headerCells, "") + `</tr></thead><tbody>` + strings.Join(body, "") + `</tbody></table>`
}

func toString(value int) string {
	return strconv.Itoa(value)
}
