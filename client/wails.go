package main

import (
	"embed"
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
		OnStartup: app.startup,
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
}

// NewApp 创建应用实例
func NewApp() *App {
	return &App{
		apiBaseURL: "http://localhost:8080/api",
	}
}

// startup 应用启动时调用
func (a *App) startup() {
	// 可以在这里初始化一些资源
}

// shutdown 应用关闭时调用
func (a *App) shutdown() {
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

// CreateSite 创建网站
func (a *App) CreateSite(name string) (*Website, error) {
	payload := map[string]string{
		"name": name,
	}

	resp, err := http.Post(a.apiBaseURL+"/sites/create", "application/json", strings.NewReader(fmt.Sprintf(`{"name":"%s"}`, name)))
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

	resp, err := http.Post(a.apiBaseURL+"/sites/delete", "application/json", strings.NewReader(fmt.Sprintf(`{"name":"%s"}`, name)))
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

// DeploySite 部署网站
func (a *App) DeploySite(name, filePath, message string) error {
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	// 创建multipart请求
	var requestBody strings.Builder
	boundary := "----Boundary" + time.Now().Format("20060102150405")

	tempFile, err := os.CreateTemp("", "deploy-*.tmp")
	if err != nil {
		return err
	}
	defer os.Remove(tempFile.Name())
	defer tempFile.Close()

	fmt.Fprintf(tempFile, "--%s\r\n", boundary)
	fmt.Fprintf(tempFile, `Content-Disposition: form-data; name="name"%s\r\n\r\n`, "\r\n")
	fmt.Fprintf(tempFile, "%s\r\n", name)

	fmt.Fprintf(tempFile, "--%s\r\n", boundary)
	fmt.Fprintf(tempFile, `Content-Disposition: form-data; name="message"%s\r\n\r\n`, "\r\n")
	fmt.Fprintf(tempFile, "%s\r\n", message)

	fmt.Fprintf(tempFile, "--%s\r\n", boundary)
	fmt.Fprintf(tempFile, `Content-Disposition: form-data; name="file"; filename="%s"%s`, filepath.Base(filePath), "\r\n")
	fmt.Fprintf(tempFile, "Content-Type: application/octet-stream\r\n\r\n")

	file.Seek(0, 0)
	io.Copy(tempFile, file)

	fmt.Fprintf(tempFile, "\r\n--%s--\r\n", boundary)
	tempFile.Close()

	tempFile2, _ := os.Open(tempFile.Name())
	defer tempFile2.Close()

	req, _ := http.NewRequest("POST", a.apiBaseURL+"/sites/deploy", tempFile2)
	req.Header.Set("Content-Type", "multipart/form-data; boundary="+boundary)

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

// ListSites 列出所有网站
func (a *App) ListSites() ([]string, error) {
	resp, err := http.Get(a.apiBaseURL + "/sites/list")
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

	result := make([]string, len(sites))
	for i, site := range sites {
		if siteName, ok := site.(string); ok {
			result[i] = siteName
		}
	}

	return result, nil
}

// GetVersions 获取版本列表
func (a *App) GetVersions(name string) ([]Version, error) {
	url := fmt.Sprintf("%s/sites/versions?name=%s", a.apiBaseURL, name)
	resp, err := http.Get(url)
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
	resp, err := http.Post(a.apiBaseURL+"/sites/rollback", "application/json", strings.NewReader(string(data)))
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
