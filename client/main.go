//go:build cli
// +build cli

package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

func main() {
	flag.Parse()

	// 加载配置
	config, err := LoadConfig()
	if err != nil {
		fmt.Printf("加载配置失败: %v\n", err)
		os.Exit(1)
	}

	apiBaseURL := config.ServerURL
	username := config.Username
	password := config.Password

	args := flag.Args()
	if len(args) < 1 {
		printUsage()
		os.Exit(1)
	}

	command := args[0]

	switch command {
	case "config":
		handleConfig(args[1:])
	case "create":
		handleCreate(apiBaseURL, username, password, args[1:])
	case "delete":
		handleDelete(apiBaseURL, username, password, args[1:])
	case "deploy":
		handleDeploy(apiBaseURL, username, password, config, args[1:])
	case "deploy-full":
		handleDeployFull(apiBaseURL, username, password, config, args[1:])
	case "deploy-inc":
		handleDeployIncremental(apiBaseURL, username, password, config, args[1:])
	case "list":
		handleList(apiBaseURL, username, password, args[1:])
	case "versions":
		handleVersions(apiBaseURL, username, password, args[1:])
	case "rollback":
		handleRollback(apiBaseURL, username, password, args[1:])
	case "pull":
		handlePull(apiBaseURL, username, password, config, args[1:])
	case "help":
		printUsage()
	default:
		fmt.Printf("未知命令: %s\n", command)
		printUsage()
		os.Exit(1)
	}
}

// addAuthToRequest 如果配置了用户名/密码，则添加到请求头
func addAuthToRequest(req *http.Request, username, password string) {
	if username != "" && password != "" {
		req.Header.Set("X-Username", username)
		req.Header.Set("X-Password", password)
	}
}

// postJSON 发送POST请求（带JSON body和可选的认证信息）
func postJSON(url string, body []byte, username, password string) (*http.Response, error) {
	req, err := http.NewRequest("POST", url, strings.NewReader(string(body)))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	addAuthToRequest(req, username, password)

	client := &http.Client{}
	return client.Do(req)
}

// getJSON 发送GET请求（带可选的认证信息）
func getJSON(url string, username, password string) (*http.Response, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	addAuthToRequest(req, username, password)

	client := &http.Client{}
	return client.Do(req)
}

func printUsage() {
	fmt.Println("AI原型快速部署工具 - 命令行客户端")
	fmt.Println("\n用法:")
	fmt.Println("  deploy-cli <command> [arguments]")
	fmt.Println("\n命令:")
	fmt.Println("  config                管理配置（服务器地址、API密钥、网站目录）")
	fmt.Println("  create <name>          创建新网站")
	fmt.Println("  delete <name>          删除网站")
	fmt.Println("  deploy [name] [dir]    部署网站（智能选择增量或全量，自动匹配网站）")
	fmt.Println("  deploy-full [name]     全量部署网站")
	fmt.Println("  deploy-inc [name]      增量部署网站")
	fmt.Println("  list                   列出所有网站")
	fmt.Println("  versions <name>        查看网站版本历史")
	fmt.Println("  rollback <name> <hash> 回滚到指定版本")
	fmt.Println("  pull [name]            从服务器覆盖本地（自动匹配网站）")
	fmt.Println("  help                   显示帮助信息")
	fmt.Println("\n示例:")
	fmt.Println("  deploy-cli config set server http://192.168.1.100:8080/api")
	fmt.Println("  deploy-cli config set username admin")
	fmt.Println("  deploy-cli config set password admin123")
	fmt.Println("  deploy-cli config set site my-prototype ./dist")
	fmt.Println("  deploy-cli config get")
	fmt.Println("  deploy-cli create my-prototype")
	fmt.Println("  deploy-cli deploy                    # 自动匹配并部署")
	fmt.Println("  deploy-cli deploy my-prototype        # 指定网站部署")
	fmt.Println("  deploy-cli deploy my-prototype ./dist # 指定目录部署")
	fmt.Println("  deploy-cli versions my-prototype")
	fmt.Println("  deploy-cli rollback my-prototype abc123")
	fmt.Println("  deploy-cli pull my-prototype             # 从服务器覆盖本地")
}

func handleCreate(apiBaseURL, username, password string, args []string) {
	if len(args) < 1 {
		fmt.Println("错误: 请提供网站名称")
		fmt.Println("用法: deploy-cli create <name>")
		os.Exit(1)
	}

	name := args[0]

	payload := map[string]string{
		"name": name,
	}

	data, err := json.Marshal(payload)
	if err != nil {
		fmt.Printf("错误: %v\n", err)
		os.Exit(1)
	}

	resp, err := postJSON(apiBaseURL+"/sites/create", data, username, password)
	if err != nil {
		fmt.Printf("错误: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK {
		fmt.Printf("创建失败: %s\n", string(body))
		os.Exit(1)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		fmt.Printf("创建成功\n")
		return
	}

	fmt.Println("✓ 网站创建成功!")
	if domain, ok := result["domain"].(string); ok {
		fmt.Printf("  域名: %s\n", domain)
	}
	if path, ok := result["path"].(string); ok {
		fmt.Printf("  路径: %s\n", path)
	}
}

func handleDelete(apiBaseURL, username, password string, args []string) {
	if len(args) < 1 {
		fmt.Println("错误: 请提供网站名称")
		fmt.Println("用法: deploy-cli delete <name>")
		os.Exit(1)
	}

	name := args[0]

	payload := map[string]string{
		"name": name,
	}

	data, err := json.Marshal(payload)
	if err != nil {
		fmt.Printf("错误: %v\n", err)
		os.Exit(1)
	}

	// 确认删除
	fmt.Printf("确定要删除网站 '%s' 吗？(y/N): ", name)
	reader := bufio.NewReader(os.Stdin)
	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(input)
	if strings.ToLower(input) != "y" {
		fmt.Println("取消删除")
		return
	}

	resp, err := postJSON(apiBaseURL+"/sites/delete", data, username, password)
	if err != nil {
		fmt.Printf("错误: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK {
		fmt.Printf("删除失败: %s\n", string(body))
		os.Exit(1)
	}

	fmt.Println("✓ 网站删除成功!")
}

func handleList(apiBaseURL, username, password string, args []string) {
	resp, err := getJSON(apiBaseURL+"/sites/list", username, password)
	if err != nil {
		fmt.Printf("错误: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK {
		fmt.Printf("获取列表失败: %s\n", string(body))
		os.Exit(1)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		fmt.Printf("解析响应失败: %v\n", err)
		os.Exit(1)
	}

	sites, ok := result["sites"].([]interface{})
	if !ok {
		fmt.Println("没有网站")
		return
	}

	fmt.Println("\n网站列表:")
	fmt.Println(strings.Repeat("-", 50))
	for i, site := range sites {
		if siteName, ok := site.(string); ok {
			fmt.Printf("%d. %s\n", i+1, siteName)
		}
	}
	fmt.Println(strings.Repeat("-", 50))
}

func handleVersions(apiBaseURL, username, password string, args []string) {
	if len(args) < 1 {
		fmt.Println("错误: 请提供网站名称")
		fmt.Println("用法: deploy-cli versions <name>")
		os.Exit(1)
	}

	name := args[0]

	url := fmt.Sprintf("%s/sites/versions?name=%s", apiBaseURL, name)
	resp, err := getJSON(url, username, password)
	if err != nil {
		fmt.Printf("错误: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK {
		fmt.Printf("获取版本失败: %s\n", string(body))
		os.Exit(1)
	}

	var versions []map[string]interface{}
	if err := json.Unmarshal(body, &versions); err != nil {
		fmt.Printf("解析响应失败: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("\n网站 '%s' 的版本历史:\n", name)
	fmt.Println(strings.Repeat("-", 80))
	if len(versions) == 0 {
		fmt.Println("暂无版本记录")
	} else {
		for i, v := range versions {
			fmt.Printf("%d. %s\n", i+1, v["hash"])
			fmt.Printf("   提交: %s\n", v["message"])
			fmt.Printf("   作者: %s\n", v["author"])
			if date, ok := v["date"].(string); ok {
				fmt.Printf("   日期: %s\n", date)
			}
			fmt.Println()
		}
	}
	fmt.Println(strings.Repeat("-", 80))
}

func handleRollback(apiBaseURL, username, password string, args []string) {
	if len(args) < 2 {
		fmt.Println("错误: 请提供网站名称和版本哈希")
		fmt.Println("用法: deploy-cli rollback <name> <hash> [message]")
		os.Exit(1)
	}

	name := args[0]
	hash := args[1]
	message := "回滚版本"

	if len(args) > 2 {
		message = strings.Join(args[2:], " ")
	}

	payload := map[string]string{
		"name":    name,
		"hash":    hash,
		"message": message,
	}

	data, err := json.Marshal(payload)
	if err != nil {
		fmt.Printf("错误: %v\n", err)
		os.Exit(1)
	}

	// 确认回滚
	fmt.Printf("确定要回滚网站 '%s' 到版本 '%s' 吗？(y/N): ", name, hash)
	reader := bufio.NewReader(os.Stdin)
	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(input)
	if strings.ToLower(input) != "y" {
		fmt.Println("取消回滚")
		return
	}

	resp, err := postJSON(apiBaseURL+"/sites/rollback", data, username, password)
	if err != nil {
		fmt.Printf("错误: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK {
		fmt.Printf("回滚失败: %s\n", string(body))
		os.Exit(1)
	}

	fmt.Println("✓ 回滚成功!")
}

// handleDeploy 智能部署（自动选择增量或全量）
func handleDeploy(apiBaseURL, username, password string, config *ClientConfig, args []string) {
	var name, dirPath string
	message := "更新部署"

	// 如果没有提供网站名称，尝试根据当前目录自动匹配
	if len(args) < 1 {
		matchedSites := findMatchingSites(config)
		if len(matchedSites) == 0 {
			fmt.Println("错误: 无法确定要部署的网站")
			fmt.Println("\n请指定网站名称：")
			fmt.Println("  deploy-cli deploy <site-name>")
			fmt.Println("\n或者先配置网站目录：")
			fmt.Println("  deploy-cli config set site <site-name> <directory>")
			os.Exit(1)
		} else if len(matchedSites) > 1 {
			fmt.Println("错误: 当前目录匹配到多个网站，请明确指定：")
			for _, site := range matchedSites {
				fmt.Printf("  - %s (%s)\n", site, config.SitePaths[site])
			}
			fmt.Println("\n使用方法: deploy-cli deploy <site-name>")
			os.Exit(1)
		}

		// 唯一匹配
		name = matchedSites[0]
		dirPath = config.SitePaths[name]
		fmt.Printf("自动匹配到网站: %s\n", name)
	} else {
		name = args[0]

		// 确定目录路径
		if len(args) >= 2 {
			// 命令行指定了目录
			dirPath = args[1]
			if len(args) > 2 {
				message = strings.Join(args[2:], " ")
			}
		} else {
			// 从配置读取目录
			var ok bool
			dirPath, ok = config.SitePaths[name]
			if !ok || dirPath == "" {
				fmt.Printf("错误: 网站 '%s' 未配置发布目录\n", name)
				fmt.Println("\n请先配置网站目录，或者直接指定目录：")
				fmt.Printf("  deploy-cli config set site %s ./dist\n", name)
				fmt.Printf("  deploy-cli deploy %s ./dist\n", name)
				os.Exit(1)
			}
		}
	}

	// 检查目录是否存在
	if _, err := os.Stat(dirPath); os.IsNotExist(err) {
		fmt.Printf("错误: 目录不存在: %s\n", dirPath)
		os.Exit(1)
	}

	// 创建部署器
	deployer := NewDeployer(apiBaseURL, name)

	// 执行智能部署
	if err := deployer.Deploy(dirPath, message); err != nil {
		fmt.Printf("部署失败: %v\n", err)
		os.Exit(1)
	}
}

// findMatchingSites 根据当前目录查找匹配的网站
// 匹配规则：
// 1. 优先匹配：当前目录是网站路径的子路径
// 2. 次优匹配：网站路径是当前目录的子路径
func findMatchingSites(config *ClientConfig) []string {
	// 获取当前工作目录
	currentDir, err := os.Getwd()
	if err != nil {
		return nil
	}

	// 转换为绝对路径（处理相对路径情况）
	absCurrentDir, err := filepath.Abs(currentDir)
	if err != nil {
		return nil
	}

	var primaryMatches []string // 当前目录是网站路径的子路径
	var secondaryMatches []string // 网站路径是当前目录的子路径

	for siteName, sitePath := range config.SitePaths {
		// 转换网站路径为绝对路径
		absSitePath, err := filepath.Abs(sitePath)
		if err != nil {
			continue
		}

		// 检查当前目录是否是网站路径的子路径（优先匹配）
		if strings.HasPrefix(absCurrentDir, absSitePath) {
			primaryMatches = append(primaryMatches, siteName)
		} else if strings.HasPrefix(absSitePath, absCurrentDir) {
			// 检查网站路径是否是当前目录的子路径（次优匹配）
			secondaryMatches = append(secondaryMatches, siteName)
		}
	}

	// 优先返回主匹配，如果没有则返回次匹配
	if len(primaryMatches) > 0 {
		return primaryMatches
	}
	return secondaryMatches
}

// handleDeployFull 全量部署
func handleDeployFull(apiBaseURL, username, password string, config *ClientConfig, args []string) {
	var name, dirPath string
	message := "全量部署"

	// 如果没有提供网站名称，尝试根据当前目录自动匹配
	if len(args) < 1 {
		matchedSites := findMatchingSites(config)
		if len(matchedSites) == 0 {
			fmt.Println("错误: 无法确定要部署的网站")
			fmt.Println("\n请指定网站名称：")
			fmt.Println("  deploy-cli deploy-full <site-name>")
			fmt.Println("\n或者先配置网站目录：")
			fmt.Println("  deploy-cli config set site <site-name> <directory>")
			os.Exit(1)
		} else if len(matchedSites) > 1 {
			fmt.Println("错误: 当前目录匹配到多个网站，请明确指定：")
			for _, site := range matchedSites {
				fmt.Printf("  - %s (%s)\n", site, config.SitePaths[site])
			}
			fmt.Println("\n使用方法: deploy-cli deploy-full <site-name>")
			os.Exit(1)
		}

		// 唯一匹配
		name = matchedSites[0]
		dirPath = config.SitePaths[name]
		fmt.Printf("自动匹配到网站: %s\n", name)
	} else {
		name = args[0]

		// 确定目录路径
		if len(args) >= 2 {
			dirPath = args[1]
			if len(args) > 2 {
				message = strings.Join(args[2:], " ")
			}
		} else {
			var ok bool
			dirPath, ok = config.SitePaths[name]
			if !ok || dirPath == "" {
				fmt.Printf("错误: 网站 '%s' 未配置发布目录\n", name)
				fmt.Println("\n请先配置网站目录，或者直接指定目录：")
				fmt.Printf("  deploy-cli config set site %s ./dist\n", name)
				fmt.Printf("  deploy-cli deploy-full %s ./dist\n", name)
				os.Exit(1)
			}
		}
	}

	// 检查目录是否存在
	if _, err := os.Stat(dirPath); os.IsNotExist(err) {
		fmt.Printf("错误: 目录不存在: %s\n", dirPath)
		os.Exit(1)
	}

	// 创建部署器
	deployer := NewDeployer(apiBaseURL, name)

	// 执行全量部署
	if err := deployer.DeployFull(dirPath, message); err != nil {
		fmt.Printf("部署失败: %v\n", err)
		os.Exit(1)
	}
}

// handleDeployIncremental 增量部署
func handleDeployIncremental(apiBaseURL, username, password string, config *ClientConfig, args []string) {
	var name, dirPath string
	message := "增量部署"

	// 如果没有提供网站名称，尝试根据当前目录自动匹配
	if len(args) < 1 {
		matchedSites := findMatchingSites(config)
		if len(matchedSites) == 0 {
			fmt.Println("错误: 无法确定要部署的网站")
			fmt.Println("\n请指定网站名称：")
			fmt.Println("  deploy-cli deploy-inc <site-name>")
			fmt.Println("\n或者先配置网站目录：")
			fmt.Println("  deploy-cli config set site <site-name> <directory>")
			os.Exit(1)
		} else if len(matchedSites) > 1 {
			fmt.Println("错误: 当前目录匹配到多个网站，请明确指定：")
			for _, site := range matchedSites {
				fmt.Printf("  - %s (%s)\n", site, config.SitePaths[site])
			}
			fmt.Println("\n使用方法: deploy-cli deploy-inc <site-name>")
			os.Exit(1)
		}

		// 唯一匹配
		name = matchedSites[0]
		dirPath = config.SitePaths[name]
		fmt.Printf("自动匹配到网站: %s\n", name)
	} else {
		name = args[0]

		// 确定目录路径
		if len(args) >= 2 {
			dirPath = args[1]
			if len(args) > 2 {
				message = strings.Join(args[2:], " ")
			}
		} else {
			var ok bool
			dirPath, ok = config.SitePaths[name]
			if !ok || dirPath == "" {
				fmt.Printf("错误: 网站 '%s' 未配置发布目录\n", name)
				fmt.Println("\n请先配置网站目录，或者直接指定目录：")
				fmt.Printf("  deploy-cli config set site %s ./dist\n", name)
				fmt.Printf("  deploy-cli deploy-inc %s ./dist\n", name)
				os.Exit(1)
			}
		}
	}

	// 检查目录是否存在
	if _, err := os.Stat(dirPath); os.IsNotExist(err) {
		fmt.Printf("错误: 目录不存在: %s\n", dirPath)
		os.Exit(1)
	}

	// 创建部署器
	deployer := NewDeployer(apiBaseURL, name)

	// 执行增量部署
	if err := deployer.DeployIncremental(dirPath, message); err != nil {
		fmt.Printf("部署失败: %v\n", err)
		os.Exit(1)
	}
}

// handleConfig 处理配置命令
func handleConfig(args []string) {
	if len(args) < 1 {
		printConfigHelp()
		os.Exit(1)
	}

	subCommand := args[0]

	switch subCommand {
	case "set":
		handleConfigSet(args[1:])
	case "get":
		handleConfigGet()
	case "remove":
		handleConfigRemove(args[1:])
	default:
		fmt.Printf("未知配置命令: %s\n", subCommand)
		printConfigHelp()
		os.Exit(1)
	}
}

// printConfigHelp 打印配置命令帮助
func printConfigHelp() {
	fmt.Println("配置管理命令")
	fmt.Println("\n用法:")
	fmt.Println("  deploy-cli config <sub-command> [arguments]")
	fmt.Println("\n子命令:")
	fmt.Println("  set server <url>      设置服务器地址")
	fmt.Println("  set username <name>   设置用户名")
	fmt.Println("  set password <pwd>    设置密码")
	fmt.Println("  set site <name> <dir> 设置网站发布目录")
	fmt.Println("  get                   查看当前配置")
	fmt.Println("  remove site <name>    移除网站发布目录")
	fmt.Println("\n示例:")
	fmt.Println("  deploy-cli config set server http://192.168.1.100:8080/api")
	fmt.Println("  deploy-cli config set username admin")
	fmt.Println("  deploy-cli config set password admin123")
	fmt.Println("  deploy-cli config set site my-prototype ./dist")
	fmt.Println("  deploy-cli config set site my-project /path/to/dist")
	fmt.Println("  deploy-cli config get")
	fmt.Println("  deploy-cli config remove site my-prototype")
}

// handleConfigSet 设置配置值
func handleConfigSet(args []string) {
	if len(args) < 1 {
		fmt.Println("错误: 缺少参数")
		fmt.Println("用法: deploy-cli config set <server|username|password|site> <value>")
		os.Exit(1)
	}

	key := args[0]

	// 加载当前配置
	config, err := LoadConfig()
	if err != nil {
		// 如果配置文件不存在，创建新的
		config = &ClientConfig{
			ServerURL: "http://localhost:8080/api",
			SitePaths: make(map[string]string),
		}
	}

	// 确保SitePaths初始化
	if config.SitePaths == nil {
		config.SitePaths = make(map[string]string)
	}

	switch key {
	case "server":
		if len(args) < 2 {
			fmt.Println("错误: 请提供服务器地址")
			fmt.Println("用法: deploy-cli config set server <url>")
			os.Exit(1)
		}
		config.ServerURL = args[1]
		fmt.Printf("✓ 服务器地址已设置为: %s\n", args[1])

	case "username":
		if len(args) < 2 {
			fmt.Println("错误: 请提供用户名")
			fmt.Println("用法: deploy-cli config set username <name>")
			os.Exit(1)
		}
		config.Username = args[1]
		fmt.Printf("✓ 用户名已设置为: %s\n", args[1])

	case "password":
		if len(args) < 2 {
			fmt.Println("错误: 请提供密码")
			fmt.Println("用法: deploy-cli config set password <pwd>")
			os.Exit(1)
		}
		config.Password = args[1]
		fmt.Println("✓ 密码已设置")

	case "site":
		if len(args) < 3 {
			fmt.Println("错误: 请提供网站名称和目录路径")
			fmt.Println("用法: deploy-cli config set site <name> <directory>")
			os.Exit(1)
		}
		siteName := args[1]
		sitePath := args[2]

		// 检查目录是否存在
		if _, err := os.Stat(sitePath); os.IsNotExist(err) {
			fmt.Printf("警告: 目录不存在: %s\n", sitePath)
			fmt.Print("是否仍然保存此路径？(y/N): ")
			reader := bufio.NewReader(os.Stdin)
			input, _ := reader.ReadString('\n')
			input = strings.TrimSpace(input)
			if strings.ToLower(input) != "y" {
				fmt.Println("取消操作")
				return
			}
		}

		// 转换为绝对路径
		absPath, err := filepath.Abs(sitePath)
		if err == nil {
			sitePath = absPath
		}

		config.SitePaths[siteName] = sitePath
		fmt.Printf("✓ 网站 '%s' 的发布目录已设置为: %s\n", siteName, sitePath)
		fmt.Println("  现在可以使用 'deploy-cli deploy %s' 直接部署\n", siteName)

	default:
		fmt.Printf("错误: 未知的配置项 '%s'\n", key)
		fmt.Println("支持的配置项: server, username, password, site")
		os.Exit(1)
	}

	// 保存配置
	if err := SaveConfig(config); err != nil {
		fmt.Printf("错误: 保存配置失败: %v\n", err)
		os.Exit(1)
	}
}

// handleConfigRemove 移除配置项
func handleConfigRemove(args []string) {
	if len(args) < 1 {
		fmt.Println("错误: 缺少参数")
		fmt.Println("用法: deploy-cli config remove <site> <name>")
		os.Exit(1)
	}

	key := args[0]

	if key != "site" {
		fmt.Printf("错误: 不支持的移除操作 '%s'\n", key)
		fmt.Println("当前只支持 'remove site <name>'")
		os.Exit(1)
	}

	if len(args) < 2 {
		fmt.Println("错误: 请提供网站名称")
		fmt.Println("用法: deploy-cli config remove site <name>")
		os.Exit(1)
	}

	siteName := args[1]

	// 加载当前配置
	config, err := LoadConfig()
	if err != nil {
		fmt.Printf("错误: 加载配置失败: %v\n", err)
		os.Exit(1)
	}

	// 检查网站是否在配置中
	if _, exists := config.SitePaths[siteName]; !exists {
		fmt.Printf("错误: 网站 '%s' 未配置发布目录\n", siteName)
		os.Exit(1)
	}

	// 删除配置
	delete(config.SitePaths, siteName)
	fmt.Printf("✓ 已移除网站 '%s' 的发布目录配置\n", siteName)

	// 保存配置
	if err := SaveConfig(config); err != nil {
		fmt.Printf("错误: 保存配置失败: %v\n", err)
		os.Exit(1)
	}
}

// handleConfigGet 查看当前配置
func handleConfigGet() {
	config, err := LoadConfig()
	if err != nil {
		fmt.Printf("错误: 加载配置失败: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("\n当前配置:")
	fmt.Println(strings.Repeat("-", 60))
	fmt.Printf("服务器地址: %s\n", config.ServerURL)

	if config.Username == "" {
		fmt.Println("用户名:     （未设置）")
	} else {
		fmt.Printf("用户名:     %s\n", config.Username)
	}

	if config.Password == "" {
		fmt.Println("密码:       （未设置）")
	} else {
		fmt.Println("密码:       ******")
	}

	// 显示网站目录配置
	if len(config.SitePaths) > 0 {
		fmt.Println("\n网站发布目录:")
		for name, path := range config.SitePaths {
			fmt.Printf("  %-20s -> %s\n", name, path)
		}
	} else {
		fmt.Println("\n网站发布目录: (未配置)")
	}

	fmt.Println(strings.Repeat("-", 60))
	fmt.Printf("\n配置文件位置: %s\n", getConfigPath())
	fmt.Println("\n提示:")
	fmt.Println("  使用 'deploy-cli config set site <name> <dir>' 配置网站目录")
	fmt.Println("  使用 'deploy-cli config remove site <name>' 移除网站目录")
	fmt.Println("  使用 'deploy-cli deploy <name>' 直接部署已配置的网站")
	fmt.Println("  使用 'deploy-cli pull <name>' 从服务器覆盖本地文件")
}

// handlePull 从服务器覆盖本地
func handlePull(apiBaseURL, username, password string, config *ClientConfig, args []string) {
	var name, dirPath string

	// 如果没有提供网站名称，尝试根据当前目录自动匹配
	if len(args) < 1 {
		matchedSites := findMatchingSites(config)
		if len(matchedSites) == 0 {
			fmt.Println("错误: 无法确定要下载的网站")
			fmt.Println("\n请指定网站名称：")
			fmt.Println("  deploy-cli pull <site-name>")
			fmt.Println("\n或者先配置网站目录：")
			fmt.Println("  deploy-cli config set site <site-name> <directory>")
			os.Exit(1)
		} else if len(matchedSites) > 1 {
			fmt.Println("错误: 当前目录匹配到多个网站，请明确指定：")
			for _, site := range matchedSites {
				fmt.Printf("  - %s (%s)\n", site, config.SitePaths[site])
			}
			fmt.Println("\n使用方法: deploy-cli pull <site-name>")
			os.Exit(1)
		}

		// 唯一匹配
		name = matchedSites[0]
		dirPath = config.SitePaths[name]
		fmt.Printf("自动匹配到网站: %s\n", name)
	} else {
		name = args[0]

		// 确定目录路径
		if len(args) >= 2 {
			// 命令行指定了目录
			dirPath = args[1]
		} else {
			// 从配置读取目录
			var ok bool
			dirPath, ok = config.SitePaths[name]
			if !ok || dirPath == "" {
				fmt.Printf("错误: 网站 '%s' 未配置发布目录\n", name)
				fmt.Println("\n请先配置网站目录，或者直接指定目录：")
				fmt.Printf("  deploy-cli config set site %s ./dist\n", name)
				fmt.Printf("  deploy-cli pull %s ./dist\n", name)
				os.Exit(1)
			}
		}
	}

	// 确认操作
	fmt.Printf("确定要从服务器下载网站 '%s' 并覆盖本地目录 '%s' 吗？(y/N): ", name, dirPath)
	reader := bufio.NewReader(os.Stdin)
	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(input)
	if strings.ToLower(input) != "y" {
		fmt.Println("取消操作")
		return
	}

	// 创建部署器
	deployer := NewDeployer(apiBaseURL, name)

	// 执行下载并覆盖
	if err := deployer.PullFromServer(dirPath); err != nil {
		fmt.Printf("下载失败: %v\n", err)
		os.Exit(1)
	}
}
