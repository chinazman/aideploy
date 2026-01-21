package server

import (
	"archive/tar"
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// 上下文键类型
type contextKey string

const userContextKey contextKey = "user"

// Config 服务器配置
type Config struct {
	BaseDomain       string            `json:"base_domain"`       // 基础域名（子域名模式）
	WebRoot          string            `json:"web_root"`          // 网站根目录
	Mode             string            `json:"mode"`              // 部署模式：subdomain 或 path
	SingleDomain     string            `json:"single_domain"`     // 单域名模式下的域名
	Port             int               `json:"port"`              // 服务器端口
	EnableVersioning bool              `json:"enable_versioning"` // 是否启用版本控制
	APIKey           string            `json:"api_key,omitempty"` // API密钥（已弃用，保留兼容）
	Sites            map[string]Site   `json:"sites"`             // 网站配置
	Users            map[string]User   `json:"users"`             // 用户配置
}

// Site 网站配置
type Site struct {
	Name string   `json:"name"` // 网站名称
	Desc string   `json:"desc"` // 网站描述
	Owner string   `json:"owner"` // 所有者
	Users []string `json:"users"` // 授权用户列表
}

// User 用户配置
type User struct {
	Name     string `json:"name"`     // 用户名
	Password string `json:"pass"`     // 密码
	IsAdmin  bool   `json:"isAdmin"` // 是否是管理员
}

// Website 网站信息
type Website struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Domain      string    `json:"domain"`
	Path        string    `json:"path"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// Version 版本信息
type Version struct {
	Hash        string    `json:"hash"`
	Message     string    `json:"message"`
	Author      string    `json:"author"`
	Date        time.Time `json:"date"`
}

// DeployServer 部署服务器
type DeployServer struct {
	config         Config
	sites          map[string]*Website
	mu             sync.RWMutex
	configPath     string // 配置文件路径
}

// NewDeployServer 创建新的部署服务器
func NewDeployServer(config Config, configPath string) *DeployServer {
	s := &DeployServer{
		config:     config,
		sites:      make(map[string]*Website),
		configPath: configPath,
	}
	// 从配置文件加载网站信息
	s.reloadSitesFromConfig()
	return s
}

// saveConfig 保存配置到文件
func (s *DeployServer) saveConfig() error {
	data, err := json.MarshalIndent(s.config, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(s.configPath, data, 0644)
}

// reloadSitesFromConfig 从配置重新加载网站信息到内存
func (s *DeployServer) reloadSitesFromConfig() {
	// 清空现有的内存缓存
	s.sites = make(map[string]*Website)

	// 从配置中重新加载
	for name := range s.config.Sites {
		sitePath := filepath.Join(s.config.WebRoot, name)

		var domain string
		if s.config.Mode == "subdomain" {
			domain = fmt.Sprintf("%s.%s", name, s.config.BaseDomain)
		} else {
			domain = fmt.Sprintf("%s/%s", s.config.SingleDomain, name)
		}

		// 尝试读取文件信息获取时间戳
		var created, updated time.Time
		if info, err := os.Stat(sitePath); err == nil {
			created = info.ModTime()
			updated = info.ModTime()
		} else {
			created = time.Now()
			updated = time.Now()
		}

		s.sites[name] = &Website{
			ID:        name,
			Name:      name,
			Domain:    domain,
			Path:      sitePath,
			CreatedAt: created,
			UpdatedAt: updated,
		}
	}
}

// authenticate 用户认证
func (s *DeployServer) authenticate(username, password string) (*User, error) {
	user, exists := s.config.Users[username]
	if !exists {
		return nil, fmt.Errorf("用户不存在")
	}

	if user.Password != password {
		return nil, fmt.Errorf("密码错误")
	}

	return &user, nil
}

// canAccessSite 检查用户是否有权限访问网站
func (s *DeployServer) canAccessSite(siteName, username string) bool {
	site, exists := s.config.Sites[siteName]
	if !exists {
		return false
	}

	// 检查是否是所有者
	if site.Owner == username {
		return true
	}

	// 检查是否在授权用户列表中
	for _, u := range site.Users {
		if u == username {
			return true
		}
	}

	return false
}

// isSiteOwner 检查用户是否是网站的所有者
func (s *DeployServer) isSiteOwner(siteName, username string) bool {
	site, exists := s.config.Sites[siteName]
	if !exists {
		return false
	}
	return site.Owner == username
}

// contextWithUser 将用户信息添加到上下文
func contextWithUser(ctx context.Context, user *User) context.Context {
	return context.WithValue(ctx, userContextKey, user)
}

// userFromContext 从上下文中获取用户信息
func userFromContext(ctx context.Context) *User {
	if user, ok := ctx.Value(userContextKey).(*User); ok {
		return user
	}
	return nil
}

// requireAdmin 要求管理员权限的中间件
func (s *DeployServer) requireAdmin(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user := userFromContext(r.Context())
		if user == nil || !user.IsAdmin {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusForbidden)
			json.NewEncoder(w).Encode(map[string]string{
				"error": "需要管理员权限",
			})
			return
		}
		next(w, r)
	}
}

// Start 启动服务器
func (s *DeployServer) Start() error {
	// 确保web根目录存在
	if err := os.MkdirAll(s.config.WebRoot, 0755); err != nil {
		return fmt.Errorf("创建web根目录失败: %v", err)
	}

	// 创建API路由
	mux := http.NewServeMux()

	// API路由
	mux.HandleFunc("/api/sites", s.corsMiddleware(s.authMiddleware(s.handleSites)))
	mux.HandleFunc("/api/sites/create", s.corsMiddleware(s.authMiddleware(s.handleCreateSite)))
	mux.HandleFunc("/api/sites/update", s.corsMiddleware(s.authMiddleware(s.handleUpdateSite)))
	mux.HandleFunc("/api/sites/delete", s.corsMiddleware(s.authMiddleware(s.handleDeleteSite)))
	mux.HandleFunc("/api/sites/deploy", s.corsMiddleware(s.authMiddleware(s.handleDeploy)))
	mux.HandleFunc("/api/sites/deploy-full", s.corsMiddleware(s.authMiddleware(s.handleDeployFull))) // 全量部署
	mux.HandleFunc("/api/sites/deploy-incremental", s.corsMiddleware(s.authMiddleware(s.handleDeployIncremental))) // 增量部署
	mux.HandleFunc("/api/sites/versions", s.corsMiddleware(s.authMiddleware(s.handleVersions)))
	mux.HandleFunc("/api/sites/rollback", s.corsMiddleware(s.authMiddleware(s.handleRollback)))
	mux.HandleFunc("/api/sites/list", s.corsMiddleware(s.authMiddleware(s.handleListSites)))
	mux.HandleFunc("/api/sites/export", s.corsMiddleware(s.authMiddleware(s.handleExport)))

	// 用户管理路由（需要管理员权限）
	mux.HandleFunc("/api/users/list", s.corsMiddleware(s.authMiddleware(s.requireAdmin(s.handleListUsers))))
	mux.HandleFunc("/api/users/create", s.corsMiddleware(s.authMiddleware(s.requireAdmin(s.handleCreateUser))))
	mux.HandleFunc("/api/users/update", s.corsMiddleware(s.authMiddleware(s.requireAdmin(s.handleUpdateUser))))
	mux.HandleFunc("/api/users/delete", s.corsMiddleware(s.authMiddleware(s.requireAdmin(s.handleDeleteUser))))

	// 网站授权路由
	mux.HandleFunc("/api/sites/authorize", s.corsMiddleware(s.authMiddleware(s.handleAuthorizeSite)))
	mux.HandleFunc("/api/sites/unauthorize", s.corsMiddleware(s.authMiddleware(s.handleUnauthorizeSite)))

	// 创建静态文件处理器
	var staticHandler http.Handler
	if s.config.Mode == "subdomain" {
		staticHandler = NewStaticFileHandler(s.config.WebRoot, s.config.Mode, s.config.BaseDomain, s.config.SingleDomain)
	} else {
		staticHandler = &PathModeHandler{
			StaticFileHandler: NewStaticFileHandler(s.config.WebRoot, s.config.Mode, s.config.BaseDomain, s.config.SingleDomain),
		}
	}

	// 创建最终处理器
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// API请求使用mux处理
		if strings.HasPrefix(r.URL.Path, "/api/") {
			mux.ServeHTTP(w, r)
			return
		}

		// 其他请求使用静态文件处理器
		staticHandler.ServeHTTP(w, r)
	})

	addr := fmt.Sprintf(":%d", s.config.Port)
	fmt.Printf("服务器启动在 http://localhost%s\n", addr)
	fmt.Printf("部署模式: %s, 基础域名: %s\n", s.config.Mode, s.config.BaseDomain)
	if s.config.Mode == "subdomain" {
		fmt.Printf("访问格式: http://site-name.%s\n", s.config.BaseDomain)
	} else {
		fmt.Printf("访问格式: http://%s/site-name\n", s.config.SingleDomain)
	}

	return http.ListenAndServe(addr, handler)
}

// corsMiddleware CORS中间件
func (s *DeployServer) corsMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS, DELETE")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, X-API-Key, X-Username, X-Password")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next(w, r)
	}
}

// getUserFromRequest 从请求中获取用户信息
func (s *DeployServer) getUserFromRequest(r *http.Request) (*User, error) {
	// 优先使用新的用户名/密码认证
	username := r.Header.Get("X-Username")
	password := r.Header.Get("X-Password")

	if username != "" && password != "" {
		return s.authenticate(username, password)
	}

	// 兼容旧的 API Key 认证
	if s.config.APIKey != "" {
		providedKey := r.Header.Get("X-API-Key")
		if providedKey == s.config.APIKey {
			// API Key 认证通过，返回管理员权限
			return &User{
				Name:    "admin",
				IsAdmin: true,
			}, nil
		}
		return nil, fmt.Errorf("无效的API密钥")
	}

	// 如果都没有配置，返回错误
	return nil, fmt.Errorf("未提供认证信息")
}

// authMiddleware 用户认证中间件
func (s *DeployServer) authMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// 如果配置了用户系统，使用用户认证
		if len(s.config.Users) > 0 {
			user, err := s.getUserFromRequest(r)
			if err != nil {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusUnauthorized)
				json.NewEncoder(w).Encode(map[string]string{
					"error": "未授权：" + err.Error(),
				})
				return
			}

			// 将用户信息存储到请求上下文中
			ctx := contextWithUser(r.Context(), user)
			next(w, r.WithContext(ctx))
			return
		}

		// 兼容旧的 API Key 认证
		if s.config.APIKey != "" {
			providedKey := r.Header.Get("X-API-Key")
			if providedKey != s.config.APIKey {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusUnauthorized)
				json.NewEncoder(w).Encode(map[string]string{
					"error": "未授权：无效的API密钥",
				})
				return
			}
		}

		// 如果没有配置认证，继续处理请求
		next(w, r)
	}
}

// handleSites 处理网站列表（单个网站信息）
func (s *DeployServer) handleSites(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		s.respondError(w, "方法不允许", http.StatusMethodNotAllowed)
		return
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	sites := make([]*Website, 0, len(s.sites))
	for _, site := range s.sites {
		sites = append(sites, site)
	}

	s.respondJSON(w, sites)
}

// handleListSites 列出所有网站
func (s *DeployServer) handleListSites(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		s.respondError(w, "方法不允许", http.StatusMethodNotAllowed)
		return
	}

	entries, err := os.ReadDir(s.config.WebRoot)
	if err != nil {
		s.respondError(w, fmt.Sprintf("读取目录失败: %v", err), http.StatusInternalServerError)
		return
	}

	// 获取当前用户
	user := userFromContext(r.Context())

	// 获取请求的协议和主机
	scheme := "http"
	if r.TLS != nil {
		scheme = "https"
	}
	host := r.Host

	// 构建网站信息列表
	type SiteInfo struct {
		Name   string `json:"name"`
		Domain string `json:"domain"`
		Desc   string `json:"desc"`
		URL    string `json:"url"`
	}

	sites := []SiteInfo{}
	for _, entry := range entries {
		if entry.IsDir() {
			siteName := entry.Name()

			// 如果配置了用户系统，进行权限过滤
			if user != nil {
				// 检查用户是否有权限访问此网站
				if !s.canAccessSite(siteName, user.Name) {
					continue
				}
			}

			// 从配置中获取网站信息
			siteConfig, exists := s.config.Sites[siteName]
			var desc string
			if exists {
				desc = siteConfig.Desc
			}

			var siteURL string
			var domain string

			// 根据模式生成 URL
			if s.config.Mode == "subdomain" {
				// 子域名模式: siteName.baseDomain
				domain = fmt.Sprintf("%s.%s", siteName, s.config.BaseDomain)
				siteURL = fmt.Sprintf("%s://%s", scheme, domain)
				// 如果有端口，添加端口
				if s.config.Port != 80 && s.config.Port != 443 {
					siteURL = fmt.Sprintf("%s:%d", siteURL, s.config.Port)
					domain = fmt.Sprintf("%s:%d", domain, s.config.Port)
				}
			} else {
				// 路径模式: host/siteName
				domain = fmt.Sprintf("%s/%s", s.config.BaseDomain, siteName)
				siteURL = fmt.Sprintf("%s://%s", scheme, host)
				if s.config.Port != 80 && s.config.Port != 443 {
					siteURL = fmt.Sprintf("%s:%d", siteURL, s.config.Port)
				}
			}

			sites = append(sites, SiteInfo{
				Name:   siteName,
				Domain: domain,
				Desc:   desc,
				URL:    siteURL,
			})
		}
	}

	s.respondJSON(w, map[string]interface{}{
		"sites": sites,
	})
}

// handleCreateSite 创建网站
func (s *DeployServer) handleCreateSite(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		s.respondError(w, "方法不允许", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Name string `json:"name"`
		Desc string `json:"desc"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.respondError(w, "无效的请求", http.StatusBadRequest)
		return
	}

	if req.Name == "" {
		s.respondError(w, "网站名称不能为空", http.StatusBadRequest)
		return
	}

	// 获取当前用户
	user := userFromContext(r.Context())
	if user == nil {
		s.respondError(w, "未授权", http.StatusUnauthorized)
		return
	}

	// 清理名称，只保留字母数字和连字符
	name := strings.TrimSpace(req.Name)
	name = strings.ToLower(name)
	name = strings.ReplaceAll(name, " ", "-")

	// 移除不安全字符
	var safeName strings.Builder
	for _, r := range name {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-' || r == '_' {
			safeName.WriteRune(r)
		}
	}
	name = safeName.String()

	if name == "" {
		s.respondError(w, "网站名称格式不正确", http.StatusBadRequest)
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	// 检查配置文件中是否已存在
	if _, exists := s.config.Sites[name]; exists {
		s.respondError(w, "网站已存在", http.StatusConflict)
		return
	}

	sitePath := filepath.Join(s.config.WebRoot, name)
	if _, err := os.Stat(sitePath); err == nil {
		s.respondError(w, "网站目录已存在", http.StatusConflict)
		return
	}

	// 创建网站目录
	if err := os.MkdirAll(sitePath, 0755); err != nil {
		s.respondError(w, fmt.Sprintf("创建目录失败: %v", err), http.StatusInternalServerError)
		return
	}

	// 初始化git仓库
	if s.config.EnableVersioning {
		if err := s.initGitRepo(sitePath); err != nil {
			s.respondError(w, fmt.Sprintf("初始化git仓库失败: %v", err), http.StatusInternalServerError)
			return
		}
	}

	// 在配置中添加网站
	s.config.Sites[name] = Site{
		Name:   name,
		Desc:   req.Desc,
		Owner:  user.Name,
		Users:  []string{},
	}

	// 保存配置
	if err := s.saveConfig(); err != nil {
		s.respondError(w, fmt.Sprintf("保存配置失败: %v", err), http.StatusInternalServerError)
		return
	}

	// 重新加载网站到内存
	s.reloadSitesFromConfig()

	var domain string
	var url string
	if s.config.Mode == "subdomain" {
		domain = fmt.Sprintf("%s.%s", name, s.config.BaseDomain)
		url = fmt.Sprintf("http://%s", domain)
	} else {
		domain = fmt.Sprintf("%s/%s", s.config.SingleDomain, name)
		url = fmt.Sprintf("http://%s/%s", s.config.BaseDomain, name)
	}

	s.respondJSON(w, map[string]interface{}{
		"id":         name,
		"name":       name,
		"domain":     domain,
		"path":       sitePath,
		"desc":       req.Desc,
		"url":        url,
		"created_at": time.Now().Format("2006-01-02 15:04:05"),
		"updated_at": time.Now().Format("2006-01-02 15:04:05"),
	})
}

// handleDeleteSite 删除网站
func (s *DeployServer) handleDeleteSite(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		s.respondError(w, "方法不允许", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Name string `json:"name"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.respondError(w, "无效的请求", http.StatusBadRequest)
		return
	}

	if req.Name == "" {
		s.respondError(w, "网站名称不能为空", http.StatusBadRequest)
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	// 检查配置中是否存在
	siteConfig, exists := s.config.Sites[req.Name]
	if !exists {
		s.respondError(w, "网站不存在", http.StatusNotFound)
		return
	}

	// 检查权限（只有所有者或管理员可以删除）
	user := userFromContext(r.Context())
	if user == nil || (!user.IsAdmin && siteConfig.Owner != user.Name) {
		s.respondError(w, "没有权限删除此网站", http.StatusForbidden)
		return
	}

	sitePath := filepath.Join(s.config.WebRoot, req.Name)
	if _, err := os.Stat(sitePath); os.IsNotExist(err) {
		// 即使目录不存在，也从配置中删除
		delete(s.config.Sites, req.Name)
		s.saveConfig()
		s.reloadSitesFromConfig()
		s.respondError(w, "网站目录不存在，但已从配置中删除", http.StatusNotFound)
		return
	}

	// 删除目录
	if err := os.RemoveAll(sitePath); err != nil {
		s.respondError(w, fmt.Sprintf("删除失败: %v", err), http.StatusInternalServerError)
		return
	}

	// 从配置中删除
	delete(s.config.Sites, req.Name)

	// 保存配置
	if err := s.saveConfig(); err != nil {
		s.respondError(w, fmt.Sprintf("保存配置失败: %v", err), http.StatusInternalServerError)
		return
	}

	// 重新加载网站到内存
	s.reloadSitesFromConfig()

	s.respondJSON(w, map[string]interface{}{
		"message": "删除成功",
	})
}

// handleUpdateSite 更新网站信息
func (s *DeployServer) handleUpdateSite(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		s.respondError(w, "方法不允许", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Name string `json:"name"`
		Desc string `json:"desc"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.respondError(w, "无效的请求", http.StatusBadRequest)
		return
	}

	if req.Name == "" {
		s.respondError(w, "网站名称不能为空", http.StatusBadRequest)
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	// 检查配置中是否存在
	siteConfig, exists := s.config.Sites[req.Name]
	if !exists {
		s.respondError(w, "网站不存在", http.StatusNotFound)
		return
	}

	// 检查权限（只有所有者或管理员可以更新）
	user := userFromContext(r.Context())
	if user == nil || (!user.IsAdmin && siteConfig.Owner != user.Name) {
		s.respondError(w, "没有权限更新此网站", http.StatusForbidden)
		return
	}

	// 更新描述
	siteConfig.Desc = req.Desc
	s.config.Sites[req.Name] = siteConfig

	// 保存配置
	if err := s.saveConfig(); err != nil {
		s.respondError(w, fmt.Sprintf("保存配置失败: %v", err), http.StatusInternalServerError)
		return
	}

	// 重新加载网站到内存
	s.reloadSitesFromConfig()

	s.respondJSON(w, map[string]interface{}{
		"message": "更新成功",
	})
}

// handleDeploy 部署网站
func (s *DeployServer) handleDeploy(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		s.respondError(w, "方法不允许", http.StatusMethodNotAllowed)
		return
	}

	// 解析multipart表单
	if err := r.ParseMultipartForm(32 << 20); err != nil {
		s.respondError(w, "解析表单失败", http.StatusBadRequest)
		return
	}

	name := r.FormValue("name")
	if name == "" {
		s.respondError(w, "网站名称不能为空", http.StatusBadRequest)
		return
	}

	message := r.FormValue("message")
	if message == "" {
		message = "更新部署"
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		s.respondError(w, "获取文件失败", http.StatusBadRequest)
		return
	}
	defer file.Close()

	sitePath := filepath.Join(s.config.WebRoot, name)
	if _, err := os.Stat(sitePath); os.IsNotExist(err) {
		s.respondError(w, "网站不存在", http.StatusNotFound)
		return
	}

	// 清空现有文件（保留.git目录）
	entries, _ := os.ReadDir(sitePath)
	for _, entry := range entries {
		if entry.Name() != ".git" {
			os.RemoveAll(filepath.Join(sitePath, entry.Name()))
		}
	}

	// 保存上传的文件
	filename := header.Filename
	if filename == "" {
		filename = "index.html"
	}

	destPath := filepath.Join(sitePath, filename)
	destFile, err := os.Create(destPath)
	if err != nil {
		s.respondError(w, fmt.Sprintf("创建文件失败: %v", err), http.StatusInternalServerError)
		return
	}
	defer destFile.Close()

	if _, err := io.Copy(destFile, file); err != nil {
		s.respondError(w, "保存文件失败", http.StatusInternalServerError)
		return
	}

	// 如果是HTML文件，解压相关资源
	if strings.HasSuffix(strings.ToLower(filename), ".html") {
		s.extractHTMLResources(destPath, sitePath)
	}

	// Git提交
	if s.config.EnableVersioning {
		if err := s.commitChanges(sitePath, message); err != nil {
			// 提交失败不影响部署
			fmt.Printf("Git提交失败: %v\n", err)
		}
	}

	s.mu.Lock()
	if site, exists := s.sites[name]; exists {
		site.UpdatedAt = time.Now()
	}
	s.mu.Unlock()

	s.respondJSON(w, map[string]interface{}{
		"message": "部署成功",
		"path":    destPath,
	})
}

// extractHTMLResources 提取HTML中的资源（base64图片等）
func (s *DeployServer) extractHTMLResources(htmlPath, sitePath string) {
	// 这里可以添加逻辑来处理HTML中嵌入的资源
	// 例如将base64图片提取为独立文件
}

// handleVersions 查看版本
func (s *DeployServer) handleVersions(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		s.respondError(w, "方法不允许", http.StatusMethodNotAllowed)
		return
	}

	name := r.URL.Query().Get("name")
	if name == "" {
		s.respondError(w, "网站名称不能为空", http.StatusBadRequest)
		return
	}

	sitePath := filepath.Join(s.config.WebRoot, name)
	if _, err := os.Stat(sitePath); os.IsNotExist(err) {
		s.respondError(w, "网站不存在", http.StatusNotFound)
		return
	}

	if !s.config.EnableVersioning {
		s.respondError(w, "版本控制未启用", http.StatusBadRequest)
		return
	}

	versions, err := s.getGitVersions(sitePath)
	if err != nil {
		s.respondError(w, fmt.Sprintf("获取版本失败: %v", err), http.StatusInternalServerError)
		return
	}

	s.respondJSON(w, versions)
}

// handleRollback 恢复版本
func (s *DeployServer) handleRollback(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		s.respondError(w, "方法不允许", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Name    string `json:"name"`
		Hash    string `json:"hash"`
		Message string `json:"message"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.respondError(w, "无效的请求", http.StatusBadRequest)
		return
	}

	if req.Name == "" || req.Hash == "" {
		s.respondError(w, "网站名称和版本哈希不能为空", http.StatusBadRequest)
		return
	}

	if req.Message == "" {
		req.Message = "回滚版本"
	}

	sitePath := filepath.Join(s.config.WebRoot, req.Name)
	if _, err := os.Stat(sitePath); os.IsNotExist(err) {
		s.respondError(w, "网站不存在", http.StatusNotFound)
		return
	}

	if !s.config.EnableVersioning {
		s.respondError(w, "版本控制未启用", http.StatusBadRequest)
		return
	}

	if err := s.rollbackVersion(sitePath, req.Hash, req.Message); err != nil {
		s.respondError(w, fmt.Sprintf("回滚失败: %v", err), http.StatusInternalServerError)
		return
	}

	s.respondJSON(w, map[string]interface{}{
		"message": "回滚成功",
	})
}

// initGitRepo 初始化git仓库
func (s *DeployServer) initGitRepo(path string) error {
	cmd := exec.Command("git", "init")
	cmd.Dir = path
	if err := cmd.Run(); err != nil {
		return err
	}

	// 配置用户（如果需要）
	cmd = exec.Command("git", "config", "user.email", "deployer@local")
	cmd.Dir = path
	_ = cmd.Run()

	cmd = exec.Command("git", "config", "user.name", "Deployer")
	cmd.Dir = path
	_ = cmd.Run()

	return nil
}

// commitChanges 提交更改
func (s *DeployServer) commitChanges(path, message string) error {
	// 添加所有文件
	cmd := exec.Command("git", "add", ".")
	cmd.Dir = path
	if err := cmd.Run(); err != nil {
		return err
	}

	// 提交
	cmd = exec.Command("git", "commit", "-m", message)
	cmd.Dir = path
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%v: %s", err, string(output))
	}

	return nil
}

// getGitVersions 获取git版本列表
func (s *DeployServer) getGitVersions(path string) ([]Version, error) {
	cmd := exec.Command("git", "log", "--pretty=format:%H|%s|%an|%ai", "-20")
	cmd.Dir = path
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	lines := strings.Split(string(output), "\n")
	versions := make([]Version, 0, len(lines))

	for _, line := range lines {
		if line == "" {
			continue
		}
		parts := strings.SplitN(line, "|", 4)
		if len(parts) != 4 {
			continue
		}

		date, err := time.Parse("2006-01-02 15:04:05 -0700", parts[3])
		if err != nil {
			continue
		}

		versions = append(versions, Version{
			Hash:    parts[0],
			Message: parts[1],
			Author:  parts[2],
			Date:    date,
		})
	}

	return versions, nil
}

// rollbackVersion 回滚到指定版本
func (s *DeployServer) rollbackVersion(path, hash, message string) error {
	// 先检查是否有未提交的更改
	cmd := exec.Command("git", "status", "--porcelain")
	cmd.Dir = path
	output, _ := cmd.Output()

	// 如果有未提交的更改，先暂存
	if len(strings.TrimSpace(string(output))) > 0 {
		// 创建临时提交保存当前状态
		cmd = exec.Command("git", "add", ".")
		cmd.Dir = path
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("暂存当前状态失败: %v", err)
		}

		cmd = exec.Command("git", "commit", "-m", "临时提交：回滚前保存")
		cmd.Dir = path
		_ = cmd.Run() // 忽略错误，可能没有内容需要提交
	}

	// 使用 git checkout 恢复到指定版本
	cmd = exec.Command("git", "checkout", hash, "--", ".")
	cmd.Dir = path
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("回滚失败: %v: %s", err, string(output))
	}

	// 添加更改的文件
	cmd = exec.Command("git", "add", "-A")
	cmd.Dir = path
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("添加文件失败: %v", err)
	}

	// 创建回滚提交
	cmd = exec.Command("git", "commit", "-m", message)
	cmd.Dir = path
	output, err = cmd.CombinedOutput()
	if err != nil {
		// 如果没有变更，说明已经在该版本，不算错误
		if strings.Contains(string(output), "nothing to commit") {
			return nil
		}
		return fmt.Errorf("创建回滚提交失败: %v: %s", err, string(output))
	}

	return nil
}

// respondJSON 返回JSON响应
func (s *DeployServer) respondJSON(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}

// respondError 返回错误响应
func (s *DeployServer) respondError(w http.ResponseWriter, message string, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"error": message,
	})
}

// handleDeployFull 全量部署
func (s *DeployServer) handleDeployFull(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		s.respondError(w, "方法不允许", http.StatusMethodNotAllowed)
		return
	}

	name := r.FormValue("name")
	if name == "" {
		s.respondError(w, "网站名称不能为空", http.StatusBadRequest)
		return
	}

	message := r.FormValue("message")
	if message == "" {
		message = "全量部署"
	}

	// 获取上传的文件
	file, _, err := r.FormFile("package")
	if err != nil {
		s.respondError(w, "获取部署包失败", http.StatusBadRequest)
		return
	}
	defer file.Close()

	sitePath := filepath.Join(s.config.WebRoot, name)
	if _, err := os.Stat(sitePath); os.IsNotExist(err) {
		s.respondError(w, "网站不存在", http.StatusNotFound)
		return
	}

	// 解压部署包
	if err := s.extractPackage(file, sitePath); err != nil {
		s.respondError(w, fmt.Sprintf("解压失败: %v", err), http.StatusInternalServerError)
		return
	}

	// Git提交
	if s.config.EnableVersioning {
		if err := s.commitChanges(sitePath, message); err != nil {
			fmt.Printf("Git提交失败: %v\n", err)
		}
	}

	s.mu.Lock()
	if site, exists := s.sites[name]; exists {
		site.UpdatedAt = time.Now()
	}
	s.mu.Unlock()

	s.respondJSON(w, map[string]interface{}{
		"message": "全量部署成功",
		"mode":    "full",
	})
}

// handleDeployIncremental 增量部署
func (s *DeployServer) handleDeployIncremental(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		s.respondError(w, "方法不允许", http.StatusMethodNotAllowed)
		return
	}

	name := r.FormValue("name")
	if name == "" {
		s.respondError(w, "网站名称不能为空", http.StatusBadRequest)
		return
	}

	message := r.FormValue("message")
	if message == "" {
		message = "增量部署"
	}

	// 获取上传的文件
	file, _, err := r.FormFile("package")
	if err != nil {
		s.respondError(w, "获取部署包失败", http.StatusBadRequest)
		return
	}
	defer file.Close()

	sitePath := filepath.Join(s.config.WebRoot, name)
	if _, err := os.Stat(sitePath); os.IsNotExist(err) {
		s.respondError(w, "网站不存在", http.StatusNotFound)
		return
	}

	// 解压增量包
	if err := s.extractPackage(file, sitePath); err != nil {
		s.respondError(w, fmt.Sprintf("解压失败: %v", err), http.StatusInternalServerError)
		return
	}

	// Git提交
	if s.config.EnableVersioning {
		if err := s.commitChanges(sitePath, message); err != nil {
			fmt.Printf("Git提交失败: %v\n", err)
		}
	}

	s.mu.Lock()
	if site, exists := s.sites[name]; exists {
		site.UpdatedAt = time.Now()
	}
	s.mu.Unlock()

	s.respondJSON(w, map[string]interface{}{
		"message": "增量部署成功",
		"mode":    "incremental",
	})
}

// extractPackage 解压部署包
func (s *DeployServer) extractPackage(packageFile io.Reader, destPath string) error {
	// 创建gzip reader
	gzReader, err := gzip.NewReader(packageFile)
	if err != nil {
		return fmt.Errorf("创建gzip reader失败: %v", err)
	}
	defer gzReader.Close()

	// 创建tar reader
	tarReader := tar.NewReader(gzReader)

	// 遍历tar文件
	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("读取tar条目失败: %v", err)
		}

		// 构建目标路径
		targetPath := filepath.Join(destPath, header.Name)

		// 检查路径安全
		if !strings.HasPrefix(targetPath, destPath) {
			return fmt.Errorf("非法路径: %s", header.Name)
		}

		// 根据文件类型处理
		switch header.Typeflag {
		case tar.TypeDir:
			// 创建目录
			if err := os.MkdirAll(targetPath, 0755); err != nil {
				return fmt.Errorf("创建目录失败: %v", err)
			}

		case tar.TypeReg:
			// 创建文件
			if err := os.MkdirAll(filepath.Dir(targetPath), 0755); err != nil {
				return fmt.Errorf("创建父目录失败: %v", err)
			}

			outFile, err := os.Create(targetPath)
			if err != nil {
				return fmt.Errorf("创建文件失败: %v", err)
			}

			if _, err := io.Copy(outFile, tarReader); err != nil {
				outFile.Close()
				return fmt.Errorf("写入文件失败: %v", err)
			}
			outFile.Close()

		case tar.TypeSymlink:
			// 忽略符号链接
			continue
		}
	}

	return nil
}

// FileHash 文件哈希信息
type FileHash struct {
	Path string `json:"path"`
	Hash string `json:"hash"`
}

// FileHashList 文件哈希列表
type FileHashList struct {
	Files []FileHash `json:"files"`
}

// handleExport 导出网站文件（打包下载）
func (s *DeployServer) handleExport(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		s.respondError(w, "方法不允许", http.StatusMethodNotAllowed)
		return
	}

	name := r.URL.Query().Get("name")
	if name == "" {
		s.respondError(w, "网站名称不能为空", http.StatusBadRequest)
		return
	}

	sitePath := filepath.Join(s.config.WebRoot, name)
	if _, err := os.Stat(sitePath); os.IsNotExist(err) {
		s.respondError(w, "网站不存在", http.StatusNotFound)
		return
	}

	// 设置响应头为tar.gz文件
	w.Header().Set("Content-Type", "application/x-gzip")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s.tar.gz", name))

	// 创建gzip writer
	gzWriter := gzip.NewWriter(w)
	defer gzWriter.Close()

	// 创建tar writer
	tarWriter := tar.NewWriter(gzWriter)
	defer tarWriter.Close()

	// 遍历网站目录并打包
	err := filepath.Walk(sitePath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// 跳过根目录本身
		if path == sitePath {
			return nil
		}

		// 跳过.git目录
		if strings.Contains(path, ".git") {
			if info.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		// 获取相对路径
		relPath, err := filepath.Rel(sitePath, path)
		if err != nil {
			return err
		}

		// 创建tar头
		header, err := tar.FileInfoHeader(info, "")
		if err != nil {
			return err
		}
		header.Name = relPath

		// 写入头
		if err := tarWriter.WriteHeader(header); err != nil {
			return err
		}

		// 如果是文件，写入内容
		if !info.IsDir() {
			file, err := os.Open(path)
			if err != nil {
				return err
			}
			defer file.Close()

			_, err = io.Copy(tarWriter, file)
			return err
		}

		return nil
	})

	if err != nil {
		// 如果出错，尝试写入错误信息（可能已经写入了一些数据）
		fmt.Printf("导出文件失败: %v\n", err)
	}
}

// handleListUsers 列出所有用户（管理员）
func (s *DeployServer) handleListUsers(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		s.respondError(w, "方法不允许", http.StatusMethodNotAllowed)
		return
	}

	users := make([]User, 0, len(s.config.Users))
	for _, user := range s.config.Users {
		users = append(users, User{
			Name:    user.Name,
			IsAdmin: user.IsAdmin,
			// 不返回密码
		})
	}

	s.respondJSON(w, map[string]interface{}{
		"users": users,
	})
}

// handleCreateUser 创建用户（管理员）
func (s *DeployServer) handleCreateUser(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		s.respondError(w, "方法不允许", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Name     string `json:"name"`
		Password string `json:"password"`
		IsAdmin  bool   `json:"isAdmin"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.respondError(w, "无效的请求", http.StatusBadRequest)
		return
	}

	if req.Name == "" || req.Password == "" {
		s.respondError(w, "用户名和密码不能为空", http.StatusBadRequest)
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.config.Users[req.Name]; exists {
		s.respondError(w, "用户已存在", http.StatusConflict)
		return
	}

	// 创建用户
	s.config.Users[req.Name] = User{
		Name:     req.Name,
		Password: req.Password,
		IsAdmin:  req.IsAdmin,
	}

	// 保存配置
	if err := s.saveConfig(); err != nil {
		s.respondError(w, "保存配置失败: "+err.Error(), http.StatusInternalServerError)
		return
	}

	s.respondJSON(w, map[string]string{
		"message": "用户创建成功",
		"name":    req.Name,
	})
}

// handleUpdateUser 更新用户（管理员）
func (s *DeployServer) handleUpdateUser(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		s.respondError(w, "方法不允许", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Name     string `json:"name"`
		Password string `json:"password,omitempty"`
		IsAdmin  *bool  `json:"isAdmin,omitempty"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.respondError(w, "无效的请求", http.StatusBadRequest)
		return
	}

	if req.Name == "" {
		s.respondError(w, "用户名不能为空", http.StatusBadRequest)
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	user, exists := s.config.Users[req.Name]
	if !exists {
		s.respondError(w, "用户不存在", http.StatusNotFound)
		return
	}

	// 更新密码
	if req.Password != "" {
		user.Password = req.Password
	}

	// 更新管理员权限
	if req.IsAdmin != nil {
		user.IsAdmin = *req.IsAdmin
	}

	s.config.Users[req.Name] = user

	// 保存配置
	if err := s.saveConfig(); err != nil {
		s.respondError(w, "保存配置失败: "+err.Error(), http.StatusInternalServerError)
		return
	}

	s.respondJSON(w, map[string]string{
		"message": "用户更新成功",
		"name":    req.Name,
	})
}

// handleDeleteUser 删除用户（管理员）
func (s *DeployServer) handleDeleteUser(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		s.respondError(w, "方法不允许", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Name string `json:"name"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.respondError(w, "无效的请求", http.StatusBadRequest)
		return
	}

	if req.Name == "" {
		s.respondError(w, "用户名不能为空", http.StatusBadRequest)
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.config.Users[req.Name]; !exists {
		s.respondError(w, "用户不存在", http.StatusNotFound)
		return
	}

	delete(s.config.Users, req.Name)

	// 保存配置
	if err := s.saveConfig(); err != nil {
		s.respondError(w, "保存配置失败: "+err.Error(), http.StatusInternalServerError)
		return
	}

	s.respondJSON(w, map[string]string{
		"message": "用户删除成功",
		"name":    req.Name,
	})
}

// handleAuthorizeSite 授权网站给用户（owner）
func (s *DeployServer) handleAuthorizeSite(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		s.respondError(w, "方法不允许", http.StatusMethodNotAllowed)
		return
	}

	user := userFromContext(r.Context())
	if user == nil {
		s.respondError(w, "未授权", http.StatusUnauthorized)
		return
	}

	var req struct {
		SiteName  string   `json:"siteName"`
		Username  string   `json:"username"`
		Usernames []string `json:"usernames"` // 批量授权
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.respondError(w, "无效的请求", http.StatusBadRequest)
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	site, exists := s.config.Sites[req.SiteName]
	if !exists {
		s.respondError(w, "网站不存在", http.StatusNotFound)
		return
	}

	// 检查是否是 owner 或管理员
	if site.Owner != user.Name && !user.IsAdmin {
		s.respondError(w, "只有网站所有者或管理员可以授权", http.StatusForbidden)
		return
	}

	// 获取要授权的用户列表
	usersToAuthorize := req.Usernames
	if req.Username != "" {
		usersToAuthorize = append(usersToAuthorize, req.Username)
	}

	// 授权用户
	for _, username := range usersToAuthorize {
		// 检查用户是否存在
		if _, exists := s.config.Users[username]; !exists {
			continue
		}

		// 检查是否已授权
		alreadyAuthorized := false
		for _, u := range site.Users {
			if u == username {
				alreadyAuthorized = true
				break
			}
		}

		if !alreadyAuthorized {
			site.Users = append(site.Users, username)
		}
	}

	s.config.Sites[req.SiteName] = site

	// 保存配置
	if err := s.saveConfig(); err != nil {
		s.respondError(w, "保存配置失败: "+err.Error(), http.StatusInternalServerError)
		return
	}

	s.respondJSON(w, map[string]interface{}{
		"message": "授权成功",
		"site":    site,
	})
}

// handleUnauthorizeSite 取消网站授权（owner）
func (s *DeployServer) handleUnauthorizeSite(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		s.respondError(w, "方法不允许", http.StatusMethodNotAllowed)
		return
	}

	user := userFromContext(r.Context())
	if user == nil {
		s.respondError(w, "未授权", http.StatusUnauthorized)
		return
	}

	var req struct {
		SiteName string `json:"siteName"`
		Username string `json:"username"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.respondError(w, "无效的请求", http.StatusBadRequest)
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	site, exists := s.config.Sites[req.SiteName]
	if !exists {
		s.respondError(w, "网站不存在", http.StatusNotFound)
		return
	}

	// 检查是否是 owner 或管理员
	if site.Owner != user.Name && !user.IsAdmin {
		s.respondError(w, "只有网站所有者或管理员可以取消授权", http.StatusForbidden)
		return
	}

	// 移除用户授权
	newUsers := make([]string, 0)
	for _, u := range site.Users {
		if u != req.Username {
			newUsers = append(newUsers, u)
		}
	}
	site.Users = newUsers

	s.config.Sites[req.SiteName] = site

	// 保存配置
	if err := s.saveConfig(); err != nil {
		s.respondError(w, "保存配置失败: "+err.Error(), http.StatusInternalServerError)
		return
	}

	s.respondJSON(w, map[string]interface{}{
		"message": "取消授权成功",
		"site":    site,
	})
}
