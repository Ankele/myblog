package service

import (
	"context"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"myblog/internal/conf"
	"myblog/internal/data"
)

const (
	PostStatusDraft     = "draft"
	PostStatusPublished = "published"
)

var (
	ErrUnauthorized = errors.New("username or password is incorrect")
	ErrNotFound     = errors.New("resource not found")
)

type AppService struct {
	cfg  *conf.Config
	repo *data.Repository
}

type PostInput struct {
	Title           string `json:"title"`
	Slug            string `json:"slug"`
	Summary         string `json:"summary"`
	CoverImage      string `json:"cover_image"`
	MarkdownContent string `json:"markdown_content"`
	Status          string `json:"status"`
	CategoryID      *uint  `json:"category_id"`
	TagIDs          []uint `json:"tag_ids"`
}

type CategoryInput struct {
	Name        string `json:"name"`
	Slug        string `json:"slug"`
	Description string `json:"description"`
}

type TagInput struct {
	Name        string `json:"name"`
	Slug        string `json:"slug"`
	Description string `json:"description"`
}

type SiteSettingInput struct {
	SiteName       string `json:"site_name"`
	Tagline        string `json:"tagline"`
	Description    string `json:"description"`
	Avatar         string `json:"avatar"`
	FooterText     string `json:"footer_text"`
	SEOTitle       string `json:"seo_title"`
	SEODescription string `json:"seo_description"`
	GithubURL      string `json:"github_url"`
	TwitterURL     string `json:"twitter_url"`
	LinkedInURL    string `json:"linkedin_url"`
	Email          string `json:"email"`
	AboutTitle     string `json:"about_title"`
	AboutContent   string `json:"about_content"`
}

func NewAppService(cfg *conf.Config, repo *data.Repository) *AppService {
	return &AppService{cfg: cfg, repo: repo}
}

func (s *AppService) Authenticate(ctx context.Context, username, password string) (*data.AdminUser, error) {
	admin, err := s.repo.FindAdminByUsername(ctx, strings.TrimSpace(username))
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrUnauthorized
		}
		return nil, err
	}
	if err := bcrypt.CompareHashAndPassword([]byte(admin.PasswordHash), []byte(password)); err != nil {
		return nil, ErrUnauthorized
	}
	return admin, nil
}

func (s *AppService) CurrentAdmin(ctx context.Context, id uint) (*data.AdminUser, error) {
	admin, err := s.repo.FindAdminByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return admin, nil
}

func (s *AppService) ListPublicPosts(ctx context.Context, filter data.ListPostsFilter) ([]data.Post, int64, error) {
	return s.repo.ListPublicPosts(ctx, normalizeFilter(filter))
}

func (s *AppService) ListAllPublishedPosts(ctx context.Context) ([]data.Post, error) {
	return s.repo.ListAllPublishedPosts(ctx)
}

func (s *AppService) GetPublicPost(ctx context.Context, slug string) (*data.Post, error) {
	post, err := s.repo.GetPublicPostBySlug(ctx, slug)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return post, nil
}

func (s *AppService) ListAdminPosts(ctx context.Context, filter data.ListPostsFilter) ([]data.Post, int64, error) {
	return s.repo.ListAdminPosts(ctx, normalizeFilter(filter))
}

func (s *AppService) GetAdminPost(ctx context.Context, id uint) (*data.Post, error) {
	post, err := s.repo.GetAdminPostByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return post, nil
}

func (s *AppService) CreatePost(ctx context.Context, input PostInput) (*data.Post, error) {
	post, tagIDs, err := s.buildPost(ctx, 0, nil, input)
	if err != nil {
		return nil, err
	}
	if err := s.repo.CreatePost(ctx, post, tagIDs); err != nil {
		return nil, err
	}
	return s.repo.GetAdminPostByID(ctx, post.ID)
}

func (s *AppService) UpdatePost(ctx context.Context, id uint, input PostInput) (*data.Post, error) {
	current, err := s.repo.GetAdminPostByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}

	post, tagIDs, err := s.buildPost(ctx, id, current, input)
	if err != nil {
		return nil, err
	}
	if err := s.repo.UpdatePost(ctx, post, tagIDs); err != nil {
		return nil, err
	}
	return s.repo.GetAdminPostByID(ctx, post.ID)
}

func (s *AppService) DeletePost(ctx context.Context, id uint) error {
	return s.repo.DeletePost(ctx, id)
}

func (s *AppService) ListCategories(ctx context.Context) ([]data.Category, error) {
	return s.repo.ListCategories(ctx)
}

func (s *AppService) CreateCategory(ctx context.Context, input CategoryInput) (*data.Category, error) {
	slug := s.uniqueSlug(ctx, &data.Category{}, strings.TrimSpace(input.Slug), input.Name, 0)
	category := &data.Category{
		Name:        strings.TrimSpace(input.Name),
		Slug:        slug,
		Description: strings.TrimSpace(input.Description),
	}
	if category.Name == "" {
		return nil, errors.New("category name is required")
	}
	if err := s.repo.CreateCategory(ctx, category); err != nil {
		return nil, err
	}
	return category, nil
}

func (s *AppService) UpdateCategory(ctx context.Context, id uint, input CategoryInput) (*data.Category, error) {
	category, err := s.repo.GetCategoryByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	if strings.TrimSpace(input.Name) == "" {
		return nil, errors.New("category name is required")
	}
	category.Name = strings.TrimSpace(input.Name)
	category.Description = strings.TrimSpace(input.Description)
	category.Slug = s.uniqueSlug(ctx, &data.Category{}, strings.TrimSpace(input.Slug), category.Name, category.ID)
	if err := s.repo.UpdateCategory(ctx, category); err != nil {
		return nil, err
	}
	return category, nil
}

func (s *AppService) DeleteCategory(ctx context.Context, id uint) error {
	return s.repo.DeleteCategory(ctx, id)
}

func (s *AppService) ListTags(ctx context.Context) ([]data.Tag, error) {
	return s.repo.ListTags(ctx)
}

func (s *AppService) CreateTag(ctx context.Context, input TagInput) (*data.Tag, error) {
	if strings.TrimSpace(input.Name) == "" {
		return nil, errors.New("tag name is required")
	}
	tag := &data.Tag{
		Name:        strings.TrimSpace(input.Name),
		Slug:        s.uniqueSlug(ctx, &data.Tag{}, strings.TrimSpace(input.Slug), input.Name, 0),
		Description: strings.TrimSpace(input.Description),
	}
	if err := s.repo.CreateTag(ctx, tag); err != nil {
		return nil, err
	}
	return tag, nil
}

func (s *AppService) UpdateTag(ctx context.Context, id uint, input TagInput) (*data.Tag, error) {
	tag, err := s.repo.GetTagByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	if strings.TrimSpace(input.Name) == "" {
		return nil, errors.New("tag name is required")
	}
	tag.Name = strings.TrimSpace(input.Name)
	tag.Description = strings.TrimSpace(input.Description)
	tag.Slug = s.uniqueSlug(ctx, &data.Tag{}, strings.TrimSpace(input.Slug), tag.Name, tag.ID)
	if err := s.repo.UpdateTag(ctx, tag); err != nil {
		return nil, err
	}
	return tag, nil
}

func (s *AppService) DeleteTag(ctx context.Context, id uint) error {
	return s.repo.DeleteTag(ctx, id)
}

func (s *AppService) GetSiteSetting(ctx context.Context) (*data.SiteSetting, error) {
	site, err := s.repo.GetSiteSetting(ctx)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return site, nil
}

func (s *AppService) UpdateSiteSetting(ctx context.Context, input SiteSettingInput) (*data.SiteSetting, error) {
	site, err := s.repo.GetSiteSetting(ctx)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}

	if strings.TrimSpace(input.SiteName) == "" {
		return nil, errors.New("site name is required")
	}

	site.SiteName = strings.TrimSpace(input.SiteName)
	site.Tagline = strings.TrimSpace(input.Tagline)
	site.Description = strings.TrimSpace(input.Description)
	site.Avatar = strings.TrimSpace(input.Avatar)
	site.FooterText = strings.TrimSpace(input.FooterText)
	site.SEOTitle = strings.TrimSpace(input.SEOTitle)
	site.SEODescription = strings.TrimSpace(input.SEODescription)
	site.GithubURL = strings.TrimSpace(input.GithubURL)
	site.TwitterURL = strings.TrimSpace(input.TwitterURL)
	site.LinkedInURL = strings.TrimSpace(input.LinkedInURL)
	site.Email = strings.TrimSpace(input.Email)
	site.AboutTitle = strings.TrimSpace(input.AboutTitle)
	site.AboutContent = strings.TrimSpace(input.AboutContent)

	if err := s.repo.UpdateSiteSetting(ctx, site); err != nil {
		return nil, err
	}
	return site, nil
}

func (s *AppService) RenderPreview(markdown string) string {
	return RenderMarkdown(markdown)
}

func (s *AppService) SaveUpload(file multipart.File, header *multipart.FileHeader) (string, error) {
	defer file.Close()

	if err := os.MkdirAll(s.cfg.App.UploadDir, 0o755); err != nil {
		return "", err
	}

	ext := strings.ToLower(filepath.Ext(header.Filename))
	if ext == "" {
		ext = ".bin"
	}

	name := fmt.Sprintf("%d-%s%s", time.Now().UnixNano(), slugify(strings.TrimSuffix(header.Filename, filepath.Ext(header.Filename))), ext)
	if strings.Contains(name, "--") || strings.HasSuffix(name, "-"+ext) {
		name = fmt.Sprintf("%d-asset%s", time.Now().UnixNano(), ext)
	}
	fullPath := filepath.Join(s.cfg.App.UploadDir, name)

	dst, err := os.Create(fullPath)
	if err != nil {
		return "", err
	}
	defer dst.Close()

	if _, err := io.Copy(dst, file); err != nil {
		return "", err
	}

	return "/uploads/" + name, nil
}

func (s *AppService) buildPost(ctx context.Context, id uint, current *data.Post, input PostInput) (*data.Post, []uint, error) {
	title := strings.TrimSpace(input.Title)
	if title == "" {
		return nil, nil, errors.New("title is required")
	}

	markdown := strings.TrimSpace(input.MarkdownContent)
	if markdown == "" {
		return nil, nil, errors.New("markdown content is required")
	}

	status := strings.TrimSpace(input.Status)
	if status == "" {
		status = PostStatusDraft
	}
	if status != PostStatusDraft && status != PostStatusPublished {
		return nil, nil, errors.New("invalid post status")
	}

	slug := s.uniqueSlug(ctx, &data.Post{}, strings.TrimSpace(input.Slug), title, id)
	now := time.Now()

	post := &data.Post{
		Title:           title,
		Slug:            slug,
		Summary:         strings.TrimSpace(input.Summary),
		CoverImage:      strings.TrimSpace(input.CoverImage),
		MarkdownContent: markdown,
		HTMLContent:     RenderMarkdown(markdown),
		Status:          status,
		CategoryID:      input.CategoryID,
	}
	if current != nil {
		post.ID = current.ID
		post.CreatedAt = current.CreatedAt
	}

	if status == PostStatusPublished {
		if current != nil && current.PublishedAt != nil {
			post.PublishedAt = current.PublishedAt
		} else {
			post.PublishedAt = &now
		}
	}

	return post, input.TagIDs, nil
}

func (s *AppService) uniqueSlug(ctx context.Context, model any, requested, fallback string, excludeID uint) string {
	base := slugify(requested)
	if base == "" {
		base = slugify(fallback)
	}
	if base == "" {
		base = "item"
	}

	slug := base
	for i := 1; i < 1000; i++ {
		exists, err := s.repo.SlugExists(ctx, model, slug, excludeID)
		if err == nil && !exists {
			return slug
		}
		slug = base + "-" + strconv.Itoa(i+1)
	}
	return fmt.Sprintf("%s-%d", base, time.Now().Unix())
}

var nonSlugPattern = regexp.MustCompile(`[^a-z0-9]+`)

func slugify(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	value = strings.ReplaceAll(value, "_", "-")
	value = nonSlugPattern.ReplaceAllString(value, "-")
	value = strings.Trim(value, "-")
	return value
}

func normalizeFilter(filter data.ListPostsFilter) data.ListPostsFilter {
	if filter.Page <= 0 {
		filter.Page = 1
	}
	if filter.PageSize <= 0 {
		filter.PageSize = 10
	}
	if filter.PageSize > 100 {
		filter.PageSize = 100
	}
	return filter
}
