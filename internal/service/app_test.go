package service

import (
	"testing"

	"myblog/internal/data"
)

func TestSlugify(t *testing.T) {
	if got := slugify(" Hello, Vue + Go Blog! "); got != "hello-vue-go-blog" {
		t.Fatalf("unexpected slug: %s", got)
	}
}

func TestNormalizeFilter(t *testing.T) {
	filter := normalizeFilter(data.ListPostsFilter{
		Page:     0,
		PageSize: 500,
	})

	if filter.Page != 1 {
		t.Fatalf("expected default page 1, got %d", filter.Page)
	}
	if filter.PageSize != 100 {
		t.Fatalf("expected page size cap 100, got %d", filter.PageSize)
	}
}

func TestNormalizeUserCredentials(t *testing.T) {
	username, email, password, err := normalizeUserCredentials("Alice_01", "Alice@example.com", "supersecret1")
	if err != nil {
		t.Fatalf("expected valid credentials, got %v", err)
	}
	if username != "alice_01" {
		t.Fatalf("unexpected username: %s", username)
	}
	if email != "alice@example.com" {
		t.Fatalf("unexpected email: %s", email)
	}
	if password != "supersecret1" {
		t.Fatalf("unexpected password normalization: %s", password)
	}
}

func TestNormalizeUserCredentialsRejectsWeakPassword(t *testing.T) {
	if _, _, _, err := normalizeUserCredentials("alice", "alice@example.com", "short"); err == nil {
		t.Fatal("expected short password to be rejected")
	}
}

func TestParseMarkdownFrontMatter(t *testing.T) {
	meta, body := parseMarkdownFrontMatter("---\ntitle: Imported Post\nslug: imported-post\nsummary: custom summary\n---\n# Imported Post\n\nHello world.")
	if meta.Title != "Imported Post" || meta.Slug != "imported-post" {
		t.Fatalf("unexpected metadata: %+v", meta)
	}
	if body != "# Imported Post\n\nHello world." {
		t.Fatalf("unexpected body: %q", body)
	}
}

func TestExtractMarkdownTitleAndSummary(t *testing.T) {
	title := extractMarkdownTitle("# Hello Title\n\nThis is the first paragraph.", "demo-file.md")
	if title != "Hello Title" {
		t.Fatalf("unexpected title: %s", title)
	}

	summary := extractMarkdownSummary("# Title\n\nThis is the first paragraph with [link](https://example.com) and `code`.")
	if summary == "" || summary == "# Title" {
		t.Fatalf("unexpected summary: %q", summary)
	}
}
