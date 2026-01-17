package server

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

// StaticFileHandler 静态文件处理器
type StaticFileHandler struct {
	webRoot  string
	mode     string
	baseDomain string
	singleDomain string
}

// NewStaticFileHandler 创建静态文件处理器
func NewStaticFileHandler(webRoot, mode, baseDomain, singleDomain string) *StaticFileHandler {
	return &StaticFileHandler{
		webRoot:  webRoot,
		mode:     mode,
		baseDomain: baseDomain,
		singleDomain: singleDomain,
	}
}

// ServeHTTP 处理HTTP请求
func (h *StaticFileHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// 获取网站名称
	siteName, err := h.extractSiteName(r.Host)
	if err != nil {
		http.Error(w, "Website not found", http.StatusNotFound)
		return
	}

	// 构建网站路径
	sitePath := filepath.Join(h.webRoot, siteName)

	// 检查网站是否存在
	if _, err := os.Stat(sitePath); os.IsNotExist(err) {
		http.Error(w, "Website not found", http.StatusNotFound)
		return
	}

	// 处理请求路径
	requestPath := r.URL.Path

	// 如果是根路径，尝试 index.html
	if requestPath == "/" {
		requestPath = "/index.html"
	}

	// 构建完整文件路径
	filePath := filepath.Join(sitePath, requestPath)

	// 清理路径，防止目录遍历攻击
	filePath = filepath.Clean(filePath)
	if !strings.HasPrefix(filePath, sitePath) {
		http.Error(w, "Access denied", http.StatusForbidden)
		return
	}

	// 检查文件是否存在
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		// 尝试返回 index.html (SPA 路由支持)
		indexPath := filepath.Join(sitePath, "index.html")
		if _, err := os.Stat(indexPath); err == nil {
			h.serveFile(w, r, indexPath)
			return
		}
		http.Error(w, "File not found", http.StatusNotFound)
		return
	}

	// 服务文件
	h.serveFile(w, r, filePath)
}

// serveFile 服务单个文件
func (h *StaticFileHandler) serveFile(w http.ResponseWriter, r *http.Request, filePath string) {
	// 获取文件信息
	info, err := os.Stat(filePath)
	if err != nil {
		http.Error(w, "File not found", http.StatusNotFound)
		return
	}

	// 如果是目录，尝试 index.html
	if info.IsDir() {
		indexPath := filepath.Join(filePath, "index.html")
		if _, err := os.Stat(indexPath); err == nil {
			h.serveFile(w, r, indexPath)
			return
		}
		http.Error(w, "Directory listing not allowed", http.StatusForbidden)
		return
	}

	// 打开文件
	file, err := os.Open(filePath)
	if err != nil {
		http.Error(w, "Failed to open file", http.StatusInternalServerError)
		return
	}
	defer file.Close()

	// 设置 Content-Type
	contentType := h.getContentType(filePath)
	w.Header().Set("Content-Type", contentType)

	// 设置缓存头
	if h.shouldCache(filePath) {
		w.Header().Set("Cache-Control", "public, max-age=31536000") // 1年
	} else {
		w.Header().Set("Cache-Control", "no-cache")
	}

	// 设置 ETag
	etag := fmt.Sprintf(`"%x"`, info.ModTime().Unix())
	w.Header().Set("ETag", etag)

	// 检查 If-None-Match
	if r.Header.Get("If-None-Match") == etag {
		w.WriteHeader(http.StatusNotModified)
		return
	}

	// 返回文件内容
	http.ServeContent(w, r, filePath, info.ModTime(), file)
}

// extractSiteName 从请求中提取网站名称
func (h *StaticFileHandler) extractSiteName(host string) (string, error) {
	if h.mode == "subdomain" {
		// 子域名模式: site.example.com -> site
		parts := strings.Split(host, ".")
		if len(parts) >= 2 {
			// 检查是否匹配基础域名
			domain := strings.Join(parts[len(parts)-2:], ".")
			if domain == h.baseDomain && len(parts) > 2 {
				return parts[0], nil
			}
		}
		return "", fmt.Errorf("invalid subdomain")
	} else {
		// 路径模式: example.com/site -> site
		// 这个在 ServeHTTP 中通过 r.URL.Path 处理
		// 这里返回默认网站
		return "", fmt.Errorf("use path mode")
	}
}

// getContentType 根据文件扩展名获取 Content-Type
func (h *StaticFileHandler) getContentType(filePath string) string {
	ext := strings.ToLower(filepath.Ext(filePath))
	switch ext {
	case ".html":
		return "text/html; charset=utf-8"
	case ".css":
		return "text/css; charset=utf-8"
	case ".js":
		return "application/javascript; charset=utf-8"
	case ".json":
		return "application/json; charset=utf-8"
	case ".png":
		return "image/png"
	case ".jpg", ".jpeg":
		return "image/jpeg"
	case ".gif":
		return "image/gif"
	case ".svg":
		return "image/svg+xml"
	case ".ico":
		return "image/x-icon"
	case ".woff":
		return "font/woff"
	case ".woff2":
		return "font/woff2"
	case ".ttf":
		return "font/ttf"
	case ".eot":
		return "application/vnd.ms-fontobject"
	case ".pdf":
		return "application/pdf"
	case ".xml":
		return "application/xml; charset=utf-8"
	default:
		return "application/octet-stream"
	}
}

// shouldCache 判断文件是否应该缓存
func (h *StaticFileHandler) shouldCache(filePath string) bool {
	ext := strings.ToLower(filepath.Ext(filePath))
	cacheableExts := map[string]bool{
		".js":  true,
		".css": true,
		".png": true,
		".jpg": true,
		".jpeg": true,
		".gif": true,
		".svg": true,
		".ico": true,
		".woff": true,
		".woff2": true,
		".ttf": true,
		".eot": true,
	}
	return cacheableExts[ext]
}

// PathModeHandler 路径模式的处理器
type PathModeHandler struct {
	*StaticFileHandler
}

// ServeHTTP 路径模式的HTTP处理
func (h *PathModeHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// 从路径中提取网站名称: /site/path -> site
	path := strings.TrimPrefix(r.URL.Path, "/")
	parts := strings.SplitN(path, "/", 2)

	siteName := parts[0]
	if siteName == "" {
		// 列出所有网站
		h.listSites(w, r)
		return
	}

	// 构建网站路径
	sitePath := filepath.Join(h.webRoot, siteName)

	// 检查网站是否存在
	if _, err := os.Stat(sitePath); os.IsNotExist(err) {
		http.Error(w, "Website not found", http.StatusNotFound)
		return
	}

	// 获取文件路径
	var requestPath string
	if len(parts) > 1 {
		requestPath = "/" + parts[1]
	} else {
		requestPath = "/"
	}

	// 构建完整文件路径
	filePath := filepath.Join(sitePath, requestPath)

	// 清理路径
	filePath = filepath.Clean(filePath)
	if !strings.HasPrefix(filePath, sitePath) {
		http.Error(w, "Access denied", http.StatusForbidden)
		return
	}

	// 如果是根路径，尝试 index.html
	if requestPath == "/" {
		filePath = filepath.Join(sitePath, "index.html")
	}

	// 检查文件是否存在
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		// 尝试返回 index.html (SPA 路由支持)
		indexPath := filepath.Join(sitePath, "index.html")
		if _, err := os.Stat(indexPath); err == nil {
			h.serveFile(w, r, indexPath)
			return
		}
		http.Error(w, "File not found", http.StatusNotFound)
		return
	}

	// 服务文件
	h.serveFile(w, r, filePath)
}

// listSites 列出所有网站
func (h *PathModeHandler) listSites(w http.ResponseWriter, r *http.Request) {
	entries, err := os.ReadDir(h.webRoot)
	if err != nil {
		http.Error(w, "Failed to read websites", http.StatusInternalServerError)
		return
	}

	sites := make([]string, 0)
	for _, entry := range entries {
		if entry.IsDir() {
			sites = append(sites, entry.Name())
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"sites": sites,
	})
}
