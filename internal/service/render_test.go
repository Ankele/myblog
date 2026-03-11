package service

import (
	"strings"
	"testing"
)

func TestRenderMarkdownSanitizesDangerousHTML(t *testing.T) {
	input := "# Title\n\n<script>alert('x')</script>\n\n[bad](javascript:alert(1))\n\n[good](/posts/demo)"
	output := RenderMarkdown(input)

	if strings.Contains(output, "<script") {
		t.Fatalf("expected script tags to be stripped, got %s", output)
	}
	if strings.Contains(output, "javascript:alert") {
		t.Fatalf("expected javascript urls to be stripped, got %s", output)
	}
	if !strings.Contains(output, `<a href="/posts/demo"`) {
		t.Fatalf("expected safe relative link to be kept, got %s", output)
	}
}

func TestRenderMarkdownKeepsCodeBlocks(t *testing.T) {
	input := "```go\nfmt.Println(\"hello\")\n```"
	output := RenderMarkdown(input)

	if !strings.Contains(output, "<pre") || !strings.Contains(output, "<code") {
		t.Fatalf("expected fenced code block markup, got %s", output)
	}
}
