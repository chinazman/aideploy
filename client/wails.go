//go:build !cli
// +build !cli

package main

import (
	"context"
	"embed"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os/exec"
	"runtime"
	"strings"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	wailsRuntime "github.com/wailsapp/wails/v2/pkg/runtime"
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {
	// 创建应用实例
	app := NewApp()

	// 启动应用
	err := wails.Run(&options.App{
		Title:  "AI原型部署工具",
		Width:  1024,
		Height: 768,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		OnStartup:  app.startup,
		OnShutdown: app.shutdown,
		Bind: []interface{}{
			app,
		},
	})

	if err != nil {
		println("Error:", err.Error())
	}
}

// App 应用结构
type App struct {
	ctx        context.Context
	apiBaseURL string
	username   string
	password   string
	config     *ClientConfig
}

// NewApp 创建应用实例
func NewApp() *App {
	// 从配置文件加载服务端地址和用户名/密码
	config, err := LoadConfig()
	apiBaseURL := "http://localhost:8080/api"
	username := ""
	password := ""

	if err == nil {
		if config.ServerURL != "" {
			apiBaseURL = config.ServerURL
		}
		username = config.Username
		password = config.Password
	}

	return &App{
		apiBaseURL: apiBaseURL,
		username:   username,
		password:   password,
		config:     config,
	}
}

// startup 应用启动时调用
func (a *App) startup(ctx context.Context) {
	// 保存 context 供后续使用
	a.ctx = ctx
}

// shutdown 应用关闭时调用
func (a *App) shutdown(ctx context.Context) {
	// 清理资源
}

// SelectDirectory 选择目录
func (a *App) SelectDirectory() (string, error) {
	selection, err := wailsRuntime.OpenDirectoryDialog(a.ctx, wailsRuntime.OpenDialogOptions{
		Title: "选择网站目录",
	})

	if err != nil {
		return "", err
	}

	return selection, nil
}

// OpenDirectory 打开本地目录
func (a *App) OpenDirectory(path string) error {
	if path == "" {
		return fmt.Errorf("目录路径不能为空")
	}

	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("explorer", path)
	case "darwin":
		cmd = exec.Command("open", path)
	default: // linux and others
		cmd = exec.Command("xdg-open", path)
	}

	return cmd.Start()
}

// Website 网站信息
type Website struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Domain    string `json:"domain"`
	Path      string `json:"path"`
	Desc      string `json:"desc"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

// Version 版本信息
type Version struct {
	Hash    string `json:"hash"`
	Message string `json:"message"`
	Author  string `json:"author"`
	Date    string `json:"date"`
}

// addAuthToRequest 添加认证信息到请求头
func (a *App) addAuthToRequest(req *http.Request) {
	// 优先使用用户名/密码认证
	if a.username != "" && a.password != "" {
		req.Header.Set("X-Username", a.username)
		req.Header.Set("X-Password", a.password)
	}
}

// CreateSite 创建网站
func (a *App) CreateSite(name, desc string) (*Website, error) {
	payload := map[string]string{
		"name": name,
		"desc": desc,
	}

	data, _ := json.Marshal(payload)
	req, err := http.NewRequest("POST", a.apiBaseURL+"/sites/create", strings.NewReader(string(data)))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	a.addAuthToRequest(req)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf(string(body))
	}

	var site Website
	if err := json.Unmarshal(body, &site); err != nil {
		return nil, err
	}

	return &site, nil
}

// DeleteSite 删除网站
func (a *App) DeleteSite(name string) error {
	payload := map[string]string{
		"name": name,
	}

	data, _ := json.Marshal(payload)
	req, err := http.NewRequest("POST", a.apiBaseURL+"/sites/delete", strings.NewReader(string(data)))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	a.addAuthToRequest(req)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf(string(body))
	}

	return nil
}

// UpdateSiteDesc 更新网站描述
func (a *App) UpdateSiteDesc(name, desc string) error {
	payload := map[string]string{
		"name": name,
		"desc": desc,
	}

	data, _ := json.Marshal(payload)
	req, err := http.NewRequest("POST", a.apiBaseURL+"/sites/update", strings.NewReader(string(data)))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	a.addAuthToRequest(req)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf(string(body))
	}

	return nil
}

// DeploySite 部署网站 (使用绑定的目录)
func (a *App) DeploySite(name, message string) error {
	// 从配置中获取绑定的目录
	dirPath, ok := a.config.SitePaths[name]
	if !ok || dirPath == "" {
		return fmt.Errorf("网站 '%s' 未绑定发布目录", name)
	}

	// 创建部署器
	deployer := NewDeployer(a.apiBaseURL, name)

	// 执行智能部署
	if err := deployer.Deploy(dirPath, message); err != nil {
		return err
	}

	return nil
}

// SiteInfo 网站信息
type SiteInfo struct {
	Name   string `json:"name"`
	Domain string `json:"domain"`
	Desc   string `json:"desc"`
	URL    string `json:"url"`
}

// ListSites 列出所有网站
func (a *App) ListSites() ([]SiteInfo, error) {
	req, err := http.NewRequest("GET", a.apiBaseURL+"/sites/list", nil)
	if err != nil {
		return nil, err
	}

	a.addAuthToRequest(req)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf(string(body))
	}

	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}

	sites, ok := result["sites"].([]interface{})
	if !ok {
		return []SiteInfo{}, nil
	}

	siteList := make([]SiteInfo, 0, len(sites))
	for _, site := range sites {
		if siteMap, ok := site.(map[string]interface{}); ok {
			name, _ := siteMap["name"].(string)
			domain, _ := siteMap["domain"].(string)
			desc, _ := siteMap["desc"].(string)
			url, _ := siteMap["url"].(string)
			siteList = append(siteList, SiteInfo{
				Name:   name,
				Domain: domain,
				Desc:   desc,
				URL:    url,
			})
		}
	}

	return siteList, nil
}

// GetVersions 获取版本列表
func (a *App) GetVersions(name string) ([]Version, error) {
	url := fmt.Sprintf("%s/sites/versions?name=%s", a.apiBaseURL, name)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	a.addAuthToRequest(req)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf(string(body))
	}

	var versions []Version
	if err := json.Unmarshal(body, &versions); err != nil {
		return nil, err
	}

	return versions, nil
}

// Rollback 回滚版本
func (a *App) Rollback(name, hash, message string) error {
	if message == "" {
		message = "回滚版本"
	}

	payload := map[string]string{
		"name":    name,
		"hash":    hash,
		"message": message,
	}

	data, _ := json.Marshal(payload)
	req, err := http.NewRequest("POST", a.apiBaseURL+"/sites/rollback", strings.NewReader(string(data)))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	a.addAuthToRequest(req)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf(string(body))
	}

	return nil
}

// GetConfig 获取当前配置
func (a *App) GetConfig() (*ClientConfig, error) {
	config, err := LoadConfig()
	if err != nil {
		return nil, err
	}
	return config, nil
}

// SaveConfig 保存配置
func (a *App) SaveConfig(config *ClientConfig) error {
	// 更新内存中的配置
	a.config = config
	a.apiBaseURL = config.ServerURL
	a.username = config.Username
	a.password = config.Password

	// 保存到文件
	if err := SaveConfig(config); err != nil {
		return err
	}

	return nil
}

// BindSiteDirectory 绑定网站目录
func (a *App) BindSiteDirectory(siteName, dirPath string) error {
	// 加载当前配置
	config, err := LoadConfig()
	if err != nil {
		return err
	}

	// 确保SitePaths初始化
	if config.SitePaths == nil {
		config.SitePaths = make(map[string]string)
	}

	// 绑定目录
	config.SitePaths[siteName] = dirPath

	// 保存配置
	if err := SaveConfig(config); err != nil {
		return err
	}

	// 更新内存中的配置
	a.config = config

	return nil
}

// FileChange 文件变更信息
type FileChange struct {
	Path     string `json:"path"`
	Type     string `json:"type"` // "modified", "added", "deleted"
	Size     int64  `json:"size"`
}

// ChangesResult 变更检查结果
type ChangesResult struct {
	HasChanges bool         `json:"has_changes"`
	Changes    []FileChange `json:"changes"`
	Summary    string       `json:"summary"`
}

// CheckChanges 检查文件变更
func (a *App) CheckChanges(siteName string) (*ChangesResult, error) {
	// 获取绑定的目录
	config, err := LoadConfig()
	if err != nil {
		return nil, err
	}

	if config.SitePaths == nil {
		return nil, fmt.Errorf("没有绑定任何目录")
	}

	sitePath, exists := config.SitePaths[siteName]
	if !exists {
		return nil, fmt.Errorf("网站 %s 未绑定目录", siteName)
	}

	// 创建部署器
	deployer := NewDeployer(a.apiBaseURL, siteName)

	// 扫描当前文件
	currentFiles, err := deployer.ScanFiles(sitePath)
	if err != nil {
		return nil, fmt.Errorf("扫描文件失败: %v", err)
	}

	// 加载之前的跟踪信息
	trackingData, err := deployer.LoadTracking()
	if err != nil {
		// 第一次部署，所有文件都是新增的
		changes := make([]FileChange, 0, len(currentFiles))
		for _, f := range currentFiles {
			changes = append(changes, FileChange{
				Path: f.Path,
				Type: "added",
				Size: f.Size,
			})
		}

		return &ChangesResult{
			HasChanges: len(changes) > 0,
			Changes:    changes,
			Summary:    fmt.Sprintf("首次部署，共 %d 个文件", len(changes)),
		}, nil
	}

	// 找出变更和删除的文件
	changedFiles := deployer.FindChangedFiles(currentFiles, trackingData.Files)
	deletedFiles := deployer.FindDeletedFiles(currentFiles, trackingData.Files)

	// 构建变更列表
	changes := make([]FileChange, 0, len(changedFiles)+len(deletedFiles))

	// 添加修改和新增的文件
	prevMap := make(map[string]FileStatus)
	for _, f := range trackingData.Files {
		prevMap[f.Path] = f
	}

	for _, f := range changedFiles {
		changeType := "modified"
		if _, exists := prevMap[f.Path]; !exists {
			changeType = "added"
		}

		changes = append(changes, FileChange{
			Path: f.Path,
			Type: changeType,
			Size: f.Size,
		})
	}

	// 添加删除的文件
	for _, path := range deletedFiles {
		changes = append(changes, FileChange{
			Path: path,
			Type: "deleted",
			Size: 0,
		})
	}

	// 生成摘要
	addedCount := 0
	modifiedCount := 0
	deletedCount := 0
	for _, c := range changes {
		switch c.Type {
		case "added":
			addedCount++
		case "modified":
			modifiedCount++
		case "deleted":
			deletedCount++
		}
	}

	summary := ""
	if addedCount > 0 && modifiedCount > 0 && deletedCount > 0 {
		summary = fmt.Sprintf("新增 %d，修改 %d，删除 %d", addedCount, modifiedCount, deletedCount)
	} else if addedCount > 0 && modifiedCount > 0 {
		summary = fmt.Sprintf("新增 %d，修改 %d", addedCount, modifiedCount)
	} else if addedCount > 0 && deletedCount > 0 {
		summary = fmt.Sprintf("新增 %d，删除 %d", addedCount, deletedCount)
	} else if modifiedCount > 0 && deletedCount > 0 {
		summary = fmt.Sprintf("修改 %d，删除 %d", modifiedCount, deletedCount)
	} else if addedCount > 0 {
		summary = fmt.Sprintf("新增 %d 个文件", addedCount)
	} else if modifiedCount > 0 {
		summary = fmt.Sprintf("修改 %d 个文件", modifiedCount)
	} else if deletedCount > 0 {
		summary = fmt.Sprintf("删除 %d 个文件", deletedCount)
	} else {
		summary = "没有变更"
	}

	return &ChangesResult{
		HasChanges: len(changes) > 0,
		Changes:    changes,
		Summary:    summary,
	}, nil
}

// PullSite 从服务器覆盖本地
func (a *App) PullSite(name string) error {
	// 从配置中获取绑定的目录
	dirPath, ok := a.config.SitePaths[name]
	if !ok || dirPath == "" {
		return fmt.Errorf("网站 '%s' 未绑定发布目录", name)
	}

	// 创建部署器
	deployer := NewDeployer(a.apiBaseURL, name)

	// 执行下载并覆盖
	if err := deployer.PullFromServer(dirPath); err != nil {
		return err
	}

	return nil
}
