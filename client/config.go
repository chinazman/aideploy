package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// ClientConfig 客户端配置
type ClientConfig struct {
	ServerURL  string            `json:"server_url"`
	Username   string            `json:"username"`
	Password   string            `json:"password"`
	SitePaths  map[string]string `json:"site_paths"` // 网站名称 -> 本地发布目录映射
}

// LoadConfig 加载客户端配置
// 从 ~/.aideploy/config.json 加载配置，如果不存在则使用默认值
func LoadConfig() (*ClientConfig, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("无法获取用户目录: %v", err)
	}

	configPath := filepath.Join(homeDir, ".aideploy", "config.json")

	// 尝试读取配置文件
	data, err := os.ReadFile(configPath)
	if err != nil {
		// 配置文件不存在，返回默认配置
		return &ClientConfig{
			ServerURL: "http://localhost:8080/api",
			SitePaths: make(map[string]string),
		}, nil
	}

	// 解析配置文件
	var config ClientConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("解析配置文件失败: %v", err)
	}

	// 如果 server_url 为空，使用默认值
	if config.ServerURL == "" {
		config.ServerURL = "http://localhost:8080/api"
	}

	// 确保 SitePaths 不为 nil
	if config.SitePaths == nil {
		config.SitePaths = make(map[string]string)
	}

	return &config, nil
}

// GetAPIBaseURL 获取 API 基础 URL
// 这是一个便捷方法，用于获取配置中的 server_url
func GetAPIBaseURL() (string, error) {
	config, err := LoadConfig()
	if err != nil {
		return "", err
	}
	return config.ServerURL, nil
}

// SaveConfig 保存客户端配置
// 将配置保存到 ~/.aideploy/config.json
func SaveConfig(config *ClientConfig) error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("无法获取用户目录: %v", err)
	}

	// 确保配置目录存在
	configDir := filepath.Join(homeDir, ".aideploy")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("创建配置目录失败: %v", err)
	}

	// 序列化配置
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("序列化配置失败: %v", err)
	}

	// 写入配置文件
	configPath := filepath.Join(configDir, "config.json")
	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return fmt.Errorf("写入配置文件失败: %v", err)
	}

	return nil
}

// getConfigPath 获取配置文件路径
func getConfigPath() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "未知"
	}
	return filepath.Join(homeDir, ".aideploy", "config.json")
}
