package server

import (
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

// Config 服务器配置
type Config struct {
	BaseDomain      string `json:"base_domain"`      // 基础域名（子域名模式）
	WebRoot         string `json:"web_root"`         // 网站根目录
	Mode            string `json:"mode"`             // 部署模式：subdomain 或 path
	SingleDomain    string `json:"single_domain"`    // 单域名模式下的域名
	Port            int    `json:"port"`             // 服务器端口
	EnableVersioning bool   `json:"enable_versioning"` // 是否启用版本控制
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
	config Config
	sites  map[string]*Website
	mu     sync.RWMutex
}

// NewDeployServer 创建新的部署服务器
func NewDeployServer(config Config) *DeployServer {
	return &DeployServer{
		config: config,
		sites:  make(map[string]*Website),
	}
}

// Start 启动服务器
func (s *DeployServer) Start() error {
	// 确保web根目录存在
	if err := os.MkdirAll(s.config.WebRoot, 0755); err != nil {
		return fmt.Errorf("创建web根目录失败: %v", err)
	}

	// 注册路由
	http.HandleFunc("/api/sites", s.corsMiddleware(s.handleSites))
	http.HandleFunc("/api/sites/create", s.corsMiddleware(s.handleCreateSite))
	http.HandleFunc("/api/sites/delete", s.corsMiddleware(s.handleDeleteSite))
	http.HandleFunc("/api/sites/deploy", s.corsMiddleware(s.handleDeploy))
	http.HandleFunc("/api/sites/versions", s.corsMiddleware(s.handleVersions))
	http.HandleFunc("/api/sites/rollback", s.corsMiddleware(s.handleRollback))
	http.HandleFunc("/api/sites/list", s.corsMiddleware(s.handleListSites))

	addr := fmt.Sprintf(":%d", s.config.Port)
	fmt.Printf("服务器启动在 http://localhost%s\n", addr)
	fmt.Printf("部署模式: %s\n", s.config.Mode)
	if s.config.Mode == "subdomain" {
		fmt.Printf("基础域名: %s\n", s.config.BaseDomain)
	} else {
		fmt.Printf("域名: %s\n", s.config.SingleDomain)
	}

	return http.ListenAndServe(addr, nil)
}

// corsMiddleware CORS中间件
func (s *DeployServer) corsMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS, DELETE")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

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

	sites := []string{}
	for _, entry := range entries {
		if entry.IsDir() {
			sites = append(sites, entry.Name())
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
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.respondError(w, "无效的请求", http.StatusBadRequest)
		return
	}

	if req.Name == "" {
		s.respondError(w, "网站名称不能为空", http.StatusBadRequest)
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

	sitePath := filepath.Join(s.config.WebRoot, name)
	if _, err := os.Stat(sitePath); err == nil {
		s.respondError(w, "网站已存在", http.StatusConflict)
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

	var domain string
	if s.config.Mode == "subdomain" {
		domain = fmt.Sprintf("%s.%s", name, s.config.BaseDomain)
	} else {
		domain = fmt.Sprintf("%s/%s", s.config.SingleDomain, name)
	}

	site := &Website{
		ID:        name,
		Name:      name,
		Domain:    domain,
		Path:      sitePath,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	s.sites[name] = site

	s.respondJSON(w, site)
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

	sitePath := filepath.Join(s.config.WebRoot, req.Name)
	if _, err := os.Stat(sitePath); os.IsNotExist(err) {
		s.respondError(w, "网站不存在", http.StatusNotFound)
		return
	}

	// 删除目录
	if err := os.RemoveAll(sitePath); err != nil {
		s.respondError(w, fmt.Sprintf("删除失败: %v", err), http.StatusInternalServerError)
		return
	}

	delete(s.sites, req.Name)

	s.respondJSON(w, map[string]interface{}{
		"message": "删除成功",
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
	// 重置到指定版本
	cmd := exec.Command("git", "reset", "--hard", hash)
	cmd.Dir = path
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%v: %s", err, string(output))
	}

	// 创建回滚提交
	if err := s.commitChanges(path, message); err != nil {
		return err
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
