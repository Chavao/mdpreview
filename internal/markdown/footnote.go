package markdown

import (
	"regexp"
	"strings"
)

var (
	footnoteDefPattern = regexp.MustCompile(`^\[\^([a-zA-Z0-9_-]+)\]:\s+(.*)$`)
	footnoteRefPattern = regexp.MustCompile(`\[\^([a-zA-Z0-9_-]+)\]`)
)

type footnoteItem struct {
	id      string
	content string
}

func parseFootnoteDef(line string) *footnoteItem {
	m := footnoteDefPattern.FindStringSubmatch(strings.TrimSpace(line))
	if m == nil {
		return nil
	}
	return &footnoteItem{
		id:      m[1],
		content: m[2],
	}
}

func renderFootnotes(items []*footnoteItem) string {
	if len(items) == 0 {
		return ""
	}

	var buf strings.Builder
	buf.WriteString(`<section class="footnotes"><hr class="footnotes-sep" /><ol class="footnotes-list">`)

	for _, item := range items {
		contentHTML := renderInline(item.content)
		backref := ` <a href="#fnref:` + item.id + `" class="footnote-backref">↩</a>`
		buf.WriteString(`<li id="fn:` + item.id + `"><p>` + contentHTML + backref + `</p></li>`)
	}

	buf.WriteString(`</ol></section>`)
	return buf.String()
}
