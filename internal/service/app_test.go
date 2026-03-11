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
