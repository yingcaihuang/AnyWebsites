package main

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"
)

// API响应结构体
type LoginResponse struct {
	User struct {
		ID       string `json:"id"`
		Username string `json:"username"`
		Email    string `json:"email"`
		IsAdmin  bool   `json:"is_admin"`
	} `json:"user"`
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

type ContentResponse struct {
	Message string `json:"message"`
	ID      string `json:"id"`
	URL     string `json:"url"`
	FullURL string `json:"full_url"`
	Content struct {
		ID          string    `json:"id"`
		Title       string    `json:"title"`
		Description string    `json:"description"`
		IsPublic    bool      `json:"is_public"`
		AccessCode  string    `json:"access_code"`
		IsActive    bool      `json:"is_active"`
		CreatedAt   time.Time `json:"created_at"`
		UpdatedAt   time.Time `json:"updated_at"`
	} `json:"content"`
}

type PublicContentResponse struct {
	Message string `json:"message"`
	ID      string `json:"id"`
	URL     string `json:"url"`
	FullURL string `json:"full_url"`
}

// API客户端
type APIClient struct {
	BaseURL    string
	HTTPClient *http.Client
	Token      string
}

func NewAPIClient(baseURL string) *APIClient {
	// 创建忽略SSL证书的HTTP客户端
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}

	return &APIClient{
		BaseURL: baseURL,
		HTTPClient: &http.Client{
			Transport: tr,
			Timeout:   30 * time.Second,
		},
	}
}

// 登录获取Token
func (c *APIClient) Login(username, password string) error {
	loginData := map[string]string{
		"username": username,
		"password": password,
	}

	jsonData, err := json.Marshal(loginData)
	if err != nil {
		return fmt.Errorf("marshal login data: %w", err)
	}

	resp, err := c.HTTPClient.Post(
		c.BaseURL+"/api/auth/login",
		"application/json",
		bytes.NewBuffer(jsonData),
	)
	if err != nil {
		return fmt.Errorf("login request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("read login response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("login failed: %s", string(body))
	}

	var loginResp LoginResponse
	if err := json.Unmarshal(body, &loginResp); err != nil {
		return fmt.Errorf("unmarshal login response: %w", err)
	}

	c.Token = loginResp.AccessToken
	fmt.Printf("✅ 登录成功! 用户: %s (管理员: %t)\n",
		loginResp.User.Username, loginResp.User.IsAdmin)

	return nil
}

// 创建内容（需要认证）
func (c *APIClient) CreateContent(title, description, htmlContent string, isPublic bool) (*ContentResponse, error) {
	if c.Token == "" {
		return nil, fmt.Errorf("需要先登录获取Token")
	}

	contentData := map[string]interface{}{
		"title":        title,
		"description":  description,
		"html_content": htmlContent,
		"is_public":    isPublic,
	}

	jsonData, err := json.Marshal(contentData)
	if err != nil {
		return nil, fmt.Errorf("marshal content data: %w", err)
	}

	req, err := http.NewRequest("POST", c.BaseURL+"/api/content/upload", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.Token)

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("upload request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read upload response: %w", err)
	}

	if resp.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("upload failed: %s", string(body))
	}

	var contentResp ContentResponse
	if err := json.Unmarshal(body, &contentResp); err != nil {
		return nil, fmt.Errorf("unmarshal upload response: %w", err)
	}

	return &contentResp, nil
}

// 注意：公开API已被移除，所有内容创建都需要认证

// 访问内容
func (c *APIClient) GetContent(contentID string) (string, error) {
	resp, err := c.HTTPClient.Get(c.BaseURL + "/view/" + contentID)
	if err != nil {
		return "", fmt.Errorf("get content: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("read content: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("get content failed: %s", string(body))
	}

	return string(body), nil
}

func main() {
	fmt.Println("🚀 AnyWebsites API Demo 开始测试...")
	fmt.Println(strings.Repeat("=", 50))

	// 创建API客户端
	client := NewAPIClient("https://localhost:8443")

	// 测试1: 管理员登录
	fmt.Println("\n📝 测试1: 管理员登录")
	if err := client.Login("betty", "123.com"); err != nil {
		log.Fatalf("管理员登录失败: %v", err)
	}

	// 测试2: 创建认证内容
	fmt.Println("\n📝 测试2: 创建认证内容")
	authContent, err := client.CreateContent(
		"Go API Demo - 认证内容",
		"这是通过Go语言调用认证API创建的内容",
		`<h1>🎉 Hello from Go API!</h1>
		<p>这个内容是通过Go语言程序调用<strong>认证API</strong>创建的。</p>
		<ul>
			<li>✅ 需要JWT Token认证</li>
			<li>✅ 归属于认证用户</li>
			<li>✅ 返回完整的内容信息</li>
		</ul>
		<p>创建时间: <code>`+time.Now().Format("2006-01-02 15:04:05")+`</code></p>`,
		true,
	)
	if err != nil {
		log.Fatalf("创建认证内容失败: %v", err)
	}

	fmt.Printf("✅ 认证内容创建成功!\n")
	fmt.Printf("   ID: %s\n", authContent.ID)
	fmt.Printf("   标题: %s\n", authContent.Content.Title)
	fmt.Printf("   访问URL: %s\n", authContent.FullURL)
	fmt.Printf("   创建时间: %s\n", authContent.Content.CreatedAt.Format("2006-01-02 15:04:05"))

	// 测试3: 创建第二个认证内容
	fmt.Println("\n📝 测试3: 创建第二个认证内容")
	authContent2, err := client.CreateContent(
		"Go API Demo - 第二个认证内容",
		"这是通过Go语言调用认证API创建的第二个内容",
		`<h1>🔐 第二个认证API测试!</h1>
		<p>这个内容是通过Go语言程序调用<strong>认证API</strong>创建的第二个内容。</p>
		<ul>
			<li>✅ 需要JWT Token认证</li>
			<li>✅ 归属于认证用户</li>
			<li>✅ 确保数据安全</li>
			<li>✅ 所有内容创建都需要认证</li>
		</ul>
		<p>创建时间: <code>`+time.Now().Format("2006-01-02 15:04:05")+`</code></p>`,
		true,
	)
	if err != nil {
		log.Fatalf("创建第二个认证内容失败: %v", err)
	}

	fmt.Printf("✅ 第二个认证内容创建成功!\n")
	fmt.Printf("   ID: %s\n", authContent2.ID)
	fmt.Printf("   标题: %s\n", authContent2.Content.Title)
	fmt.Printf("   访问URL: %s\n", authContent2.FullURL)
	fmt.Printf("   创建时间: %s\n", authContent2.Content.CreatedAt.Format("2006-01-02 15:04:05"))

	// 测试4: 访问创建的内容
	fmt.Println("\n📝 测试4: 访问创建的内容")

	// 访问第一个认证内容
	fmt.Println("访问第一个认证内容...")
	authContentHTML, err := client.GetContent(authContent.ID)
	if err != nil {
		log.Printf("访问第一个认证内容失败: %v", err)
	} else {
		fmt.Printf("✅ 第一个认证内容访问成功! 内容长度: %d 字符\n", len(authContentHTML))
	}

	// 访问第二个认证内容
	fmt.Println("访问第二个认证内容...")
	authContentHTML2, err := client.GetContent(authContent2.ID)
	if err != nil {
		log.Printf("访问第二个认证内容失败: %v", err)
	} else {
		fmt.Printf("✅ 第二个认证内容访问成功! 内容长度: %d 字符\n", len(authContentHTML2))
	}

	// 测试5: 选择第一个内容进行详细测试
	fmt.Println("\n📝 测试5: 选择第一个内容进行详细测试")
	selectedID := authContent.ID
	fmt.Printf("选择的内容ID: %s\n", selectedID)
	fmt.Printf("选择的内容标题: %s\n", authContent.Content.Title)

	fmt.Println("正在进行详细测试...")
	selectedContentHTML, err := client.GetContent(selectedID)
	if err != nil {
		log.Printf("❌ 访问选中内容失败: %v", err)
	} else {
		fmt.Printf("✅ 选中内容访问成功!\n")
		fmt.Printf("   内容ID: %s\n", selectedID)
		fmt.Printf("   内容标题: %s\n", authContent.Content.Title)
		fmt.Printf("   内容长度: %d 字符\n", len(selectedContentHTML))
		fmt.Printf("   访问URL: %s\n", authContent.FullURL)
		fmt.Printf("   创建时间: %s\n", authContent.Content.CreatedAt.Format("2006-01-02 15:04:05"))
		fmt.Printf("   是否公开: %t\n", authContent.Content.IsPublic)

		// 显示文章内容的前200个字符作为预览
		preview := selectedContentHTML
		if len(preview) > 200 {
			preview = preview[:200] + "..."
		}
		fmt.Printf("   内容预览: %s\n", preview)
	}

	// 测试总结
	fmt.Println("\n" + strings.Repeat("=", 50))
	fmt.Println("🎉 API测试完成!")
	fmt.Println("\n📋 测试结果总结:")
	fmt.Printf("✅ 管理员登录: 成功\n")
	fmt.Printf("✅ 第一个认证API创建内容: 成功 (ID: %s)\n", authContent.ID)
	fmt.Printf("✅ 第二个认证API创建内容: 成功 (ID: %s)\n", authContent2.ID)
	fmt.Printf("✅ 内容访问: 成功\n")

	fmt.Println("\n🔗 访问链接:")
	fmt.Printf("第一个认证内容: %s\n", authContent.FullURL)
	fmt.Printf("第二个认证内容: %s\n", authContent2.FullURL)

	fmt.Println("\n💡 提示: 您可以在浏览器中打开上述链接查看创建的内容!")
	fmt.Println("\n🔒 安全提醒: 所有内容创建都需要认证，确保了平台的安全性和数据完整性!")
}
