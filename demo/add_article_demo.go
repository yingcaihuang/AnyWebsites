package main

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// ContentRequest 上传内容请求结构
type ContentRequest struct {
	Title       string `json:"title"`
	Description string `json:"description,omitempty"`
	HTMLContent string `json:"html_content"`
	IsPublic    bool   `json:"is_public"`
}

// ContentResponse 上传内容响应结构
type ContentResponse struct {
	Message string `json:"message"`
	ID      string `json:"id"`
	URL     string `json:"url"`
}

// ErrorResponse 错误响应结构
type ErrorResponse struct {
	Error string `json:"error"`
}

const (
	// API 配置
	BaseURL = "https://localhost:8443"
	APIKey  = "1e278ff1-881a-47e6-ad8c-f779e715"
)

func main() {
	fmt.Println("🚀 AnyWebsites API Demo - 添加文章")
	fmt.Println("=====================================")

	// 创建 HTTP 客户端（跳过 SSL 验证，仅用于本地测试）
	client := &http.Client{
		Timeout: 30 * time.Second,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}

	// 示例文章内容
	articles := []ContentRequest{
		{
			Title:       "我的第一篇技术博客",
			Description: "关于 Go 语言开发的入门指南",
			HTMLContent: `<!DOCTYPE html>
<html lang="zh-CN">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Go 语言开发入门指南</title>
    <style>
        body { font-family: 'Microsoft YaHei', sans-serif; line-height: 1.6; margin: 40px; background: #f5f5f5; }
        .container { max-width: 800px; margin: 0 auto; background: white; padding: 30px; border-radius: 10px; box-shadow: 0 2px 10px rgba(0,0,0,0.1); }
        h1 { color: #2c3e50; border-bottom: 3px solid #3498db; padding-bottom: 10px; }
        h2 { color: #34495e; margin-top: 30px; }
        code { background: #f8f9fa; padding: 2px 6px; border-radius: 3px; font-family: 'Courier New', monospace; }
        pre { background: #2d3748; color: #e2e8f0; padding: 20px; border-radius: 8px; overflow-x: auto; }
        .highlight { background: #fff3cd; padding: 15px; border-left: 4px solid #ffc107; margin: 20px 0; }
    </style>
</head>
<body>
    <div class="container">
        <h1>🚀 Go 语言开发入门指南</h1>
        
        <div class="highlight">
            <strong>💡 提示：</strong> 本文适合有一定编程基础的开发者快速上手 Go 语言。
        </div>

        <h2>📖 什么是 Go 语言？</h2>
        <p>Go（也称为 Golang）是由 Google 开发的开源编程语言，具有以下特点：</p>
        <ul>
            <li><strong>简洁高效</strong>：语法简单，编译速度快</li>
            <li><strong>并发支持</strong>：原生支持 goroutine 和 channel</li>
            <li><strong>跨平台</strong>：支持多种操作系统和架构</li>
            <li><strong>强类型</strong>：静态类型检查，减少运行时错误</li>
        </ul>

        <h2>🛠️ 环境搭建</h2>
        <p>首先需要安装 Go 开发环境：</p>
        <pre><code># 下载并安装 Go
# 访问 https://golang.org/dl/ 下载对应版本

# 验证安装
go version

# 设置工作目录
mkdir ~/go-projects
cd ~/go-projects</code></pre>

        <h2>👨‍💻 第一个 Go 程序</h2>
        <p>创建 <code>hello.go</code> 文件：</p>
        <pre><code>package main

import "fmt"

func main() {
    fmt.Println("Hello, 世界！")
    fmt.Println("欢迎学习 Go 语言！")
}</code></pre>

        <p>运行程序：</p>
        <pre><code>go run hello.go</code></pre>

        <h2>🔧 基础语法示例</h2>
        <pre><code>package main

import (
    "fmt"
    "time"
)

func main() {
    // 变量声明
    var name string = "Go 开发者"
    age := 25  // 短变量声明
    
    // 条件语句
    if age >= 18 {
        fmt.Printf("%s 已成年\n", name)
    }
    
    // 循环
    for i := 1; i <= 3; i++ {
        fmt.Printf("第 %d 次循环\n", i)
        time.Sleep(time.Second)
    }
    
    // 函数调用
    result := add(10, 20)
    fmt.Printf("10 + 20 = %d\n", result)
}

func add(a, b int) int {
    return a + b
}</code></pre>

        <h2>🎯 下一步学习</h2>
        <ul>
            <li>学习 Go 的包管理（go mod）</li>
            <li>掌握 goroutine 和 channel</li>
            <li>了解接口（interface）的使用</li>
            <li>实践 Web 开发（gin 框架）</li>
            <li>数据库操作（GORM）</li>
        </ul>

        <div class="highlight">
            <strong>🎉 恭喜！</strong> 您已经迈出了 Go 语言学习的第一步。继续加油！
        </div>
        
        <hr style="margin: 30px 0; border: none; border-top: 1px solid #eee;">
        <p style="text-align: center; color: #666; font-size: 14px;">
            📝 发布时间：2025-06-24 | 🏷️ 标签：Go语言, 编程入门, 后端开发
        </p>
    </div>
</body>
</html>`,
			IsPublic: true,
		},
		{
			Title:       "Docker 容器化部署实践",
			Description: "从零开始学习 Docker 容器化技术",
			HTMLContent: `<!DOCTYPE html>
<html lang="zh-CN">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Docker 容器化部署实践</title>
    <style>
        body { font-family: 'Microsoft YaHei', sans-serif; line-height: 1.6; margin: 40px; background: linear-gradient(135deg, #667eea 0%, #764ba2 100%); min-height: 100vh; }
        .container { max-width: 900px; margin: 0 auto; background: white; padding: 40px; border-radius: 15px; box-shadow: 0 10px 30px rgba(0,0,0,0.2); }
        h1 { color: #2c3e50; text-align: center; margin-bottom: 30px; font-size: 2.5em; }
        h2 { color: #3498db; border-left: 4px solid #3498db; padding-left: 15px; margin-top: 35px; }
        .docker-logo { text-align: center; font-size: 4em; margin: 20px 0; }
        code { background: #f1f2f6; padding: 3px 8px; border-radius: 4px; font-family: 'Fira Code', monospace; color: #e74c3c; }
        pre { background: #2f3542; color: #f1f2f6; padding: 25px; border-radius: 10px; overflow-x: auto; border-left: 5px solid #3498db; }
        .tip { background: #e8f5e8; border: 1px solid #4caf50; padding: 20px; border-radius: 8px; margin: 20px 0; }
        .warning { background: #fff3e0; border: 1px solid #ff9800; padding: 20px; border-radius: 8px; margin: 20px 0; }
        ul li { margin: 8px 0; }
        .command-box { background: #34495e; color: #ecf0f1; padding: 15px; border-radius: 8px; margin: 15px 0; }
    </style>
</head>
<body>
    <div class="container">
        <div class="docker-logo">🐳</div>
        <h1>Docker 容器化部署实践</h1>
        
        <div class="tip">
            <strong>🎯 学习目标：</strong> 通过本教程，您将掌握 Docker 的基本概念、常用命令和实际部署技巧。
        </div>

        <h2>🤔 什么是 Docker？</h2>
        <p>Docker 是一个开源的容器化平台，它可以让开发者将应用程序及其依赖项打包到一个轻量级、可移植的容器中。</p>
        
        <h2>🚀 Docker 的优势</h2>
        <ul>
            <li><strong>环境一致性</strong>：开发、测试、生产环境完全一致</li>
            <li><strong>快速部署</strong>：秒级启动，比虚拟机快得多</li>
            <li><strong>资源高效</strong>：共享宿主机内核，资源占用少</li>
            <li><strong>易于扩展</strong>：支持水平扩展和负载均衡</li>
            <li><strong>版本控制</strong>：镜像支持版本管理</li>
        </ul>

        <h2>📦 核心概念</h2>
        <ul>
            <li><strong>镜像（Image）</strong>：应用程序的只读模板</li>
            <li><strong>容器（Container）</strong>：镜像的运行实例</li>
            <li><strong>Dockerfile</strong>：构建镜像的脚本文件</li>
            <li><strong>仓库（Repository）</strong>：存储镜像的地方</li>
        </ul>

        <h2>🛠️ 安装 Docker</h2>
        <div class="command-box">
            <strong>Ubuntu/Debian:</strong><br>
            curl -fsSL https://get.docker.com -o get-docker.sh<br>
            sudo sh get-docker.sh
        </div>
        
        <div class="command-box">
            <strong>验证安装:</strong><br>
            docker --version<br>
            docker run hello-world
        </div>

        <h2>📝 编写 Dockerfile</h2>
        <p>以一个简单的 Node.js 应用为例：</p>
        <pre><code># 使用官方 Node.js 镜像
FROM node:18-alpine

# 设置工作目录
WORKDIR /app

# 复制 package.json 文件
COPY package*.json ./

# 安装依赖
RUN npm install --only=production

# 复制应用代码
COPY . .

# 暴露端口
EXPOSE 3000

# 定义启动命令
CMD ["npm", "start"]</code></pre>

        <h2>🔨 构建和运行</h2>
        <pre><code># 构建镜像
docker build -t my-app:1.0 .

# 运行容器
docker run -d -p 3000:3000 --name my-app-container my-app:1.0

# 查看运行中的容器
docker ps

# 查看日志
docker logs my-app-container</code></pre>

        <h2>🐙 Docker Compose</h2>
        <p>对于多服务应用，使用 <code>docker-compose.yml</code>：</p>
        <pre><code>version: '3.8'

services:
  web:
    build: .
    ports:
      - "3000:3000"
    environment:
      - NODE_ENV=production
    depends_on:
      - db
      - redis

  db:
    image: postgres:15
    environment:
      POSTGRES_DB: myapp
      POSTGRES_USER: user
      POSTGRES_PASSWORD: password
    volumes:
      - postgres_data:/var/lib/postgresql/data

  redis:
    image: redis:7-alpine
    ports:
      - "6379:6379"

volumes:
  postgres_data:</code></pre>

        <div class="command-box">
            <strong>启动所有服务:</strong><br>
            docker-compose up -d
        </div>

        <h2>🎯 最佳实践</h2>
        <div class="tip">
            <ul>
                <li>使用 <code>.dockerignore</code> 文件排除不必要的文件</li>
                <li>多阶段构建减小镜像体积</li>
                <li>使用非 root 用户运行容器</li>
                <li>合理使用缓存层，优化构建速度</li>
                <li>定期清理无用的镜像和容器</li>
            </ul>
        </div>

        <div class="warning">
            <strong>⚠️ 注意事项：</strong>
            <ul>
                <li>生产环境中要设置资源限制</li>
                <li>敏感信息使用 Docker Secrets</li>
                <li>定期更新基础镜像以修复安全漏洞</li>
                <li>监控容器的健康状态</li>
            </ul>
        </div>

        <h2>🎉 总结</h2>
        <p>Docker 容器化技术已经成为现代应用部署的标准。通过本教程的学习，您应该已经掌握了：</p>
        <ul>
            <li>Docker 的基本概念和优势</li>
            <li>如何编写 Dockerfile</li>
            <li>容器的构建、运行和管理</li>
            <li>Docker Compose 的使用</li>
            <li>生产环境的最佳实践</li>
        </ul>
        
        <hr style="margin: 40px 0; border: none; border-top: 2px solid #ecf0f1;">
        <p style="text-align: center; color: #7f8c8d; font-size: 14px;">
            🐳 Docker 让部署变得简单 | 📅 2025-06-24 | 🏷️ Docker, 容器化, DevOps
        </p>
    </div>
</body>
</html>`,
			IsPublic: true,
		},
	}

	// 上传每篇文章
	for i, article := range articles {
		fmt.Printf("\n📝 正在上传第 %d 篇文章: %s\n", i+1, article.Title)
		
		contentID, viewURL, err := uploadContent(client, article)
		if err != nil {
			fmt.Printf("❌ 上传失败: %v\n", err)
			continue
		}
		
		fmt.Printf("✅ 上传成功!\n")
		fmt.Printf("   📋 内容ID: %s\n", contentID)
		fmt.Printf("   🔗 访问链接: %s%s\n", BaseURL, viewURL)
		fmt.Printf("   🌐 浏览器访问: %s%s\n", BaseURL, viewURL)
		
		// 等待一秒再上传下一篇
		time.Sleep(1 * time.Second)
	}
	
	fmt.Println("\n🎉 所有文章上传完成！")
	fmt.Println("💡 您可以通过上面的链接访问您的文章。")
}

// uploadContent 上传内容到 AnyWebsites
func uploadContent(client *http.Client, content ContentRequest) (string, string, error) {
	// 将请求转换为 JSON
	jsonData, err := json.Marshal(content)
	if err != nil {
		return "", "", fmt.Errorf("JSON 编码失败: %v", err)
	}

	// 创建 HTTP 请求
	req, err := http.NewRequest("POST", BaseURL+"/api/content/upload", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", "", fmt.Errorf("创建请求失败: %v", err)
	}

	// 设置请求头
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-API-Key", APIKey)

	// 发送请求
	resp, err := client.Do(req)
	if err != nil {
		return "", "", fmt.Errorf("发送请求失败: %v", err)
	}
	defer resp.Body.Close()

	// 读取响应
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", "", fmt.Errorf("读取响应失败: %v", err)
	}

	// 检查状态码
	if resp.StatusCode != http.StatusCreated {
		var errorResp ErrorResponse
		if err := json.Unmarshal(body, &errorResp); err == nil {
			return "", "", fmt.Errorf("API 错误 (%d): %s", resp.StatusCode, errorResp.Error)
		}
		return "", "", fmt.Errorf("HTTP 错误 (%d): %s", resp.StatusCode, string(body))
	}

	// 解析成功响应
	var contentResp ContentResponse
	if err := json.Unmarshal(body, &contentResp); err != nil {
		return "", "", fmt.Errorf("解析响应失败: %v", err)
	}

	return contentResp.ID, contentResp.URL, nil
}
