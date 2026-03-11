package service

import (
	"bufio"
	"context"
	"errors"
	"io"
	"mime/multipart"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"gopkg.in/yaml.v3"

	"myblog/internal/data"
)

var markdownExtPattern = map[string]bool{
	".md":       true,
	".markdown": true,
	".txt":      true,
}

type MarkdownImportInput struct {
	Status     string
	CategoryID *uint
	TagIDs     []uint
	CoverImage string
}

type markdownFrontMatter struct {
	Title      string `yaml:"title"`
	Slug       string `yaml:"slug"`
	Summary    string `yaml:"summary"`
	CoverImage string `yaml:"cover_image"`
	Status     string `yaml:"status"`
}

func (s *AppService) ImportMarkdownPost(ctx context.Context, file multipart.File, header *multipart.FileHeader, input MarkdownImportInput) (*data.Post, error) {
	defer file.Close()

	ext := strings.ToLower(filepath.Ext(header.Filename))
	if !markdownExtPattern[ext] {
		return nil, errors.New("only markdown files are supported")
	}

	content, err := io.ReadAll(io.LimitReader(file, 5<<20))
	if err != nil {
		return nil, err
	}

	markdown := strings.TrimSpace(string(content))
	if markdown == "" {
		return nil, errors.New("markdown file is empty")
	}

	meta, body := parseMarkdownFrontMatter(markdown)
	title := strings.TrimSpace(meta.Title)
	if title == "" {
		title = extractMarkdownTitle(body, header.Filename)
	}

	status := strings.TrimSpace(input.Status)
	if status == "" {
		status = strings.TrimSpace(meta.Status)
	}
	if status == "" {
		status = PostStatusDraft
	}

	summary := strings.TrimSpace(meta.Summary)
	if summary == "" {
		summary = extractMarkdownSummary(body)
	}

	coverImage := strings.TrimSpace(input.CoverImage)
	if coverImage == "" {
		coverImage = strings.TrimSpace(meta.CoverImage)
	}

	return s.CreatePost(ctx, PostInput{
		Title:           title,
		Slug:            strings.TrimSpace(meta.Slug),
		Summary:         summary,
		CoverImage:      coverImage,
		MarkdownContent: body,
		Status:          status,
		CategoryID:      input.CategoryID,
		TagIDs:          input.TagIDs,
	})
}

func ParseUintCSV(raw string) ([]uint, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil, nil
	}

	parts := strings.Split(raw, ",")
	values := make([]uint, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		parsed, err := strconv.ParseUint(part, 10, 64)
		if err != nil {
			return nil, errors.New("tag_ids must be a comma separated list of integers")
		}
		values = append(values, uint(parsed))
	}
	return values, nil
}

func ParseOptionalUint(raw string) (*uint, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil, nil
	}
	parsed, err := strconv.ParseUint(raw, 10, 64)
	if err != nil {
		return nil, errors.New("category_id must be an integer")
	}
	value := uint(parsed)
	return &value, nil
}

func parseMarkdownFrontMatter(input string) (markdownFrontMatter, string) {
	var meta markdownFrontMatter
	if !strings.HasPrefix(input, "---\n") && !strings.HasPrefix(input, "---\r\n") {
		return meta, input
	}

	scanner := bufio.NewScanner(strings.NewReader(input))
	var lines []string
	var bodyLines []string
	lineIndex := 0
	inFrontMatter := false
	foundClosing := false

	for scanner.Scan() {
		line := scanner.Text()
		switch {
		case lineIndex == 0 && strings.TrimSpace(line) == "---":
			inFrontMatter = true
		case inFrontMatter && strings.TrimSpace(line) == "---":
			foundClosing = true
			inFrontMatter = false
		case inFrontMatter:
			lines = append(lines, line)
		case !inFrontMatter && foundClosing:
			bodyLines = append(bodyLines, line)
		}
		lineIndex++
	}

	if !foundClosing {
		return meta, input
	}

	_ = yaml.Unmarshal([]byte(strings.Join(lines, "\n")), &meta)
	body := strings.TrimSpace(strings.Join(bodyLines, "\n"))
	if body == "" {
		body = input
	}
	return meta, body
}

func extractMarkdownTitle(markdown string, filename string) string {
	scanner := bufio.NewScanner(strings.NewReader(markdown))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(line, "# ") {
			return strings.TrimSpace(strings.TrimPrefix(line, "# "))
		}
	}

	base := strings.TrimSuffix(filepath.Base(filename), filepath.Ext(filename))
	base = strings.ReplaceAll(base, "_", " ")
	base = strings.ReplaceAll(base, "-", " ")
	base = strings.TrimSpace(base)
	if base == "" {
		return "未命名文章"
	}
	return base
}

var markdownDecorationPattern = regexp.MustCompile("(```[\\s\\S]*?```|`[^`]+`|\\!\\[[^\\]]*\\]\\([^\\)]*\\)|\\[[^\\]]+\\]\\([^\\)]*\\)|[#>*_~-])")

func extractMarkdownSummary(markdown string) string {
	blocks := strings.Split(markdown, "\n\n")
	for _, block := range blocks {
		text := strings.TrimSpace(block)
		if text == "" || strings.HasPrefix(text, "#") || strings.HasPrefix(text, "```") {
			continue
		}

		cleaned := markdownDecorationPattern.ReplaceAllString(text, " ")
		cleaned = strings.Join(strings.Fields(cleaned), " ")
		if cleaned == "" {
			continue
		}
		runes := []rune(cleaned)
		if len(runes) > 140 {
			return string(runes[:140]) + "..."
		}
		return cleaned
	}
	return ""
}
