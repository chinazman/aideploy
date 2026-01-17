package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"aideploy/server"
)

func main() {
	configPath := flag.String("config", "config.json", "配置文件路径")
	createConfig := flag.Bool("init", false, "创建默认配置文件")
	flag.Parse()

	// 如果需要创建默认配置
	if *createConfig {
		if err := createDefaultConfig(*configPath); err != nil {
			fmt.Printf("创建配置文件失败: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("已创建默认配置文件: %s\n", *configPath)
		fmt.Println("请编辑配置文件后重新启动服务器")
		return
	}

	// 读取配置
	config, err := loadConfig(*configPath)
	if err != nil {
		fmt.Printf("加载配置失败: %v\n", err)
		fmt.Println("提示: 使用 -init 参数创建默认配置文件")
		os.Exit(1)
	}

	// 创建并启动服务器
	srv := server.NewDeployServer(config)
	if err := srv.Start(); err != nil {
		fmt.Printf("服务器启动失败: %v\n", err)
		os.Exit(1)
	}
}

// Config 配置结构（临时定义，用于加载）
type Config struct {
	BaseDomain       string `json:"base_domain"`
	WebRoot          string `json:"web_root"`
	Mode             string `json:"mode"`
	SingleDomain     string `json:"single_domain"`
	Port             int    `json:"port"`
	EnableVersioning bool   `json:"enable_versioning"`
}

// loadConfig 加载配置文件
func loadConfig(path string) (server.Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return server.Config{}, err
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return server.Config{}, err
	}

	return server.Config{
		BaseDomain:       cfg.BaseDomain,
		WebRoot:          cfg.WebRoot,
		Mode:             cfg.Mode,
		SingleDomain:     cfg.SingleDomain,
		Port:             cfg.Port,
		EnableVersioning: cfg.EnableVersioning,
	}, nil
}

// createDefaultConfig 创建默认配置文件
func createDefaultConfig(path string) error {
	// 获取当前目录
	wd, err := os.Getwd()
	if err != nil {
		return err
	}

	webRoot := filepath.Join(wd, "websites")

	cfg := Config{
		BaseDomain:       "example.com",
		WebRoot:          webRoot,
		Mode:             "subdomain",
		SingleDomain:     "",
		Port:             8080,
		EnableVersioning: true,
	}

	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0644)
}
