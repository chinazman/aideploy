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

	// 检查配置文件是否存在
	if _, err := os.Stat(*configPath); os.IsNotExist(err) {
		// 配置文件不存在,自动创建默认配置
		fmt.Printf("配置文件不存在,正在创建默认配置: %s\n", *configPath)
		if err := createDefaultConfig(*configPath); err != nil {
			fmt.Printf("创建配置文件失败: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("已创建默认配置文件: %s\n", *configPath)
		fmt.Println("默认管理员: admin / admin123")
		fmt.Println("服务器将使用默认配置启动...")
	}

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
	srv := server.NewDeployServer(config, *configPath)
	if err := srv.Start(); err != nil {
		fmt.Printf("服务器启动失败: %v\n", err)
		os.Exit(1)
	}
}

// Config 配置结构（临时定义，用于加载）
type Config struct {
	BaseDomain       string                 `json:"base_domain"`
	WebRoot          string                 `json:"web_root"`
	Mode             string                 `json:"mode"`
	SingleDomain     string                 `json:"single_domain"`
	Port             int                    `json:"port"`
	EnableVersioning bool                   `json:"enable_versioning"`
	APIKey           string                 `json:"api_key,omitempty"`
	Sites            map[string]server.Site `json:"sites"`
	Users            map[string]server.User `json:"users"`
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

	// 初始化空的 maps
	if cfg.Sites == nil {
		cfg.Sites = make(map[string]server.Site)
	}
	if cfg.Users == nil {
		cfg.Users = make(map[string]server.User)
	}

	return server.Config{
		BaseDomain:       cfg.BaseDomain,
		WebRoot:          cfg.WebRoot,
		Mode:             cfg.Mode,
		SingleDomain:     cfg.SingleDomain,
		Port:             cfg.Port,
		EnableVersioning: cfg.EnableVersioning,
		APIKey:           cfg.APIKey,
		Sites:            cfg.Sites,
		Users:            cfg.Users,
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

	// 创建默认管理员用户
	defaultUsers := map[string]server.User{
		"admin": {
			Name:     "admin",
			Password: "admin123",
			IsAdmin:  true,
		},
	}

	cfg := Config{
		BaseDomain:       "localhost",
		WebRoot:          webRoot,
		Mode:             "path",
		SingleDomain:     "",
		Port:             8080,
		EnableVersioning: true,
		APIKey:           "",
		Sites:            make(map[string]server.Site),
		Users:            defaultUsers,
	}

	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0644)
}
