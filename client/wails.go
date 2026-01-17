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
	"strings"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
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
	apiBaseURL string
	apiKey     string
	config     *ClientConfig
}

// NewApp 创建应用实例
func NewApp() *App {
	// 从配置文件加载服务端地址和API密钥
	config, err := LoadConfig()
	apiBaseURL := "http://localhost:8080/api"
	apiKey := ""

	if err == nil {
		if config.ServerURL != "" {
			apiBaseURL = config.ServerURL
		}
		apiKey = config.APIKey
	}

	return &App{
		apiBaseURL: apiBaseURL,
		apiKey:     apiKey,
		config:     config,
	}
}

// startup 应用启动时调用
func (a *App) startup(ctx context.Context) {
	// 可以在这里初始化一些资源
}

// shutdown 应用关闭时调用
func (a *App) shutdown(ctx context.Context) {
	// 清理资源
}

// Website 网站信息
type Website struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Domain    string `json:"domain"`
	Path      string `json:"path"`
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

// addAPIKeyToRequest 如果配置了API密钥，则添加到请求头
func (a *App) addAPIKeyToRequest(req *http.Request) {
	if a.apiKey != "" {
		req.Header.Set("X-API-Key", a.apiKey)
	}
}

// CreateSite 创建网站
func (a *App) CreateSite(name string) (*Website, error) {
	payload := map[string]string{
		"name": name,
	}

	data, _ := json.Marshal(payload)
	req, err := http.NewRequest("POST", a.apiBaseURL+"/sites/create", strings.NewReader(string(data)))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	a.addAPIKeyToRequest(req)

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
	a.addAPIKeyToRequest(req)

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

// ListSites 列出所有网站
func (a *App) ListSites() ([]string, error) {
	req, err := http.NewRequest("GET", a.apiBaseURL+"/sites/list", nil)
	if err != nil {
		return nil, err
	}

	a.addAPIKeyToRequest(req)

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
		return []string{}, nil
	}

	siteList := make([]string, len(sites))
	for i, site := range sites {
		if siteName, ok := site.(string); ok {
			siteList[i] = siteName
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

	a.addAPIKeyToRequest(req)

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
	a.addAPIKeyToRequest(req)

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
	a.apiKey = config.APIKey

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
