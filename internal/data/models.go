package data

import "time"

type AdminUser struct {
	ID           uint      `gorm:"primaryKey" json:"id"`
	Username     string    `gorm:"size:64;uniqueIndex;not null" json:"username"`
	PasswordHash string    `gorm:"size:255;not null" json:"-"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type User struct {
	ID           string    `gorm:"primaryKey;size:32" json:"id"`
	Username     string    `gorm:"size:64;uniqueIndex;not null" json:"username"`
	Email        string    `gorm:"size:255;uniqueIndex;not null" json:"email"`
	PasswordHash string    `gorm:"size:255;not null" json:"-"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type Category struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	Name        string    `gorm:"size:120;not null" json:"name"`
	Slug        string    `gorm:"size:150;uniqueIndex;not null" json:"slug"`
	Description string    `gorm:"size:500" json:"description"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type Tag struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	Name        string    `gorm:"size:120;not null" json:"name"`
	Slug        string    `gorm:"size:150;uniqueIndex;not null" json:"slug"`
	Description string    `gorm:"size:500" json:"description"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type Post struct {
	ID              uint       `gorm:"primaryKey" json:"id"`
	Title           string     `gorm:"size:255;not null" json:"title"`
	Slug            string     `gorm:"size:180;uniqueIndex;not null" json:"slug"`
	Summary         string     `gorm:"size:600" json:"summary"`
	CoverImage      string     `gorm:"size:500" json:"cover_image"`
	MarkdownContent string     `gorm:"type:text;not null" json:"markdown_content"`
	HTMLContent     string     `gorm:"type:text;not null" json:"html_content"`
	Status          string     `gorm:"size:20;index:idx_posts_status_published,priority:1;not null" json:"status"`
	PublishedAt     *time.Time `gorm:"index:idx_posts_status_published,priority:2" json:"published_at"`
	CategoryID      *uint      `json:"category_id"`
	Category        *Category  `json:"category,omitempty"`
	Tags            []Tag      `gorm:"many2many:post_tags;" json:"tags,omitempty"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
}

type PostTag struct {
	PostID uint `gorm:"primaryKey"`
	TagID  uint `gorm:"primaryKey"`
}

func (PostTag) TableName() string {
	return "post_tags"
}

type SiteSetting struct {
	ID             uint      `gorm:"primaryKey" json:"id"`
	SiteName       string    `gorm:"size:180;not null" json:"site_name"`
	Tagline        string    `gorm:"size:255" json:"tagline"`
	Description    string    `gorm:"size:500" json:"description"`
	Avatar         string    `gorm:"size:500" json:"avatar"`
	FooterText     string    `gorm:"size:255" json:"footer_text"`
	SEOTitle       string    `gorm:"size:255" json:"seo_title"`
	SEODescription string    `gorm:"size:500" json:"seo_description"`
	GithubURL      string    `gorm:"size:500" json:"github_url"`
	TwitterURL     string    `gorm:"size:500" json:"twitter_url"`
	LinkedInURL    string    `gorm:"size:500" json:"linkedin_url"`
	Email          string    `gorm:"size:255" json:"email"`
	AboutTitle     string    `gorm:"size:255" json:"about_title"`
	AboutContent   string    `gorm:"type:text" json:"about_content"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}
