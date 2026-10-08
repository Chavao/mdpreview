package markdown_test

import (
	"testing"

	"github.com/Chavao/mdpreview/internal/markdown"
)

func TestRenderInlineElements(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "strikethrough",
			input: "~~deleted text~~",
			want:  "<p><del>deleted text</del></p>",
		},
		{
			name:  "underscore italic",
			input: "_italic text_",
			want:  "<p><em>italic text</em></p>",
		},
		{
			name:  "underscore bold",
			input: "__bold text__",
			want:  "<p><strong>bold text</strong></p>",
		},
		{
			name:  "triple asterisk bold italic",
			input: "***bold and italic***",
			want:  "<p><strong><em>bold and italic</em></strong></p>",
		},
		{
			name:  "triple underscore bold italic",
			input: "___bold and italic___",
			want:  "<p><strong><em>bold and italic</em></strong></p>",
		},
		{
			name:  "link with title",
			input: `[OpenAI](https://openai.com "Official Site")`,
			want:  `<p><a href="https://openai.com" target="_blank" rel="noreferrer" title="Official Site">OpenAI</a></p>`,
		},
		{
			name:  "autolink delimited",
			input: `<https://golang.org>`,
			want:  `<p><a href="https://golang.org" target="_blank" rel="noreferrer">https://golang.org</a></p>`,
		},
		{
			name:  "image with title",
			input: `![Alt text](https://example.com/pic.png "Image Title")`,
			want:  `<p><img src="https://example.com/pic.png" alt="Alt text" title="Image Title" /></p>`,
		},
		{
			name:  "image without title",
			input: `![Alt text](https://example.com/pic.png)`,
			want:  `<p><img src="https://example.com/pic.png" alt="Alt text" /></p>`,
		},
		{
			name:  "inline math",
			input: `The equation is $E = mc^2$.`,
			want:  `<p>The equation is <span class="math-inline"><code>E = mc^2</code></span>.</p>`,
		},
		{
			name:  "safe HTML span preserved",
			input: `<span style="color: red;">Red text</span>`,
			want:  `<p><span style="color: red;">Red text</span></p>`,
		},
		{
			name:  "dangerous script tag escaped",
			input: `<script>alert(1)</script>`,
			want:  `<p>&lt;script&gt;alert(1)&lt;/script&gt;</p>`,
		},
		{
			name:  "dangerous event handlers stripped",
			input: `<img src="pic.png" onerror="alert(1)" alt="photo" />`,
			want:  `<p><img src="pic.png" alt="photo" /></p>`,
		},
		{
			name:  "code span preserves raw HTML characters",
			input: "`<script>` and `<div>`",
			want:  "<p><code>&lt;script&gt;</code> and <code>&lt;div&gt;</code></p>",
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got := markdown.Render(tc.input)
			if got != tc.want {
				t.Fatalf("Render(%q) = %q, want %q", tc.input, got, tc.want)
			}
		})
	}
}
