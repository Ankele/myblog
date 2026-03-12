package data

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"myblog/internal/conf"
)

type Repository struct {
	db *gorm.DB
}

type ListPostsFilter struct {
	Page         int
	PageSize     int
	CategorySlug string
	TagSlug      string
	Status       string
}

func OpenDB(cfg *conf.Config) (*gorm.DB, error) {
	if strings.ToLower(cfg.Data.Database.Type) != "postgres" {
		return nil, fmt.Errorf("unsupported database type: %s", cfg.Data.Database.Type)
	}

	return gorm.Open(postgres.Open(cfg.Data.Database.Source), &gorm.Config{})
}

func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

func MigrateAndSeed(ctx context.Context, db *gorm.DB, cfg *conf.Config) error {
	if err := db.WithContext(ctx).AutoMigrate(
		&AdminUser{},
		&User{},
		&Category{},
		&Tag{},
		&Post{},
		&PostTag{},
		&SiteSetting{},
	); err != nil {
		return err
	}

	var adminCount int64
	if err := db.WithContext(ctx).Model(&AdminUser{}).Count(&adminCount).Error; err != nil {
		return err
	}
	if adminCount == 0 {
		passwordHash, err := bcrypt.GenerateFromPassword([]byte(cfg.App.DefaultAdminPassword), bcrypt.DefaultCost)
		if err != nil {
			return err
		}
		admin := AdminUser{
			Username:     cfg.App.DefaultAdminUsername,
			PasswordHash: string(passwordHash),
		}
		if err := db.WithContext(ctx).Create(&admin).Error; err != nil {
			return err
		}
	}

	var siteCount int64
	if err := db.WithContext(ctx).Model(&SiteSetting{}).Count(&siteCount).Error; err != nil {
		return err
	}
	if siteCount == 0 {
		site := SiteSetting{
			SiteName:       "Ankele Blog",
			Tagline:        "记录技术、写作和日常观察",
			Description:    "一个使用 Go 与 Vue 构建的个人博客。",
			FooterText:     "Built with Go and Vue",
			SEOTitle:       "Ankele Blog",
			SEODescription: "一个使用 Go 与 Vue 构建的个人博客。",
			AboutTitle:     "关于我",
			AboutContent:   "在这里分享技术思考、工作笔记和长期创作。",
		}
		if err := db.WithContext(ctx).Create(&site).Error; err != nil {
			return err
		}
	}

	return nil
}

func (r *Repository) FindAdminByUsername(ctx context.Context, username string) (*AdminUser, error) {
	var admin AdminUser
	if err := r.db.WithContext(ctx).Where("username = ?", username).First(&admin).Error; err != nil {
		return nil, err
	}
	return &admin, nil
}

func (r *Repository) FindAdminByID(ctx context.Context, id uint) (*AdminUser, error) {
	var admin AdminUser
	if err := r.db.WithContext(ctx).First(&admin, id).Error; err != nil {
		return nil, err
	}
	return &admin, nil
}

func (r *Repository) CreateUser(ctx context.Context, user *User) error {
	return r.db.WithContext(ctx).Create(user).Error
}

func (r *Repository) FindUserByID(ctx context.Context, id string) (*User, error) {
	var user User
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *Repository) FindUserByAccount(ctx context.Context, account string) (*User, error) {
	var user User
	query := "LOWER(username) = LOWER(?) OR LOWER(email) = LOWER(?)"
	if err := r.db.WithContext(ctx).Where(query, account, account).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *Repository) ListPublicPosts(ctx context.Context, filter ListPostsFilter) ([]Post, int64, error) {
	query := r.publicPostQuery(ctx, filter)

	var total int64
	if err := query.Model(&Post{}).Distinct("posts.id").Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var posts []Post
	err := query.
		Preload("Category").
		Preload("Tags").
		Order("published_at desc nulls last, created_at desc").
		Limit(filter.PageSize).
		Offset((filter.Page - 1) * filter.PageSize).
		Find(&posts).Error

	return posts, total, err
}

func (r *Repository) ListAllPublishedPosts(ctx context.Context) ([]Post, error) {
	var posts []Post
	err := r.db.WithContext(ctx).
		Model(&Post{}).
		Where("status = ?", "published").
		Order("published_at desc nulls last, created_at desc").
		Find(&posts).Error
	return posts, err
}

func (r *Repository) GetPublicPostBySlug(ctx context.Context, slug string) (*Post, error) {
	var post Post
	err := r.db.WithContext(ctx).
		Preload("Category").
		Preload("Tags").
		Where("slug = ? AND status = ?", slug, "published").
		First(&post).Error
	if err != nil {
		return nil, err
	}
	return &post, nil
}

func (r *Repository) ListAdminPosts(ctx context.Context, filter ListPostsFilter) ([]Post, int64, error) {
	query := r.db.WithContext(ctx).Model(&Post{})

	if filter.Status != "" {
		query = query.Where("status = ?", filter.Status)
	}
	if filter.CategorySlug != "" {
		query = query.Joins("LEFT JOIN categories ON categories.id = posts.category_id").
			Where("categories.slug = ?", filter.CategorySlug)
	}
	if filter.TagSlug != "" {
		query = query.
			Joins("JOIN post_tags ON post_tags.post_id = posts.id").
			Joins("JOIN tags ON tags.id = post_tags.tag_id").
			Where("tags.slug = ?", filter.TagSlug)
	}

	var total int64
	if err := query.Distinct("posts.id").Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var posts []Post
	err := query.
		Preload("Category").
		Preload("Tags").
		Distinct("posts.*").
		Order("updated_at desc").
		Limit(filter.PageSize).
		Offset((filter.Page - 1) * filter.PageSize).
		Find(&posts).Error

	return posts, total, err
}

func (r *Repository) GetAdminPostByID(ctx context.Context, id uint) (*Post, error) {
	var post Post
	err := r.db.WithContext(ctx).
		Preload("Category").
		Preload("Tags").
		First(&post, id).Error
	if err != nil {
		return nil, err
	}
	return &post, nil
}

func (r *Repository) CreatePost(ctx context.Context, post *Post, tagIDs []uint) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(post).Error; err != nil {
			return err
		}
		return replacePostTags(tx, post, tagIDs)
	})
}

func (r *Repository) UpdatePost(ctx context.Context, post *Post, tagIDs []uint) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Session(&gorm.Session{FullSaveAssociations: false}).Save(post).Error; err != nil {
			return err
		}
		return replacePostTags(tx, post, tagIDs)
	})
}

func (r *Repository) DeletePost(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&Post{}, id).Error
}

func (r *Repository) ListCategories(ctx context.Context) ([]Category, error) {
	var categories []Category
	err := r.db.WithContext(ctx).Order("name asc").Find(&categories).Error
	return categories, err
}

func (r *Repository) CreateCategory(ctx context.Context, category *Category) error {
	return r.db.WithContext(ctx).Create(category).Error
}

func (r *Repository) UpdateCategory(ctx context.Context, category *Category) error {
	return r.db.WithContext(ctx).Save(category).Error
}

func (r *Repository) GetCategoryByID(ctx context.Context, id uint) (*Category, error) {
	var category Category
	if err := r.db.WithContext(ctx).First(&category, id).Error; err != nil {
		return nil, err
	}
	return &category, nil
}

func (r *Repository) DeleteCategory(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&Post{}).Where("category_id = ?", id).Update("category_id", nil).Error; err != nil {
			return err
		}
		return tx.Delete(&Category{}, id).Error
	})
}

func (r *Repository) ListTags(ctx context.Context) ([]Tag, error) {
	var tags []Tag
	err := r.db.WithContext(ctx).Order("name asc").Find(&tags).Error
	return tags, err
}

func (r *Repository) CreateTag(ctx context.Context, tag *Tag) error {
	return r.db.WithContext(ctx).Create(tag).Error
}

func (r *Repository) UpdateTag(ctx context.Context, tag *Tag) error {
	return r.db.WithContext(ctx).Save(tag).Error
}

func (r *Repository) GetTagByID(ctx context.Context, id uint) (*Tag, error) {
	var tag Tag
	if err := r.db.WithContext(ctx).First(&tag, id).Error; err != nil {
		return nil, err
	}
	return &tag, nil
}

func (r *Repository) DeleteTag(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("tag_id = ?", id).Delete(&PostTag{}).Error; err != nil {
			return err
		}
		return tx.Delete(&Tag{}, id).Error
	})
}

func (r *Repository) GetSiteSetting(ctx context.Context) (*SiteSetting, error) {
	var site SiteSetting
	if err := r.db.WithContext(ctx).First(&site).Error; err != nil {
		return nil, err
	}
	return &site, nil
}

func (r *Repository) UpdateSiteSetting(ctx context.Context, site *SiteSetting) error {
	return r.db.WithContext(ctx).Save(site).Error
}

func (r *Repository) SlugExists(ctx context.Context, model any, slug string, excludeID uint) (bool, error) {
	query := r.db.WithContext(ctx).Model(model).Where("slug = ?", slug)
	if excludeID > 0 {
		query = query.Where("id <> ?", excludeID)
	}

	var count int64
	if err := query.Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *Repository) publicPostQuery(ctx context.Context, filter ListPostsFilter) *gorm.DB {
	query := r.db.WithContext(ctx).Model(&Post{}).Where("posts.status = ?", "published")
	if filter.CategorySlug != "" {
		query = query.Joins("LEFT JOIN categories ON categories.id = posts.category_id").
			Where("categories.slug = ?", filter.CategorySlug)
	}
	if filter.TagSlug != "" {
		query = query.
			Joins("JOIN post_tags ON post_tags.post_id = posts.id").
			Joins("JOIN tags ON tags.id = post_tags.tag_id").
			Where("tags.slug = ?", filter.TagSlug)
	}
	return query
}

func replacePostTags(tx *gorm.DB, post *Post, tagIDs []uint) error {
	if tagIDs == nil {
		return nil
	}

	var tags []Tag
	if len(tagIDs) > 0 {
		if err := tx.Where("id IN ?", tagIDs).Find(&tags).Error; err != nil {
			return err
		}
		if len(tags) != len(tagIDs) {
			return errors.New("one or more tags not found")
		}
	}
	return tx.Model(post).Association("Tags").Replace(tags)
}
