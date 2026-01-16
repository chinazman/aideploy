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
	"time"
)

const (
	apiBaseURL = "http://localhost:8080/api"
)

func main() {
	flag.Parse()

	args := flag.Args()
	if len(args) < 1 {
		printUsage()
		os.Exit(1)
	}

	command := args[0]

	switch command {
	case "create":
		handleCreate(args[1:])
	case "delete":
		handleDelete(args[1:])
	case "deploy":
		handleDeploy(args[1:])
	case "list":
		handleList(args[1:])
	case "versions":
		handleVersions(args[1:])
	case "rollback":
		handleRollback(args[1:])
	case "help":
		printUsage()
	default:
		fmt.Printf("未知命令: %s\n", command)
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println("AI原型快速部署工具 - 命令行客户端")
	fmt.Println("\n用法:")
	fmt.Println("  deploy-cli <command> [arguments]")
	fmt.Println("\n命令:")
	fmt.Println("  create <name>          创建新网站")
	fmt.Println("  delete <name>          删除网站")
	fmt.Println("  deploy <name> <file>   部署网站（上传文件）")
	fmt.Println("  list                   列出所有网站")
	fmt.Println("  versions <name>        查看网站版本历史")
	fmt.Println("  rollback <name> <hash> 回滚到指定版本")
	fmt.Println("  help                   显示帮助信息")
	fmt.Println("\n示例:")
	fmt.Println("  deploy-cli create my-prototype")
	fmt.Println("  deploy-cli deploy my-prototype prototype.html")
	fmt.Println("  deploy-cli versions my-prototype")
	fmt.Println("  deploy-cli rollback my-prototype abc123")
}

func handleCreate(args []string) {
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

	resp, err := http.Post(apiBaseURL+"/sites/create", "application/json", strings.NewReader(string(data)))
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

func handleDelete(args []string) {
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

	resp, err := http.Post(apiBaseURL+"/sites/delete", "application/json", strings.NewReader(string(data)))
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

func handleDeploy(args []string) {
	if len(args) < 2 {
		fmt.Println("错误: 请提供网站名称和文件路径")
		fmt.Println("用法: deploy-cli deploy <name> <file> [message]")
		os.Exit(1)
	}

	name := args[0]
	filePath := args[1]
	message := "更新部署"

	if len(args) > 2 {
		message = strings.Join(args[2:], " ")
	}

	// 检查文件是否存在
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		fmt.Printf("错误: 文件不存在: %s\n", filePath)
		os.Exit(1)
	}

	// 打开文件
	file, err := os.Open(filePath)
	if err != nil {
		fmt.Printf("错误: %v\n", err)
		os.Exit(1)
	}
	defer file.Close()

	// 创建multipart请求
	var requestBody strings.Builder
	boundary := "----WebKitFormBoundary" + time.Now().Format("20060102150405")

	requestBody.WriteString("--" + boundary + "\r\n")
	requestBody.WriteString(fmt.Sprintf(`Content-Disposition: form-data; name="name"%s\r\n\r\n`, "\r\n"))
	requestBody.WriteString(name + "\r\n")

	requestBody.WriteString("--" + boundary + "\r\n")
	requestBody.WriteString(fmt.Sprintf(`Content-Disposition: form-data; name="message"%s\r\n\r\n`, "\r\n"))
	requestBody.WriteString(message + "\r\n")

	requestBody.WriteString("--" + boundary + "\r\n")
	requestBody.WriteString(fmt.Sprintf(`Content-Disposition: form-data; name="file"; filename="%s"%s`, filepath.Base(filePath), "\r\n"))
	requestBody.WriteString("Content-Type: application/octet-stream\r\n\r\n")

	// 这里需要将文件内容写入，为简化使用http.PostForm
	file.Seek(0, 0)

	// 使用更简单的方法
	req, err := http.NewRequest("POST", apiBaseURL+"/sites/deploy", nil)
	if err != nil {
		fmt.Printf("错误: %v\n", err)
		os.Exit(1)
	}

	// 创建临时文件来构建multipart请求
	tempFile, err := os.CreateTemp("", "deploy-*.tmp")
	if err != nil {
		fmt.Printf("错误: %v\n", err)
		os.Exit(1)
	}
	defer os.Remove(tempFile.Name())
	defer tempFile.Close()

	// 写入multipart表单
	writer := bufio.NewWriter(tempFile)
	fmt.Fprintf(writer, "--%s\r\n", boundary)
	fmt.Fprintf(writer, `Content-Disposition: form-data; name="name"%s\r\n\r\n`, "\r\n")
	fmt.Fprintf(writer, "%s\r\n", name)

	fmt.Fprintf(writer, "--%s\r\n", boundary)
	fmt.Fprintf(writer, `Content-Disposition: form-data; name="message"%s\r\n\r\n`, "\r\n")
	fmt.Fprintf(writer, "%s\r\n", message)

	fmt.Fprintf(writer, "--%s\r\n", boundary)
	fmt.Fprintf(writer, `Content-Disposition: form-data; name="file"; filename="%s"%s`, filepath.Base(filePath), "\r\n")
	fmt.Fprintf(writer, "Content-Type: application/octet-stream\r\n\r\n")
	writer.Flush()

	// 复制文件内容
	file.Seek(0, 0)
	io.Copy(tempFile, file)

	fmt.Fprintf(tempFile, "\r\n--%s--\r\n", boundary)
	tempFile.Close()

	// 重新打开临时文件读取内容
	tempFile2, _ := os.Open(tempFile.Name())
	defer tempFile2.Close()

	req, err = http.NewRequest("POST", apiBaseURL+"/sites/deploy", tempFile2)
	if err != nil {
		fmt.Printf("错误: %v\n", err)
		os.Exit(1)
	}

	req.Header.Set("Content-Type", "multipart/form-data; boundary="+boundary)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("错误: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK {
		fmt.Printf("部署失败: %s\n", string(body))
		os.Exit(1)
	}

	fmt.Println("✓ 部署成功!")
}

func handleList(args []string) {
	resp, err := http.Get(apiBaseURL + "/sites/list")
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

func handleVersions(args []string) {
	if len(args) < 1 {
		fmt.Println("错误: 请提供网站名称")
		fmt.Println("用法: deploy-cli versions <name>")
		os.Exit(1)
	}

	name := args[0]

	url := fmt.Sprintf("%s/sites/versions?name=%s", apiBaseURL, name)
	resp, err := http.Get(url)
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

func handleRollback(args []string) {
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

	resp, err := http.Post(apiBaseURL+"/sites/rollback", "application/json", strings.NewReader(string(data)))
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
