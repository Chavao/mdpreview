package markdown_test

import (
	"strings"
	"testing"

	"github.com/Chavao/mdpreview/internal/markdown"
)

func TestRenderSupportsCoreMarkdownElements(t *testing.T) {
	t.Parallel()

	input := "# Title\n\n- one\n- two\n\n| id | name |\n| --- | --- |\n| 1 | **John** |\n\n`x` and [go](https://go.dev)"
	got := markdown.Render(input)

	want := "<h1>Title</h1><ul><li>one</li><li>two</li></ul><table><thead><tr><th>id</th><th>name</th></tr></thead><tbody><tr><td>1</td><td><strong>John</strong></td></tr></tbody></table><p><code>x</code> and <a href=\"https://go.dev\" target=\"_blank\" rel=\"noreferrer\">go</a></p>"
	if got != want {
		t.Fatalf("Render() = %q, want %q", got, want)
	}
}

func TestRenderEscapesHTML(t *testing.T) {
	t.Parallel()

	got := markdown.Render("<script>alert(1)</script>")
	want := "<p>&lt;script&gt;alert(1)&lt;/script&gt;</p>"

	if got != want {
		t.Fatalf("Render() = %q, want %q", got, want)
	}
}

func TestRenderEmptyInput(t *testing.T) {
	t.Parallel()

	got := markdown.Render("\n\n")
	want := `<p class="empty">Start typing Markdown on the left.</p>`

	if got != want {
		t.Fatalf("Render() = %q, want %q", got, want)
	}
}

func TestRenderSpecification2026Showcase(t *testing.T) {
	t.Parallel()

	input := `# Header 1
## Header 2
### Header 3
#### Header 4
##### Header 5
###### Header 6

Line one  
Line two with <br /> break

> [!NOTE]
> Informational note.

> [!WARNING]
> Warning callout.

> Single quote
>> Nested quote

- Item
  - Subitem
* Asterisk item
+ Plus item

1. Step one
   1. Substep

- [x] Task done
- [ ] Task todo

` + "```typescript\nconst x = 1;\n```" + `

~~~bash
echo "test"
~~~

*italic* and _italic_
**bold** and __bold__
***bold italic*** and ___bold italic___
~~strike~~

[Link](https://example.com "Title") and <https://example.com>
![Image](https://example.com/img.png "Alt")

` + "`code`" + ` and $E = mc^2$

| Col A | Col B |
| :--- | ---: |
| Left | Right |

Fact[^1].

[^1]: Footnote details.

$$
\frac{a}{b}
$$

` + "```mermaid\ngraph TD;\nA-->B;\n```" + `

<span style="color: red;">Red</span>
`

	got := markdown.Render(input)

	expectedSubstrings := []string{
		"<h1>Header 1</h1>",
		"<h6>Header 6</h6>",
		"Line one<br />Line two",
		`class="markdown-alert markdown-alert-note"`,
		`class="markdown-alert markdown-alert-warning"`,
		"<blockquote>Single quote<blockquote>Nested quote</blockquote></blockquote>",
		"<li>Item<ul><li>Subitem</li></ul></li>",
		"<li>Asterisk item</li>",
		"<li>Plus item</li>",
		`class="task-list-item"`,
		`<input type="checkbox" disabled checked /> Task done`,
		`<input type="checkbox" disabled /> Task todo`,
		`<pre><code class="language-typescript">`,
		`<pre><code class="language-bash">`,
		"<em>italic</em>",
		"<strong>bold</strong>",
		"<strong><em>bold italic</em></strong>",
		"<del>strike</del>",
		`<a href="https://example.com" target="_blank" rel="noreferrer" title="Title">Link</a>`,
		`<a href="https://example.com" target="_blank" rel="noreferrer">https://example.com</a>`,
		`<img src="https://example.com/img.png" alt="Image" title="Alt" />`,
		"<code>code</code>",
		`<span class="math-inline"><code>E = mc^2</code></span>`,
		`<th style="text-align: left;">Col A</th>`,
		`<th style="text-align: right;">Col B</th>`,
		`<sup class="footnote-ref" id="fnref:1"><a href="#fn:1">[1]</a></sup>`,
		`<section class="footnotes">`,
		`<div class="math-block">`,
		`<div class="mermaid-block">`,
		`<span style="color: red;">Red</span>`,
	}

	for _, sub := range expectedSubstrings {
		if !contains(got, sub) {
			t.Fatalf("Render() output missing %q\nFull output:\n%s", sub, got)
		}
	}
}

func contains(s, substr string) bool {
	return strings.Contains(s, substr)
}

