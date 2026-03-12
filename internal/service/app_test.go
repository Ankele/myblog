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

func TestValidatePassword(t *testing.T) {
	cases := []struct {
		password string
		valid    bool
	}{
		{password: "abc12345", valid: true},
		{password: "12345678", valid: false},
		{password: "abcdefgh", valid: false},
		{password: "a1", valid: false},
	}

	for _, tc := range cases {
		err := validatePassword(tc.password)
		if tc.valid && err != nil {
			t.Fatalf("password %q should be valid, got error %v", tc.password, err)
		}
		if !tc.valid && err == nil {
			t.Fatalf("password %q should be invalid", tc.password)
		}
	}
}

func TestValidEmail(t *testing.T) {
	if !validEmail("demo@example.com") {
		t.Fatalf("expected valid email")
	}
	if validEmail("bad-email") {
		t.Fatalf("expected invalid email")
	}
}
