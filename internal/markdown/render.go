package markdown

import (
	"regexp"
	"strings"
	"unicode"
)

var (
	headingPattern    = regexp.MustCompile(`^(#{1,6})\s+(.*)$`)
	hrDashPattern     = regexp.MustCompile(`^---+$`)
	hrAsteriskPattern = regexp.MustCompile(`^\*\*\*+$`)
	quotePattern      = regexp.MustCompile(`^>\s?(.*)$`)
	unorderedPattern  = regexp.MustCompile(`^\s*[-*]\s+(.*)$`)
	orderedPattern    = regexp.MustCompile(`^\s*\d+\.\s+(.*)$`)
)

// Render converts supported markdown syntax to HTML.
func Render(markdown string) string {
	normalized := strings.ReplaceAll(markdown, "\r\n", "\n")
	normalized = strings.ReplaceAll(normalized, "\r", "\n")
	normalized = strings.TrimRightFunc(normalized, unicode.IsSpace)

	if strings.TrimSpace(normalized) == "" {
		return `<p class="empty">Start typing Markdown on the left.</p>`
	}

	lines := strings.Split(normalized, "\n")
	htmlBlocks := make([]string, 0, len(lines))
	paragraph := make([]string, 0, 8)
	listItems := make([]string, 0, 8)
	listType := ""
	codeBlock := make([]string, 0, 8)
	tableBlock := make([]string, 0, 8)
	inCodeBlock := false

	flushParagraph := func() {
		if len(paragraph) == 0 {
			return
		}

		htmlBlocks = append(htmlBlocks, `<p>`+renderInline(strings.Join(paragraph, " "))+`</p>`)
		paragraph = paragraph[:0]
	}

	flushList := func() {
		if len(listItems) == 0 {
			return
		}

		tag := "ul"
		if listType == "ol" {
			tag = "ol"
		}

		items := make([]string, 0, len(listItems))
		for _, item := range listItems {
			items = append(items, `<li>`+renderInline(item)+`</li>`)
		}

		htmlBlocks = append(htmlBlocks, `<`+tag+`>`+strings.Join(items, "")+`</`+tag+`>`)
		listItems = listItems[:0]
		listType = ""
	}

	flushCodeBlock := func() {
		if len(codeBlock) == 0 {
			return
		}

		htmlBlocks = append(htmlBlocks, `<pre><code>`+escapeHTML(strings.Join(codeBlock, "\n"))+`</code></pre>`)
		codeBlock = codeBlock[:0]
	}

	flushTable := func() {
		if len(tableBlock) == 0 {
			return
		}

		tableHTML := parseTable(tableBlock)
		if tableHTML != "" {
			htmlBlocks = append(htmlBlocks, tableHTML)
		} else {
			paragraph = append(paragraph, tableBlock...)
		}
		tableBlock = tableBlock[:0]
	}

	for _, line := range lines {
		if strings.HasPrefix(strings.TrimSpace(line), "```") {
			flushParagraph()
			flushList()
			flushTable()

			if inCodeBlock {
				flushCodeBlock()
			}

			inCodeBlock = !inCodeBlock
			continue
		}

		if inCodeBlock {
			codeBlock = append(codeBlock, line)
			continue
		}

		if strings.Contains(line, "|") && strings.TrimSpace(line) != "" {
			flushParagraph()
			flushList()
			tableBlock = append(tableBlock, line)
			continue
		}

		flushTable()

		if strings.TrimSpace(line) == "" {
			flushParagraph()
			flushList()
			continue
		}

		if heading := headingPattern.FindStringSubmatch(line); heading != nil {
			flushParagraph()
			flushList()
			level := len(heading[1])
			title := renderInline(strings.TrimSpace(heading[2]))
			htmlBlocks = append(htmlBlocks, `<h`+toString(level)+`>`+title+`</h`+toString(level)+`>`)
			continue
		}

		trimmed := strings.TrimSpace(line)
		if hrDashPattern.MatchString(trimmed) || hrAsteriskPattern.MatchString(trimmed) {
			flushParagraph()
			flushList()
			htmlBlocks = append(htmlBlocks, "<hr />")
			continue
		}

		if quote := quotePattern.FindStringSubmatch(line); quote != nil {
			flushParagraph()
			flushList()
			htmlBlocks = append(htmlBlocks, `<blockquote>`+renderInline(quote[1])+`</blockquote>`)
			continue
		}

		if unordered := unorderedPattern.FindStringSubmatch(line); unordered != nil {
			flushParagraph()
			if listType != "" && listType != "ul" {
				flushList()
			}
			listType = "ul"
			listItems = append(listItems, unordered[1])
			continue
		}

		if ordered := orderedPattern.FindStringSubmatch(line); ordered != nil {
			flushParagraph()
			if listType != "" && listType != "ol" {
				flushList()
			}
			listType = "ol"
			listItems = append(listItems, ordered[1])
			continue
		}

		paragraph = append(paragraph, strings.TrimSpace(line))
	}

	flushTable()
	flushParagraph()
	flushList()
	flushCodeBlock()

	return strings.Join(htmlBlocks, "")
}
