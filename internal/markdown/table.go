package markdown

import (
	"regexp"
	"strconv"
	"strings"
)

var tableSeparator = regexp.MustCompile(`^\s*\|?(?:\s*:?-{3,}:?\s*\|)+\s*:?-{3,}:?\s*\|?\s*$`)

type columnAlign string

const (
	alignLeft   columnAlign = "left"
	alignCenter columnAlign = "center"
	alignRight  columnAlign = "right"
	alignNone   columnAlign = ""
)

func parseTable(block []string) string {
	if len(block) < 2 {
		return ""
	}

	if !tableSeparator.MatchString(block[1]) {
		return ""
	}

	alignments := parseTableAlignments(block[1])

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
	for i, cell := range headers {
		styleAttr := ""
		if i < len(alignments) && alignments[i] != alignNone {
			styleAttr = ` style="text-align: ` + string(alignments[i]) + `;"`
		}
		headerCells = append(headerCells, `<th`+styleAttr+`>`+cell+`</th>`)
	}

	body := make([]string, 0, len(bodyRows))
	for _, row := range bodyRows {
		cells := make([]string, 0, len(row))
		for i, cell := range row {
			styleAttr := ""
			if i < len(alignments) && alignments[i] != alignNone {
				styleAttr = ` style="text-align: ` + string(alignments[i]) + `;"`
			}
			cells = append(cells, `<td`+styleAttr+`>`+cell+`</td>`)
		}
		body = append(body, `<tr>`+strings.Join(cells, "")+`</tr>`)
	}

	return `<table><thead><tr>` + strings.Join(headerCells, "") + `</tr></thead><tbody>` + strings.Join(body, "") + `</tbody></table>`
}

func parseTableAlignments(separatorLine string) []columnAlign {
	parts := strings.Split(strings.Trim(strings.TrimSpace(separatorLine), "|"), "|")
	aligns := make([]columnAlign, 0, len(parts))

	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		hasLeft := strings.HasPrefix(trimmed, ":")
		hasRight := strings.HasSuffix(trimmed, ":")

		switch {
		case hasLeft && hasRight:
			aligns = append(aligns, alignCenter)
		case hasRight:
			aligns = append(aligns, alignRight)
		case hasLeft:
			aligns = append(aligns, alignLeft)
		default:
			aligns = append(aligns, alignNone)
		}
	}

	return aligns
}

func toString(value int) string {
	return strconv.Itoa(value)
}
