package markdown

import (
	"html"
	"regexp"
	"strings"
	"unicode"
)

var (
	headingPattern    = regexp.MustCompile(`^(#{1,6})\s+(.*)$`)
	hrDashPattern     = regexp.MustCompile(`^---+$`)
	hrAsteriskPattern = regexp.MustCompile(`^\*\*\*+$`)
)

// Render converts supported markdown syntax to HTML according to the 2026 specification.
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
	listBlock := make([]string, 0, 8)
	quoteBlock := make([]string, 0, 8)
	codeBlock := make([]string, 0, 8)
	mathBlock := make([]string, 0, 8)
	tableBlock := make([]string, 0, 8)
	footnoteDefs := make([]*footnoteItem, 0, 4)

	inCodeBlock := false
	codeBlockLang := ""
	codeFence := ""

	inMathBlock := false

	flushParagraph := func() {
		if len(paragraph) == 0 {
			return
		}

		content := strings.Join(paragraph, " ")
		content = strings.ReplaceAll(content, "<br /> ", "<br />")
		htmlBlocks = append(htmlBlocks, `<p>`+renderInline(content)+`</p>`)
		paragraph = paragraph[:0]
	}

	flushList := func() {
		if len(listBlock) == 0 {
			return
		}

		htmlBlocks = append(htmlBlocks, parseList(listBlock))
		listBlock = listBlock[:0]
	}

	flushQuote := func() {
		if len(quoteBlock) == 0 {
			return
		}

		htmlBlocks = append(htmlBlocks, parseQuoteBlock(quoteBlock))
		quoteBlock = quoteBlock[:0]
	}

	flushCodeBlock := func() {
		if len(codeBlock) == 0 && codeBlockLang == "" {
			return
		}

		code := html.EscapeString(strings.Join(codeBlock, "\n"))
		switch {
		case codeBlockLang == "mermaid":
			htmlBlocks = append(htmlBlocks, `<div class="mermaid-block"><div class="mermaid-label">Diagram (Mermaid)</div><pre><code class="language-mermaid">`+code+`</code></pre></div>`)
		case codeBlockLang != "":
			htmlBlocks = append(htmlBlocks, `<pre><code class="language-`+codeBlockLang+`">`+code+`</code></pre>`)
		default:
			htmlBlocks = append(htmlBlocks, `<pre><code>`+code+`</code></pre>`)
		}

		codeBlock = codeBlock[:0]
		codeBlockLang = ""
		codeFence = ""
	}

	flushMathBlock := func() {
		if len(mathBlock) == 0 {
			return
		}

		mathCode := html.EscapeString(strings.Join(mathBlock, "\n"))
		htmlBlocks = append(htmlBlocks, `<div class="math-block"><div class="math-label">Math (LaTeX)</div><pre class="math-display"><code>`+mathCode+`</code></pre></div>`)
		mathBlock = mathBlock[:0]
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

	flushAllNonCode := func() {
		flushParagraph()
		flushList()
		flushQuote()
		flushTable()
	}

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		// 1. Fenced Code Blocks (``` or ~~~)
		if inCodeBlock {
			if strings.HasPrefix(trimmed, codeFence) && len(trimmed) >= 3 {
				inCodeBlock = false
				flushCodeBlock()
				continue
			}
			codeBlock = append(codeBlock, line)
			continue
		}

		if (strings.HasPrefix(trimmed, "```") || strings.HasPrefix(trimmed, "~~~")) && !inMathBlock {
			flushAllNonCode()
			inCodeBlock = true
			if strings.HasPrefix(trimmed, "```") {
				codeFence = "```"
				codeBlockLang = strings.TrimSpace(strings.TrimPrefix(trimmed, "```"))
			} else {
				codeFence = "~~~"
				codeBlockLang = strings.TrimSpace(strings.TrimPrefix(trimmed, "~~~"))
			}
			continue
		}

		// 2. Display Math Blocks ($$)
		if inMathBlock {
			if trimmed == "$$" {
				inMathBlock = false
				flushMathBlock()
				continue
			}
			mathBlock = append(mathBlock, line)
			continue
		}

		if trimmed == "$$" {
			flushAllNonCode()
			inMathBlock = true
			continue
		}

		// 3. Tables
		if strings.Contains(line, "|") && trimmed != "" {
			flushParagraph()
			flushList()
			flushQuote()
			tableBlock = append(tableBlock, line)
			continue
		}
		flushTable()

		// 4. Blank lines
		if trimmed == "" {
			flushParagraph()
			flushList()
			flushQuote()
			continue
		}

		// 5. Footnote definition
		if fn := parseFootnoteDef(line); fn != nil {
			flushParagraph()
			flushList()
			flushQuote()
			footnoteDefs = append(footnoteDefs, fn)
			continue
		}

		// 6. Blockquotes and Callouts
		if isQuoteLine(line) {
			flushParagraph()
			flushList()
			quoteBlock = append(quoteBlock, line)
			continue
		}
		flushQuote()

		// 7. Headings
		if heading := headingPattern.FindStringSubmatch(line); heading != nil {
			flushParagraph()
			flushList()
			level := len(heading[1])
			title := renderInline(strings.TrimSpace(heading[2]))
			htmlBlocks = append(htmlBlocks, `<h`+toString(level)+`>`+title+`</h`+toString(level)+`>`)
			continue
		}

		// 8. Horizontal rules
		if hrDashPattern.MatchString(trimmed) || hrAsteriskPattern.MatchString(trimmed) {
			flushParagraph()
			flushList()
			htmlBlocks = append(htmlBlocks, "<hr />")
			continue
		}

		// 9. Lists (ordered, unordered, task lists)
		if isListLine(line) {
			flushParagraph()
			listBlock = append(listBlock, line)
			continue
		}
		flushList()

		// 10. Paragraph lines (detect trailing spaces for soft line breaks)
		hasLineBreak := len(line)-len(strings.TrimRight(line, " \t")) >= 2
		pLine := strings.TrimSpace(line)
		if hasLineBreak {
			pLine += "<br />"
		}
		paragraph = append(paragraph, pLine)
	}

	// Flush any pending blocks at end of document
	if inCodeBlock {
		flushCodeBlock()
	}
	if inMathBlock {
		flushMathBlock()
	}
	flushTable()
	flushParagraph()
	flushList()
	flushQuote()

	// Append footnotes section if footnotes were defined
	if len(footnoteDefs) > 0 {
		htmlBlocks = append(htmlBlocks, renderFootnotes(footnoteDefs))
	}

	return strings.Join(htmlBlocks, "")
}
