package markdown_test

import (
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
