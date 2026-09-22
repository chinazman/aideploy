package server

import (
	"encoding/json"
	"fmt"
	"html"
	"log"
	"net/http"
	"net/url"
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
		log.Printf("[ERROR] Failed to extract site name from host=%s: %v", r.Host, err)
		http.Error(w, "Website not found", http.StatusNotFound)
		return
	}

	// 构建网站路径
	sitePath := filepath.Join(h.webRoot, siteName)

	// 检查网站是否存在
	if _, err := os.Stat(sitePath); os.IsNotExist(err) {
		log.Printf("[ERROR] Site directory not found: %s", sitePath)
		http.Error(w, "Website not found", http.StatusNotFound)
		return
	}

	// 处理请求路径
	requestPath := r.URL.Path

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
		if indexPath, ok := findIndexFile(sitePath); ok {
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

	// 如果是目录，尝试 index.html / index.htm
	if info.IsDir() {
		if indexPath, ok := findIndexFile(filePath); ok {
			h.serveFile(w, r, indexPath)
			return
		}
		// 没有索引文件，列出目录内容
		h.serveDirList(w, r, filePath)
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

// findIndexFile 在目录中查找 index.html 或 index.htm
func findIndexFile(dir string) (string, bool) {
	for _, name := range []string{"index.html", "index.htm"} {
		p := filepath.Join(dir, name)
		if info, err := os.Stat(p); err == nil && !info.IsDir() {
			return p, true
		}
	}
	return "", false
}

// serveDirList 列出目录下的所有文件和子目录，并生成可点击的 HTML 页面
func (h *StaticFileHandler) serveDirList(w http.ResponseWriter, r *http.Request, dirPath string) {
	// 保证目录 URL 以 / 结尾，方便构造相对链接
	if !strings.HasSuffix(r.URL.Path, "/") {
		http.Redirect(w, r, r.URL.Path+"/", http.StatusMovedPermanently)
		return
	}

	entries, err := os.ReadDir(dirPath)
	if err != nil {
		http.Error(w, "Failed to read directory", http.StatusInternalServerError)
		return
	}

	basePath := strings.TrimSuffix(r.URL.Path, "/")

	var buf strings.Builder
	buf.WriteString(`<!DOCTYPE html>
<html lang="zh-CN">
<head>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width, initial-scale=1">
<title>目录列表</title>
<style>
body { font-family: -apple-system, "Segoe UI", "Microsoft YaHei", sans-serif; margin: 2rem; color: #333; }
h1 { font-size: 1.2rem; word-break: break-all; }
ul { list-style: none; padding: 0; max-width: 800px; }
li { padding: 0.4rem 0.5rem; border-bottom: 1px solid #eee; display: flex; justify-content: space-between; gap: 1rem; }
li:hover { background: #f7f7f7; }
a { text-decoration: none; color: #0366d6; word-break: break-all; }
a:hover { text-decoration: underline; }
.meta { color: #888; font-size: 0.85rem; white-space: nowrap; }
.dir { font-weight: bold; }
</style>
</head>
<body>
<h1>`)
	buf.WriteString(html.EscapeString(r.URL.Path))
	buf.WriteString("</h1>\n<ul>\n")

	// 上级目录链接（站点根路径不显示）
	if basePath != "" {
		parent := basePath[:strings.LastIndex(basePath, "/")+1]
		fmt.Fprintf(&buf, `<li><a href="%s">&#8592; 上级目录</a></li>`+"\n",
			html.EscapeString(parent))
	}

	for _, entry := range entries {
		name := entry.Name()
		href := basePath + "/" + url.PathEscape(name)
		display := name
		meta := ""
		if entry.IsDir() {
			href += "/"
			display += "/"
			buf.WriteString(fmt.Sprintf(`<li><a class="dir" href="%s">%s</a><span class="meta">目录</span></li>`+"\n",
				html.EscapeString(href), html.EscapeString(display)))
			continue
		}
		if info, err := entry.Info(); err == nil {
			meta = fmt.Sprintf("%s  %s", formatSize(info.Size()), info.ModTime().Format("2006-01-02 15:04"))
		}
		fmt.Fprintf(&buf, `<li><a href="%s">%s</a><span class="meta">%s</span></li>`+"\n",
			html.EscapeString(href), html.EscapeString(display), html.EscapeString(meta))
	}

	buf.WriteString("</ul>\n</body>\n</html>\n")

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Cache-Control", "no-cache")
	w.Write([]byte(buf.String()))
}

// formatSize 格式化文件大小
func formatSize(size int64) string {
	const unit = 1024
	if size < unit {
		return fmt.Sprintf("%d B", size)
	}
	div, exp := int64(unit), 0
	for n := size / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(size)/float64(div), "KMGTPE"[exp])
}

// extractSiteName 从请求中提取网站名称
func (h *StaticFileHandler) extractSiteName(host string) (string, error) {
	if h.mode == "subdomain" {
		// 移除端口号（如果存在）
		hostParts := strings.Split(host, ":")
		host = hostParts[0]

		// 子域名模式
		// 支持两种格式:
		// 1. site.example.com -> site (二级域名)
		// 2. site.tp.example.com -> site (三级域名)
		parts := strings.Split(host, ".")

		if len(parts) >= 2 {
			// 检查是否匹配基础域名
			// 计算基础域名包含几个部分
			baseDomainParts := strings.Split(h.baseDomain, ".")
			baseDomainPartCount := len(baseDomainParts)

			if len(parts) < baseDomainPartCount + 1 {
				return "", fmt.Errorf("host too short: host=%s, baseDomain=%s", host, h.baseDomain)
			}

			// 提取实际域名部分（从host中去除站点名部分）
			// 例如: jydemo.tp.fornewtech.com，baseDomain是 tp.fornewtech.com
			// 则实际域名应该是 parts[1:] = [tp, fornewtech, com] -> "tp.fornewtech.com"
			actualDomain := strings.Join(parts[1:], ".")

			if actualDomain == h.baseDomain {
				siteName := parts[0]
				return siteName, nil
			}
		}

		return "", fmt.Errorf("invalid subdomain: host=%s, baseDomain=%s", host, h.baseDomain)
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

	// 检查文件是否存在
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		// 尝试返回 index.html (SPA 路由支持)
		if indexPath, ok := findIndexFile(sitePath); ok {
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
