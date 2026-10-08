package markdown_test

import (
	"strings"
	"testing"

	"github.com/Chavao/mdpreview/internal/markdown"
)

func TestRenderBlockElements(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		input    string
		contains []string
	}{
		{
			name:  "fenced code block with language and tilde",
			input: "~~~python\nprint(1)\n~~~",
			contains: []string{
				`<pre><code class="language-python">print(1)</code></pre>`,
			},
		},
		{
			name:  "mermaid diagram code block",
			input: "```mermaid\ngraph TD;\nA-->B;\n```",
			contains: []string{
				`<div class="mermaid-block">`,
				`<div class="mermaid-label">Diagram (Mermaid)</div>`,
				`<pre><code class="language-mermaid">graph TD;`,
				`A--&gt;B;</code></pre>`,
				`</div>`,
			},
		},
		{
			name:  "math display block",
			input: "$$\n\\frac{1}{2}\n$$",
			contains: []string{
				`<div class="math-block">`,
				`<div class="math-label">Math (LaTeX)</div>`,
				`<pre class="math-display"><code>\frac{1}{2}</code></pre>`,
				`</div>`,
			},
		},
		{
			name:  "callout alert note and tip",
			input: "> [!NOTE]\n> Important note here.\n\n> [!TIP]\n> Useful tip.",
			contains: []string{
				`<div class="markdown-alert markdown-alert-note"><p class="markdown-alert-title">NOTE</p><p>Important note here.</p></div>`,
				`<div class="markdown-alert markdown-alert-tip"><p class="markdown-alert-title">TIP</p><p>Useful tip.</p></div>`,
			},
		},
		{
			name:  "nested blockquotes",
			input: "> Outer quote\n>> Nested quote",
			contains: []string{
				`<blockquote>Outer quote<blockquote>Nested quote</blockquote></blockquote>`,
			},
		},
		{
			name:  "task list",
			input: "- [x] Done\n- [ ] Pending",
			contains: []string{
				`<ul><li class="task-list-item"><input type="checkbox" disabled checked /> Done</li><li class="task-list-item"><input type="checkbox" disabled /> Pending</li></ul>`,
			},
		},
		{
			name:  "nested unordered lists with plus and asterisk",
			input: "* Root\n  + Sub\n    - SubSub",
			contains: []string{
				`<ul><li>Root<ul><li>Sub<ul><li>SubSub</li></ul></li></ul></li></ul>`,
			},
		},
		{
			name:  "table with alignments",
			input: "| Left | Center | Right |\n| :--- | :---: | ---: |\n| A | B | C |",
			contains: []string{
				`<th style="text-align: left;">Left</th>`,
				`<th style="text-align: center;">Center</th>`,
				`<th style="text-align: right;">Right</th>`,
				`<td style="text-align: left;">A</td>`,
				`<td style="text-align: center;">B</td>`,
				`<td style="text-align: right;">C</td>`,
			},
		},
		{
			name:  "footnotes",
			input: "Important fact[^1].\n\n[^1]: Reference note text.",
			contains: []string{
				`<sup class="footnote-ref" id="fnref:1"><a href="#fn:1">[1]</a></sup>`,
				`<section class="footnotes">`,
				`<li id="fn:1"><p>Reference note text. <a href="#fnref:1" class="footnote-backref">↩</a></p></li>`,
			},
		},
		{
			name:  "soft line breaks with two trailing spaces",
			input: "First line  \nSecond line",
			contains: []string{
				`<p>First line<br />Second line</p>`,
			},
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got := markdown.Render(tc.input)
			for _, exp := range tc.contains {
				if !strings.Contains(got, exp) {
					t.Fatalf("Render(%q) = %q, missing expected substring %q", tc.input, got, exp)
				}
			}
		})
	}
}
