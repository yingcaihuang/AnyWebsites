package utils

import (
	"net"
	"strings"

	"github.com/gin-gonic/gin"
)

// GetRealClientIP 获取真实的客户端IP地址
// 支持从代理服务器（如nginx）传递的真实IP
func GetRealClientIP(c *gin.Context) string {
	// 1. 首先检查 X-Real-IP 头部（nginx proxy_set_header X-Real-IP $remote_addr;）
	if realIP := c.GetHeader("X-Real-IP"); realIP != "" {
		if ip := parseIP(realIP); ip != "" {
			return ip
		}
	}

	// 2. 检查 X-Forwarded-For 头部（可能包含多个IP，第一个是真实客户端IP）
	if forwardedFor := c.GetHeader("X-Forwarded-For"); forwardedFor != "" {
		// X-Forwarded-For 格式: client, proxy1, proxy2
		ips := strings.Split(forwardedFor, ",")
		for _, ip := range ips {
			ip = strings.TrimSpace(ip)
			if parsedIP := parseIP(ip); parsedIP != "" && !isPrivateIP(parsedIP) {
				return parsedIP
			}
		}
	}

	// 3. 检查 CF-Connecting-IP 头部（Cloudflare）
	if cfIP := c.GetHeader("CF-Connecting-IP"); cfIP != "" {
		if ip := parseIP(cfIP); ip != "" {
			return ip
		}
	}

	// 4. 检查 X-Forwarded 头部
	if forwarded := c.GetHeader("X-Forwarded"); forwarded != "" {
		// 解析 X-Forwarded 头部，格式可能是: for=192.0.2.60;proto=http;by=203.0.113.43
		if ip := parseForwardedHeader(forwarded); ip != "" {
			return ip
		}
	}

	// 5. 检查 Forwarded 头部（RFC 7239 标准）
	if forwarded := c.GetHeader("Forwarded"); forwarded != "" {
		if ip := parseForwardedHeader(forwarded); ip != "" {
			return ip
		}
	}

	// 6. 最后使用 Gin 的默认 ClientIP 方法
	return c.ClientIP()
}

// parseIP 解析并验证IP地址
func parseIP(ipStr string) string {
	ip := strings.TrimSpace(ipStr)
	if net.ParseIP(ip) != nil {
		return ip
	}
	return ""
}

// isPrivateIP 检查是否为私有IP地址
func isPrivateIP(ipStr string) bool {
	ip := net.ParseIP(ipStr)
	if ip == nil {
		return false
	}

	// IPv4 私有地址范围
	private4 := []string{
		"10.0.0.0/8",
		"172.16.0.0/12",
		"192.168.0.0/16",
		"127.0.0.0/8", // 本地回环
	}

	for _, cidr := range private4 {
		_, network, err := net.ParseCIDR(cidr)
		if err != nil {
			continue
		}
		if network.Contains(ip) {
			return true
		}
	}

	// IPv6 私有地址
	if ip.To4() == nil { // IPv6
		// 检查本地回环
		if ip.IsLoopback() {
			return true
		}
		// 检查链路本地地址
		if ip.IsLinkLocalUnicast() {
			return true
		}
		// 检查唯一本地地址 (fc00::/7)
		if len(ip) == 16 && (ip[0]&0xfe) == 0xfc {
			return true
		}
	}

	return false
}

// parseForwardedHeader 解析 Forwarded 或 X-Forwarded 头部
func parseForwardedHeader(forwarded string) string {
	// 简单解析 for= 参数
	parts := strings.Split(forwarded, ";")
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if strings.HasPrefix(part, "for=") {
			forValue := strings.TrimPrefix(part, "for=")
			// 移除可能的引号
			forValue = strings.Trim(forValue, "\"")
			// 可能包含端口号，需要分离
			if strings.Contains(forValue, ":") {
				host, _, err := net.SplitHostPort(forValue)
				if err == nil {
					forValue = host
				}
			}
			if ip := parseIP(forValue); ip != "" {
				return ip
			}
		}
	}
	return ""
}

// GetClientIPInfo 获取客户端IP信息，包括是否通过代理
func GetClientIPInfo(c *gin.Context) map[string]interface{} {
	realIP := GetRealClientIP(c)
	ginIP := c.ClientIP()
	
	info := map[string]interface{}{
		"real_ip":     realIP,
		"gin_ip":      ginIP,
		"is_proxied":  realIP != ginIP,
		"headers": map[string]string{
			"X-Real-IP":         c.GetHeader("X-Real-IP"),
			"X-Forwarded-For":   c.GetHeader("X-Forwarded-For"),
			"CF-Connecting-IP":  c.GetHeader("CF-Connecting-IP"),
			"X-Forwarded":       c.GetHeader("X-Forwarded"),
			"Forwarded":         c.GetHeader("Forwarded"),
		},
	}

	return info
}
