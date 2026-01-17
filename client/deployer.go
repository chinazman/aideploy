package main

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"crypto/md5"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// Deployer 部署器
type Deployer struct {
	serverURL   string
	apiKey      string
	siteName    string
	trackingDir string // 跟踪文件目录
}

// FileStatus 文件状态
type FileStatus struct {
	Path         string    `json:"path"`
	Hash         string    `json:"hash"`
	Size         int64     `json:"size"`
	ModTime      time.Time `json:"mod_time"`
	LastDeployed time.Time `json:"last_deployed"`
}

// TrackingData 跟踪数据
type TrackingData struct {
	SiteName   string       `json:"site_name"`
	LastSync   time.Time    `json:"last_sync"`
	Files      []FileStatus `json:"files"`
}

// NewDeployer 创建部署器
func NewDeployer(serverURL, siteName string) *Deployer {
	homeDir, _ := os.UserHomeDir()
	trackingDir := filepath.Join(homeDir, ".aideploy", "tracking")

	// 加载配置以获取API密钥
	config, err := LoadConfig()
	apiKey := ""
	if err == nil {
		apiKey = config.APIKey
	}

	return &Deployer{
		serverURL:   serverURL,
		apiKey:      apiKey,
		siteName:    siteName,
		trackingDir: trackingDir,
	}
}

// DeployFull 全量部署
func (d *Deployer) DeployFull(sitePath, message string) error {
	fmt.Println("开始全量部署...")

	// 扫描当前文件状态（用于更新跟踪信息）
	currentFiles, err := d.scanFiles(sitePath)
	if err != nil {
		return fmt.Errorf("扫描文件失败: %v", err)
	}

	// 创建临时打包文件
	tempFile, err := os.CreateTemp("", "deploy-full-*.tar.gz")
	if err != nil {
		return fmt.Errorf("创建临时文件失败: %v", err)
	}
	tempPath := tempFile.Name()
	defer os.Remove(tempPath)
	tempFile.Close()

	// 打包整个目录
	fmt.Printf("正在打包目录: %s\n", sitePath)
	if err := d.createPackage(sitePath, tempPath, nil); err != nil {
		return fmt.Errorf("打包失败: %v", err)
	}

	// 显示包文件大小
	if info, err := os.Stat(tempPath); err == nil {
		fmt.Printf("打包完成，包大小: %.2f MB\n", float64(info.Size())/1024/1024)
	}

	// 上传到服务器
	url := fmt.Sprintf("%s/sites/deploy-full", d.serverURL)
	if err := d.uploadPackage(url, tempPath, message); err != nil {
		return fmt.Errorf("上传失败: %v", err)
	}

	// 更新跟踪信息（保存所有文件状态）
	if err := d.updateTracking(sitePath, currentFiles); err != nil {
		fmt.Printf("警告: 更新跟踪信息失败: %v\n", err)
	}

	fmt.Println("✓ 全量部署成功!")
	return nil
}

// DeployIncremental 增量部署
func (d *Deployer) DeployIncremental(sitePath, message string) error {
	fmt.Println("开始增量部署...")

	// 获取当前文件状态
	currentFiles, err := d.scanFiles(sitePath)
	if err != nil {
		return fmt.Errorf("扫描文件失败: %v", err)
	}

	// 加载之前的跟踪信息
	trackingData, err := d.loadTracking()
	if err != nil {
		// 如果没有跟踪信息，回退到全量部署
		fmt.Println("未找到跟踪信息，执行全量部署...")
		return d.DeployFull(sitePath, message)
	}

	// 找出变更的文件
	changedFiles := d.findChangedFiles(currentFiles, trackingData.Files)
	deletedFiles := d.findDeletedFiles(currentFiles, trackingData.Files)

	if len(changedFiles) == 0 && len(deletedFiles) == 0 {
		fmt.Println("没有文件变更，无需部署")
		return nil
	}

	fmt.Printf("检测到 %d 个变更文件, %d 个删除文件\n", len(changedFiles), len(deletedFiles))

	// 创建增量包
	tempFile, err := os.CreateTemp("", "deploy-inc-*.tar.gz")
	if err != nil {
		return fmt.Errorf("创建临时文件失败: %v", err)
	}
	tempPath := tempFile.Name()
	defer os.Remove(tempPath)
	tempFile.Close()

	// 打包变更的文件
	if err := d.createPackage(sitePath, tempPath, changedFiles); err != nil {
		return fmt.Errorf("打包失败: %v", err)
	}

	// 上传到服务器
	url := fmt.Sprintf("%s/sites/deploy-incremental", d.serverURL)
	if err := d.uploadPackage(url, tempPath, message); err != nil {
		return fmt.Errorf("上传失败: %v", err)
	}

	// 更新跟踪信息
	if err := d.updateTracking(sitePath, currentFiles); err != nil {
		fmt.Printf("警告: 更新跟踪信息失败: %v\n", err)
	}

	fmt.Printf("✓ 增量部署成功! (变更: %d 文件)\n", len(changedFiles))
	return nil
}

// Deploy 智能部署(自动选择增量或全量)
func (d *Deployer) Deploy(sitePath, message string) error {
	// 检查是否有跟踪信息
	_, err := d.loadTracking()
	if err != nil {
		// 没有跟踪信息，使用全量部署
		return d.DeployFull(sitePath, message)
	}

	// 有跟踪信息，使用增量部署
	return d.DeployIncremental(sitePath, message)
}

// scanFiles 扫描目录中的所有文件
func (d *Deployer) scanFiles(sitePath string) ([]FileStatus, error) {
	var files []FileStatus

	err := filepath.Walk(sitePath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// 跳过隐藏文件和目录
		if strings.HasPrefix(filepath.Base(path), ".") {
			if info.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		// 跳过目录
		if info.IsDir() {
			return nil
		}

		// 计算相对路径
		relPath, err := filepath.Rel(sitePath, path)
		if err != nil {
			return err
		}

		// 计算文件哈希
		hash, err := d.calculateHash(path)
		if err != nil {
			return err
		}

		files = append(files, FileStatus{
			Path:    relPath,
			Hash:    hash,
			Size:    info.Size(),
			ModTime: info.ModTime(),
		})

		return nil
	})

	return files, err
}

// calculateHash 计算文件哈希
func (d *Deployer) calculateHash(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	hash := md5.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", err
	}

	return fmt.Sprintf("%x", hash.Sum(nil)), nil
}

// findChangedFiles 找出变更的文件
func (d *Deployer) findChangedFiles(current []FileStatus, previous []FileStatus) []FileStatus {
	prevMap := make(map[string]FileStatus)
	for _, f := range previous {
		prevMap[f.Path] = f
	}

	var changed []FileStatus
	for _, f := range current {
		prev, exists := prevMap[f.Path]
		if !exists || prev.Hash != f.Hash {
			changed = append(changed, f)
		}
	}

	return changed
}

// findDeletedFiles 找出删除的文件
func (d *Deployer) findDeletedFiles(current []FileStatus, previous []FileStatus) []string {
	currMap := make(map[string]bool)
	for _, f := range current {
		currMap[f.Path] = true
	}

	var deleted []string
	for _, f := range previous {
		if !currMap[f.Path] {
			deleted = append(deleted, f.Path)
		}
	}

	return deleted
}

// createPackage 创建部署包
func (d *Deployer) createPackage(sitePath, packagePath string, files []FileStatus) error {
	file, err := os.Create(packagePath)
	if err != nil {
		return err
	}
	defer file.Close()

	gzWriter := gzip.NewWriter(file)
	defer gzWriter.Close()

	tarWriter := tar.NewWriter(gzWriter)
	defer tarWriter.Close()

	fileCount := 0
	dirCount := 0

	// 如果没有指定文件列表，打包所有文件
	if len(files) == 0 {
		err := filepath.Walk(sitePath, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			// 跳过根目录本身
			if path == sitePath {
				return nil
			}

			// 跳过隐藏文件和目录（以"."开头的文件或目录）
			if strings.HasPrefix(filepath.Base(path), ".") {
				if info.IsDir() {
					return filepath.SkipDir
				}
				return nil
			}

			// 统计
			if info.IsDir() {
				dirCount++
			} else {
				fileCount++
			}

			return d.addToTar(tarWriter, path, sitePath, info)
		})

		if err != nil {
			return err
		}

		fmt.Printf("已打包: %d 个文件, %d 个目录\n", fileCount, dirCount)
		return nil
	}

	// 打包指定的文件
	for _, f := range files {
		fullPath := filepath.Join(sitePath, f.Path)
		info, err := os.Stat(fullPath)
		if err != nil {
			continue
		}
		if err := d.addToTar(tarWriter, fullPath, sitePath, info); err != nil {
			return err
		}
		fileCount++
	}

	fmt.Printf("已打包: %d 个文件, %d 个目录\n", fileCount, dirCount)
	return nil
}

// addToTar 添加文件到tar包
func (d *Deployer) addToTar(tarWriter *tar.Writer, filePath, baseDir string, info os.FileInfo) error {
	// 获取相对路径
	relPath, err := filepath.Rel(baseDir, filePath)
	if err != nil {
		return err
	}

	// 将路径分隔符统一转换为正斜杠（tar格式标准）
	relPath = filepath.ToSlash(relPath)

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
		file, err := os.Open(filePath)
		if err != nil {
			return err
		}
		defer file.Close()

		_, err = io.Copy(tarWriter, file)
		return err
	}

	return nil
}

// uploadPackage 上传部署包
func (d *Deployer) uploadPackage(url, packagePath, message string) error {
	// 打开包文件
	file, err := os.Open(packagePath)
	if err != nil {
		return fmt.Errorf("打开包文件失败: %v", err)
	}
	defer file.Close()

	// 创建multipart请求body
	var requestBody bytes.Buffer
	writer := multipart.NewWriter(&requestBody)

	// 添加name字段
	if err := writer.WriteField("name", d.siteName); err != nil {
		return fmt.Errorf("写入name字段失败: %v", err)
	}

	// 添加message字段
	if err := writer.WriteField("message", message); err != nil {
		return fmt.Errorf("写入message字段失败: %v", err)
	}

	// 添加package文件
	part, err := writer.CreateFormFile("package", filepath.Base(packagePath))
	if err != nil {
		return fmt.Errorf("创建文件字段失败: %v", err)
	}

	if _, err := io.Copy(part, file); err != nil {
		return fmt.Errorf("复制文件内容失败: %v", err)
	}

	// 关闭writer以完成multipart写入
	if err := writer.Close(); err != nil {
		return fmt.Errorf("关闭writer失败: %v", err)
	}

	// 创建HTTP请求
	req, err := http.NewRequest("POST", url, &requestBody)
	if err != nil {
		return fmt.Errorf("创建请求失败: %v", err)
	}

	// 设置Content-Type头
	req.Header.Set("Content-Type", writer.FormDataContentType())

	// 如果配置了API密钥，添加到请求头
	if d.apiKey != "" {
		req.Header.Set("X-API-Key", d.apiKey)
	}

	// 发送请求
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("发送请求失败: %v", err)
	}
	defer resp.Body.Close()

	// 读取响应
	body, _ := io.ReadAll(resp.Body)

	// 检查响应状态
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("服务器返回错误: %s", string(body))
	}

	return nil
}

// loadTracking 加载跟踪信息
func (d *Deployer) loadTracking() (*TrackingData, error) {
	trackingPath := d.getTrackingPath()
	data, err := os.ReadFile(trackingPath)
	if err != nil {
		return nil, err
	}

	var tracking TrackingData
	if err := json.Unmarshal(data, &tracking); err != nil {
		return nil, err
	}

	return &tracking, nil
}

// updateTracking 更新跟踪信息
func (d *Deployer) updateTracking(sitePath string, files []FileStatus) error {
	// 确保跟踪目录存在
	if err := os.MkdirAll(d.trackingDir, 0755); err != nil {
		return err
	}

	tracking := TrackingData{
		SiteName: d.siteName,
		LastSync: time.Now(),
		Files:    files,
	}

	data, err := json.MarshalIndent(tracking, "", "  ")
	if err != nil {
		return err
	}

	trackingPath := d.getTrackingPath()
	return os.WriteFile(trackingPath, data, 0644)
}

// getTrackingPath 获取跟踪文件路径
func (d *Deployer) getTrackingPath() string {
	return filepath.Join(d.trackingDir, d.siteName+".json")
}
