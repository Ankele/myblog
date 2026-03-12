package handler

import (
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/sessions"

	"myblog/internal/conf"
	"myblog/internal/data"
	"myblog/internal/service"
)

const (
	adminSessionName  = "myblog_admin"
	userSessionName   = "myblog_user"
	jsonBodyLimit     = int64(1 << 20) // 1 MiB
	maxUploadBodySize = int64(10 << 20)
)

type HTTPHandler struct {
	cfg         *conf.Config
	service     *service.AppService
	sessions    sessions.Store
	authLimiter *authRateLimiter
}

type apiError struct {
	Error string `json:"error"`
}

type sitemapURLSet struct {
	XMLName xml.Name          `xml:"urlset"`
	Xmlns   string            `xml:"xmlns,attr"`
	URLs    []sitemapURLEntry `xml:"url"`
}

type sitemapURLEntry struct {
	Loc     string `xml:"loc"`
	LastMod string `xml:"lastmod,omitempty"`
}

type authRateLimiter struct {
	mu      sync.Mutex
	windows map[string]authWindow
}

type authWindow struct {
	Count   int
	ResetAt time.Time
}

func NewHTTPHandler(cfg *conf.Config, service *service.AppService) *HTTPHandler {
	store := sessions.NewCookieStore([]byte(cfg.App.SessionSecret))
	store.Options = &sessions.Options{
		Path:     "/",
		MaxAge:   7 * 24 * 3600,
		Secure:   cfg.App.SessionCookieSecure,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	}
	return &HTTPHandler{
		cfg:         cfg,
		service:     service,
		sessions:    store,
		authLimiter: newAuthRateLimiter(),
	}
}

func (h *HTTPHandler) Routes() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("/api/posts", h.handlePublicPosts)
	mux.HandleFunc("/api/posts/", h.handlePublicPostDetail)
	mux.HandleFunc("/api/categories", h.handlePublicCategories)
	mux.HandleFunc("/api/tags", h.handlePublicTags)
	mux.HandleFunc("/api/site", h.handlePublicSite)
	mux.HandleFunc("/api/auth/register", h.handleUserRegister)
	mux.HandleFunc("/api/auth/login", h.handleUserLogin)
	mux.HandleFunc("/api/auth/logout", h.requireUser(h.handleUserLogout))
	mux.HandleFunc("/api/auth/me", h.requireUser(h.handleUserMe))

	mux.HandleFunc("/api/admin/login", h.handleAdminLogin)
	mux.HandleFunc("/api/admin/logout", h.requireAdmin(h.handleAdminLogout))
	mux.HandleFunc("/api/admin/me", h.requireAdmin(h.handleAdminMe))
	mux.HandleFunc("/api/admin/posts", h.requireAdmin(h.handleAdminPosts))
	mux.HandleFunc("/api/admin/posts/import-markdown", h.requireAdmin(h.handleAdminImportMarkdown))
	mux.HandleFunc("/api/admin/posts/preview", h.requireAdmin(h.handleAdminPreview))
	mux.HandleFunc("/api/admin/posts/", h.requireAdmin(h.handleAdminPostItem))
	mux.HandleFunc("/api/admin/categories", h.requireAdmin(h.handleAdminCategories))
	mux.HandleFunc("/api/admin/categories/", h.requireAdmin(h.handleAdminCategoryItem))
	mux.HandleFunc("/api/admin/tags", h.requireAdmin(h.handleAdminTags))
	mux.HandleFunc("/api/admin/tags/", h.requireAdmin(h.handleAdminTagItem))
	mux.HandleFunc("/api/admin/site", h.requireAdmin(h.handleAdminSite))
	mux.HandleFunc("/api/admin/upload", h.requireAdmin(h.handleAdminUpload))

	mux.HandleFunc("/sitemap.xml", h.handleSitemap)
	mux.Handle("/uploads/", http.StripPrefix("/uploads/", http.FileServer(http.Dir(h.cfg.App.UploadDir))))
	mux.Handle("/", h.frontendHandler())

	return securityHeaders(h.cors(mux))
}

func (h *HTTPHandler) handlePublicPosts(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		methodNotAllowed(w)
		return
	}

	posts, total, err := h.service.ListPublicPosts(r.Context(), data.ListPostsFilter{
		Page:         parseInt(r.URL.Query().Get("page"), 1),
		PageSize:     parseInt(r.URL.Query().Get("page_size"), 10),
		CategorySlug: strings.TrimSpace(r.URL.Query().Get("category")),
		TagSlug:      strings.TrimSpace(r.URL.Query().Get("tag")),
	})
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"data": mapPosts(posts, false),
		"meta": map[string]any{
			"page":      parseInt(r.URL.Query().Get("page"), 1),
			"page_size": parseInt(r.URL.Query().Get("page_size"), 10),
			"total":     total,
		},
	})
}

func (h *HTTPHandler) handlePublicPostDetail(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		methodNotAllowed(w)
		return
	}

	slug := strings.TrimPrefix(r.URL.Path, "/api/posts/")
	if slug == "" {
		writeError(w, http.StatusNotFound, service.ErrNotFound)
		return
	}

	post, err := h.service.GetPublicPost(r.Context(), slug)
	if err != nil {
		status := http.StatusInternalServerError
		if errors.Is(err, service.ErrNotFound) {
			status = http.StatusNotFound
		}
		writeError(w, status, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{"data": mapPost(post, false)})
}

func (h *HTTPHandler) handlePublicCategories(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		methodNotAllowed(w)
		return
	}
	items, err := h.service.ListCategories(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": items})
}

func (h *HTTPHandler) handlePublicTags(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		methodNotAllowed(w)
		return
	}
	items, err := h.service.ListTags(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": items})
}

func (h *HTTPHandler) handlePublicSite(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		methodNotAllowed(w)
		return
	}
	site, err := h.service.GetSiteSetting(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": site})
}

func (h *HTTPHandler) handleUserRegister(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		methodNotAllowed(w)
		return
	}

	var payload service.UserRegisterInput
	if err := decodeJSON(w, r, &payload); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	user, err := h.service.RegisterUser(r.Context(), payload)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrConflict):
			writeError(w, http.StatusConflict, errors.New("username or email already exists"))
		default:
			writeError(w, http.StatusBadRequest, err)
		}
		return
	}

	session, _ := h.sessions.Get(r, userSessionName)
	session.Values["user_id"] = user.ID
	if err := session.Save(r, w); err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}

	writeJSON(w, http.StatusCreated, map[string]any{"data": mapUser(user)})
}

func (h *HTTPHandler) handleUserLogin(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		methodNotAllowed(w)
		return
	}

	var payload service.UserLoginInput
	if err := decodeJSON(w, r, &payload); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	user, err := h.service.AuthenticateUser(r.Context(), payload)
	if err != nil {
		status := http.StatusUnauthorized
		if !errors.Is(err, service.ErrUnauthorized) {
			status = http.StatusInternalServerError
		}
		writeError(w, status, err)
		return
	}

	session, _ := h.sessions.Get(r, userSessionName)
	session.Values["user_id"] = user.ID
	if err := session.Save(r, w); err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{"data": mapUser(user)})
}

func (h *HTTPHandler) handleUserLogout(w http.ResponseWriter, r *http.Request, _ string) {
	if r.Method != http.MethodPost {
		methodNotAllowed(w)
		return
	}

	session, _ := h.sessions.Get(r, userSessionName)
	session.Options.MaxAge = -1
	if err := session.Save(r, w); err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": "ok"})
}

func (h *HTTPHandler) handleUserMe(w http.ResponseWriter, r *http.Request, userID string) {
	if r.Method != http.MethodGet {
		methodNotAllowed(w)
		return
	}

	user, err := h.service.CurrentUser(r.Context(), userID)
	if err != nil {
		status := http.StatusInternalServerError
		if errors.Is(err, service.ErrNotFound) {
			status = http.StatusUnauthorized
		}
		writeError(w, status, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": mapUser(user)})
}

func (h *HTTPHandler) handleAdminLogin(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		methodNotAllowed(w)
		return
	}
	if err := enforceStateChangingRequest(r); err != nil {
		writeError(w, http.StatusForbidden, err)
		return
	}

	var payload struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if err := decodeJSON(w, r, &payload); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	if !h.authLimiter.Allow(h.rateLimitKey(r, "admin-login:"+strings.ToLower(strings.TrimSpace(payload.Username))), 10, 15*time.Minute) {
		writeError(w, http.StatusTooManyRequests, errors.New("too many login attempts, please try again later"))
		return
	}

	admin, err := h.service.Authenticate(r.Context(), payload.Username, payload.Password)
	if err != nil {
		status := http.StatusUnauthorized
		if !errors.Is(err, service.ErrUnauthorized) {
			status = http.StatusInternalServerError
		}
		writeError(w, status, err)
		return
	}

	session, _ := h.sessions.Get(r, adminSessionName)
	session.Values["admin_id"] = int(admin.ID)
	if err := session.Save(r, w); err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{"data": admin})
}

func (h *HTTPHandler) handleAdminLogout(w http.ResponseWriter, r *http.Request, _ uint) {
	if r.Method != http.MethodPost {
		methodNotAllowed(w)
		return
	}
	if err := enforceStateChangingRequest(r); err != nil {
		writeError(w, http.StatusForbidden, err)
		return
	}

	session, _ := h.sessions.Get(r, adminSessionName)
	session.Options.MaxAge = -1
	if err := session.Save(r, w); err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": "ok"})
}

func (h *HTTPHandler) handleAdminMe(w http.ResponseWriter, r *http.Request, adminID uint) {
	if r.Method != http.MethodGet {
		methodNotAllowed(w)
		return
	}

	admin, err := h.service.CurrentAdmin(r.Context(), adminID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": admin})
}

func (h *HTTPHandler) handleAdminPosts(w http.ResponseWriter, r *http.Request, _ uint) {
	switch r.Method {
	case http.MethodGet:
		posts, total, err := h.service.ListAdminPosts(r.Context(), data.ListPostsFilter{
			Page:         parseInt(r.URL.Query().Get("page"), 1),
			PageSize:     parseInt(r.URL.Query().Get("page_size"), 20),
			Status:       strings.TrimSpace(r.URL.Query().Get("status")),
			CategorySlug: strings.TrimSpace(r.URL.Query().Get("category")),
			TagSlug:      strings.TrimSpace(r.URL.Query().Get("tag")),
		})
		if err != nil {
			writeError(w, http.StatusInternalServerError, err)
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{
			"data": mapPosts(posts, true),
			"meta": map[string]any{
				"page":      parseInt(r.URL.Query().Get("page"), 1),
				"page_size": parseInt(r.URL.Query().Get("page_size"), 20),
				"total":     total,
			},
		})
	case http.MethodPost:
		if err := enforceStateChangingRequest(r); err != nil {
			writeError(w, http.StatusForbidden, err)
			return
		}
		var payload service.PostInput
		if err := decodeJSON(w, r, &payload); err != nil {
			writeError(w, http.StatusBadRequest, err)
			return
		}
		post, err := h.service.CreatePost(r.Context(), payload)
		if err != nil {
			writeError(w, http.StatusBadRequest, err)
			return
		}
		writeJSON(w, http.StatusCreated, map[string]any{"data": mapPost(post, true)})
	default:
		methodNotAllowed(w)
	}
}

func (h *HTTPHandler) handleAdminPreview(w http.ResponseWriter, r *http.Request, _ uint) {
	if r.Method != http.MethodPost {
		methodNotAllowed(w)
		return
	}
	if err := enforceStateChangingRequest(r); err != nil {
		writeError(w, http.StatusForbidden, err)
		return
	}
	var payload struct {
		MarkdownContent string `json:"markdown_content"`
	}
	if err := decodeJSON(w, r, &payload); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"data": map[string]string{
			"html_content": h.service.RenderPreview(payload.MarkdownContent),
		},
	})
}

func (h *HTTPHandler) handleAdminPostItem(w http.ResponseWriter, r *http.Request, _ uint) {
	id, err := parseUintPath(r.URL.Path, "/api/admin/posts/")
	if err != nil {
		writeError(w, http.StatusNotFound, err)
		return
	}

	switch r.Method {
	case http.MethodGet:
		post, err := h.service.GetAdminPost(r.Context(), id)
		if err != nil {
			status := http.StatusInternalServerError
			if errors.Is(err, service.ErrNotFound) {
				status = http.StatusNotFound
			}
			writeError(w, status, err)
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"data": mapPost(post, true)})
	case http.MethodPut:
		if err := enforceStateChangingRequest(r); err != nil {
			writeError(w, http.StatusForbidden, err)
			return
		}
		var payload service.PostInput
		if err := decodeJSON(w, r, &payload); err != nil {
			writeError(w, http.StatusBadRequest, err)
			return
		}
		post, err := h.service.UpdatePost(r.Context(), id, payload)
		if err != nil {
			status := http.StatusBadRequest
			if errors.Is(err, service.ErrNotFound) {
				status = http.StatusNotFound
			}
			writeError(w, status, err)
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"data": mapPost(post, true)})
	case http.MethodDelete:
		if err := enforceStateChangingRequest(r); err != nil {
			writeError(w, http.StatusForbidden, err)
			return
		}
		if err := h.service.DeletePost(r.Context(), id); err != nil {
			writeError(w, http.StatusInternalServerError, err)
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"data": "ok"})
	default:
		methodNotAllowed(w)
	}
}

func (h *HTTPHandler) handleAdminCategories(w http.ResponseWriter, r *http.Request, _ uint) {
	switch r.Method {
	case http.MethodGet:
		items, err := h.service.ListCategories(r.Context())
		if err != nil {
			writeError(w, http.StatusInternalServerError, err)
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"data": items})
	case http.MethodPost:
		if err := enforceStateChangingRequest(r); err != nil {
			writeError(w, http.StatusForbidden, err)
			return
		}
		var payload service.CategoryInput
		if err := decodeJSON(w, r, &payload); err != nil {
			writeError(w, http.StatusBadRequest, err)
			return
		}
		item, err := h.service.CreateCategory(r.Context(), payload)
		if err != nil {
			writeError(w, http.StatusBadRequest, err)
			return
		}
		writeJSON(w, http.StatusCreated, map[string]any{"data": item})
	default:
		methodNotAllowed(w)
	}
}

func (h *HTTPHandler) handleAdminCategoryItem(w http.ResponseWriter, r *http.Request, _ uint) {
	id, err := parseUintPath(r.URL.Path, "/api/admin/categories/")
	if err != nil {
		writeError(w, http.StatusNotFound, err)
		return
	}
	switch r.Method {
	case http.MethodPut:
		if err := enforceStateChangingRequest(r); err != nil {
			writeError(w, http.StatusForbidden, err)
			return
		}
		var payload service.CategoryInput
		if err := decodeJSON(w, r, &payload); err != nil {
			writeError(w, http.StatusBadRequest, err)
			return
		}
		item, err := h.service.UpdateCategory(r.Context(), id, payload)
		if err != nil {
			status := http.StatusBadRequest
			if errors.Is(err, service.ErrNotFound) {
				status = http.StatusNotFound
			}
			writeError(w, status, err)
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"data": item})
	case http.MethodDelete:
		if err := enforceStateChangingRequest(r); err != nil {
			writeError(w, http.StatusForbidden, err)
			return
		}
		if err := h.service.DeleteCategory(r.Context(), id); err != nil {
			writeError(w, http.StatusInternalServerError, err)
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"data": "ok"})
	default:
		methodNotAllowed(w)
	}
}

func (h *HTTPHandler) handleAdminTags(w http.ResponseWriter, r *http.Request, _ uint) {
	switch r.Method {
	case http.MethodGet:
		items, err := h.service.ListTags(r.Context())
		if err != nil {
			writeError(w, http.StatusInternalServerError, err)
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"data": items})
	case http.MethodPost:
		if err := enforceStateChangingRequest(r); err != nil {
			writeError(w, http.StatusForbidden, err)
			return
		}
		var payload service.TagInput
		if err := decodeJSON(w, r, &payload); err != nil {
			writeError(w, http.StatusBadRequest, err)
			return
		}
		item, err := h.service.CreateTag(r.Context(), payload)
		if err != nil {
			writeError(w, http.StatusBadRequest, err)
			return
		}
		writeJSON(w, http.StatusCreated, map[string]any{"data": item})
	default:
		methodNotAllowed(w)
	}
}

func (h *HTTPHandler) handleAdminTagItem(w http.ResponseWriter, r *http.Request, _ uint) {
	id, err := parseUintPath(r.URL.Path, "/api/admin/tags/")
	if err != nil {
		writeError(w, http.StatusNotFound, err)
		return
	}
	switch r.Method {
	case http.MethodPut:
		if err := enforceStateChangingRequest(r); err != nil {
			writeError(w, http.StatusForbidden, err)
			return
		}
		var payload service.TagInput
		if err := decodeJSON(w, r, &payload); err != nil {
			writeError(w, http.StatusBadRequest, err)
			return
		}
		item, err := h.service.UpdateTag(r.Context(), id, payload)
		if err != nil {
			status := http.StatusBadRequest
			if errors.Is(err, service.ErrNotFound) {
				status = http.StatusNotFound
			}
			writeError(w, status, err)
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"data": item})
	case http.MethodDelete:
		if err := enforceStateChangingRequest(r); err != nil {
			writeError(w, http.StatusForbidden, err)
			return
		}
		if err := h.service.DeleteTag(r.Context(), id); err != nil {
			writeError(w, http.StatusInternalServerError, err)
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"data": "ok"})
	default:
		methodNotAllowed(w)
	}
}

func (h *HTTPHandler) handleAdminSite(w http.ResponseWriter, r *http.Request, _ uint) {
	switch r.Method {
	case http.MethodGet:
		site, err := h.service.GetSiteSetting(r.Context())
		if err != nil {
			writeError(w, http.StatusInternalServerError, err)
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"data": site})
	case http.MethodPut:
		if err := enforceStateChangingRequest(r); err != nil {
			writeError(w, http.StatusForbidden, err)
			return
		}
		var payload service.SiteSettingInput
		if err := decodeJSON(w, r, &payload); err != nil {
			writeError(w, http.StatusBadRequest, err)
			return
		}
		site, err := h.service.UpdateSiteSetting(r.Context(), payload)
		if err != nil {
			writeError(w, http.StatusBadRequest, err)
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"data": site})
	default:
		methodNotAllowed(w)
	}
}

func (h *HTTPHandler) handleAdminUpload(w http.ResponseWriter, r *http.Request, _ uint) {
	if r.Method != http.MethodPost {
		methodNotAllowed(w)
		return
	}
	if err := enforceStateChangingRequest(r); err != nil {
		writeError(w, http.StatusForbidden, err)
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, maxUploadBodySize+(1<<20))
	if err := r.ParseMultipartForm(maxUploadBodySize); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	if r.MultipartForm != nil {
		defer r.MultipartForm.RemoveAll()
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	path, err := h.service.SaveUpload(file, header)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}

	writeJSON(w, http.StatusCreated, map[string]any{"data": map[string]string{"path": path}})
}

func (h *HTTPHandler) handleAdminImportMarkdown(w http.ResponseWriter, r *http.Request, _ uint) {
	if r.Method != http.MethodPost {
		methodNotAllowed(w)
		return
	}
	if err := enforceStateChangingRequest(r); err != nil {
		writeError(w, http.StatusForbidden, err)
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, maxUploadBodyBytes)
	if err := r.ParseMultipartForm(5 << 20); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		writeError(w, http.StatusBadRequest, errors.New("markdown file is required"))
		return
	}

	categoryID, err := service.ParseOptionalUint(r.FormValue("category_id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	tagIDs, err := service.ParseUintCSV(r.FormValue("tag_ids"))
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	post, err := h.service.ImportMarkdownPost(r.Context(), file, header, service.MarkdownImportInput{
		Status:     strings.TrimSpace(r.FormValue("status")),
		CategoryID: categoryID,
		TagIDs:     tagIDs,
		CoverImage: strings.TrimSpace(r.FormValue("cover_image")),
	})
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	writeJSON(w, http.StatusCreated, map[string]any{"data": mapPost(post, true)})
}

func (h *HTTPHandler) handleUserRegister(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		methodNotAllowed(w)
		return
	}
	if err := enforceStateChangingRequest(r); err != nil {
		writeError(w, http.StatusForbidden, err)
		return
	}
	if !h.authLimiter.Allow(h.rateLimitKey(r, "user-register"), 5, 15*time.Minute) {
		writeError(w, http.StatusTooManyRequests, errors.New("too many registration attempts, please try again later"))
		return
	}

	var payload service.UserRegisterInput
	if err := decodeJSON(w, r, &payload, maxJSONBodyBytes); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	user, err := h.service.RegisterUser(r.Context(), payload)
	if err != nil {
		status := http.StatusBadRequest
		if errors.Is(err, service.ErrConflict) {
			status = http.StatusConflict
		}
		writeError(w, status, err)
		return
	}

	session, _ := h.sessions.Get(r, userSessionName)
	session.Values = map[interface{}]interface{}{}
	session.Values["user_id"] = int(user.ID)
	if err := session.Save(r, w); err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}

	writeJSON(w, http.StatusCreated, map[string]any{"data": mapUser(user)})
}

func (h *HTTPHandler) handleUserLogin(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		methodNotAllowed(w)
		return
	}
	if err := enforceStateChangingRequest(r); err != nil {
		writeError(w, http.StatusForbidden, err)
		return
	}

	var payload service.UserLoginInput
	if err := decodeJSON(w, r, &payload, maxJSONBodyBytes); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	if !h.authLimiter.Allow(h.rateLimitKey(r, "user-login:"+strings.ToLower(strings.TrimSpace(payload.Identifier))), 10, 15*time.Minute) {
		writeError(w, http.StatusTooManyRequests, errors.New("too many login attempts, please try again later"))
		return
	}

	user, err := h.service.AuthenticateUser(r.Context(), payload.Identifier, payload.Password)
	if err != nil {
		status := http.StatusUnauthorized
		if !errors.Is(err, service.ErrUnauthorized) {
			status = http.StatusInternalServerError
		}
		writeError(w, status, err)
		return
	}

	session, _ := h.sessions.Get(r, userSessionName)
	session.Values = map[interface{}]interface{}{}
	session.Values["user_id"] = int(user.ID)
	if err := session.Save(r, w); err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{"data": mapUser(user)})
}

func (h *HTTPHandler) handleUserLogout(w http.ResponseWriter, r *http.Request, _ uint) {
	if r.Method != http.MethodPost {
		methodNotAllowed(w)
		return
	}
	if err := enforceStateChangingRequest(r); err != nil {
		writeError(w, http.StatusForbidden, err)
		return
	}

	session, _ := h.sessions.Get(r, userSessionName)
	session.Options.MaxAge = -1
	if err := session.Save(r, w); err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{"data": "ok"})
}

func (h *HTTPHandler) handleUserMe(w http.ResponseWriter, r *http.Request, userID uint) {
	if r.Method != http.MethodGet {
		methodNotAllowed(w)
		return
	}
	user, err := h.service.CurrentUser(r.Context(), userID)
	if err != nil {
		status := http.StatusInternalServerError
		if errors.Is(err, service.ErrNotFound) {
			status = http.StatusNotFound
		}
		writeError(w, status, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": mapUser(user)})
}

func (h *HTTPHandler) handleSitemap(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		methodNotAllowed(w)
		return
	}

	posts, err := h.service.ListAllPublishedPosts(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}

	base := baseURL(r)
	entries := []sitemapURLEntry{{Loc: base}}
	for _, post := range posts {
		entry := sitemapURLEntry{
			Loc: fmt.Sprintf("%s/posts/%s", base, post.Slug),
		}
		if post.UpdatedAt.IsZero() {
			entry.LastMod = time.Now().Format(time.RFC3339)
		} else {
			entry.LastMod = post.UpdatedAt.Format(time.RFC3339)
		}
		entries = append(entries, entry)
	}

	w.Header().Set("Content-Type", "application/xml; charset=utf-8")
	_ = xml.NewEncoder(w).Encode(sitemapURLSet{
		Xmlns: "http://www.sitemaps.org/schemas/sitemap/0.9",
		URLs:  entries,
	})
}

func (h *HTTPHandler) frontendHandler() http.Handler {
	dist := h.cfg.App.FrontendDist
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/api/") || strings.HasPrefix(r.URL.Path, "/uploads/") || r.URL.Path == "/sitemap.xml" {
			http.NotFound(w, r)
			return
		}

		path := filepath.Join(dist, strings.TrimPrefix(filepath.Clean(r.URL.Path), "/"))
		if info, err := os.Stat(path); err == nil && !info.IsDir() {
			http.ServeFile(w, r, path)
			return
		}

		indexPath := filepath.Join(dist, "index.html")
		if _, err := os.Stat(indexPath); err == nil {
			http.ServeFile(w, r, indexPath)
			return
		}

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = w.Write([]byte("<!doctype html><html><body><h1>Frontend not built</h1><p>Run npm install && npm run build in /web.</p></body></html>"))
	})
}

func (h *HTTPHandler) requireAdmin(next func(http.ResponseWriter, *http.Request, uint)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		session, _ := h.sessions.Get(r, adminSessionName)
		rawID, ok := session.Values["admin_id"]
		if !ok {
			writeError(w, http.StatusUnauthorized, errors.New("authentication required"))
			return
		}

		adminID, ok := normalizeSessionID(rawID)
		if !ok {
			writeError(w, http.StatusUnauthorized, errors.New("invalid session"))
			return
		}

		next(w, r, adminID)
	}
}

func (h *HTTPHandler) requireUser(next func(http.ResponseWriter, *http.Request, string)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		session, _ := h.sessions.Get(r, userSessionName)
		rawID, ok := session.Values["user_id"]
		if !ok {
			writeError(w, http.StatusUnauthorized, errors.New("authentication required"))
			return
		}

		userID, ok := normalizeSessionString(rawID)
		if !ok || userID == "" {
			writeError(w, http.StatusUnauthorized, errors.New("invalid session"))
			return
		}

		next(w, r, userID)
	}
}

func decodeJSON(w http.ResponseWriter, r *http.Request, dest any) error {
	r.Body = http.MaxBytesReader(w, r.Body, jsonBodyLimit)
	defer r.Body.Close()
	r.Body = http.MaxBytesReader(w, r.Body, maxBytes)
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(dest); err != nil {
		return err
	}

	var extra struct{}
	if err := decoder.Decode(&extra); err != io.EOF {
		return errors.New("request body must contain a single JSON object")
	}
	return nil
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

func writeError(w http.ResponseWriter, status int, err error) {
	writeJSON(w, status, apiError{Error: err.Error()})
}

func methodNotAllowed(w http.ResponseWriter) {
	writeError(w, http.StatusMethodNotAllowed, errors.New("method not allowed"))
}

func parseUintPath(path, prefix string) (uint, error) {
	value := strings.TrimPrefix(path, prefix)
	value = strings.Trim(value, "/")
	id, err := strconv.ParseUint(value, 10, 64)
	if err != nil {
		return 0, service.ErrNotFound
	}
	return uint(id), nil
}

func parseInt(raw string, fallback int) int {
	value, err := strconv.Atoi(raw)
	if err != nil || value <= 0 {
		return fallback
	}
	return value
}

func normalizeSessionID(raw any) (uint, bool) {
	switch value := raw.(type) {
	case uint:
		return value, true
	case int:
		return uint(value), true
	case int64:
		return uint(value), true
	case float64:
		return uint(value), true
	default:
		return 0, false
	}
}

func normalizeSessionString(raw any) (string, bool) {
	switch value := raw.(type) {
	case string:
		return value, true
	case []byte:
		return string(value), true
	default:
		return "", false
	}
}

func mapPosts(posts []data.Post, includeMarkdown bool) []map[string]any {
	items := make([]map[string]any, 0, len(posts))
	for i := range posts {
		post := posts[i]
		items = append(items, mapPost(&post, includeMarkdown))
	}
	return items
}

func mapPost(post *data.Post, includeMarkdown bool) map[string]any {
	payload := map[string]any{
		"id":           post.ID,
		"title":        post.Title,
		"slug":         post.Slug,
		"summary":      post.Summary,
		"cover_image":  post.CoverImage,
		"html_content": post.HTMLContent,
		"status":       post.Status,
		"published_at": post.PublishedAt,
		"created_at":   post.CreatedAt,
		"updated_at":   post.UpdatedAt,
		"category_id":  post.CategoryID,
		"category":     post.Category,
		"tags":         post.Tags,
	}
	if includeMarkdown {
		payload["markdown_content"] = post.MarkdownContent
	}
	return payload
}

func mapUser(user *data.User) map[string]any {
	return map[string]any{
		"id":         user.ID,
		"username":   user.Username,
		"email":      user.Email,
		"created_at": user.CreatedAt,
		"updated_at": user.UpdatedAt,
	}
}

func baseURL(r *http.Request) string {
	scheme := "http"
	if r.TLS != nil {
		scheme = "https"
	}
	if forwarded := r.Header.Get("X-Forwarded-Proto"); forwarded != "" {
		scheme = forwarded
	}
	return fmt.Sprintf("%s://%s", scheme, r.Host)
}

func (h *HTTPHandler) cors(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")

		if value := origin(r); value != "" {
			w.Header().Set("Access-Control-Allow-Origin", value)
			w.Header().Set("Access-Control-Allow-Credentials", "true")
		}
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, X-Requested-With")
		w.Header().Set("Access-Control-Allow-Methods", "GET,POST,PUT,DELETE,OPTIONS")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func origin(r *http.Request) string {
	value := strings.TrimSpace(r.Header.Get("Origin"))
	if value == "" {
		return ""
	}

	u, err := url.Parse(value)
	if err != nil {
		return ""
	}
	if strings.EqualFold(u.Host, r.Host) {
		return value
	}
	if u.Host == "localhost:5173" || u.Host == "127.0.0.1:5173" {
		return value
	}

	return ""
}

func newAuthRateLimiter() *authRateLimiter {
	return &authRateLimiter{
		windows: map[string]authWindow{},
	}
}

func (l *authRateLimiter) Allow(key string, limit int, window time.Duration) bool {
	l.mu.Lock()
	defer l.mu.Unlock()

	now := time.Now()
	current, ok := l.windows[key]
	if !ok || now.After(current.ResetAt) {
		l.windows[key] = authWindow{
			Count:   1,
			ResetAt: now.Add(window),
		}
		return true
	}
	if current.Count >= limit {
		return false
	}
	current.Count++
	l.windows[key] = current
	return true
}

func (h *HTTPHandler) rateLimitKey(r *http.Request, prefix string) string {
	return prefix + ":" + clientIP(r)
}

func clientIP(r *http.Request) string {
	host := r.RemoteAddr
	if parsedHost, _, err := net.SplitHostPort(host); err == nil {
		return parsedHost
	}
	return host
}

func enforceStateChangingRequest(r *http.Request) error {
	switch r.Method {
	case http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete:
		if r.Header.Get(csrfHeaderName) != csrfHeaderValue {
			return errors.New("missing required csrf request header")
		}
	}
	return nil
}

func securityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
		w.Header().Set("Cross-Origin-Opener-Policy", "same-origin")
		next.ServeHTTP(w, r)
	})
}
