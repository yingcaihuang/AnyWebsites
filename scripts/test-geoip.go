package main

import (
	"fmt"
	"log"

	"anywebsites/internal/services"
)

func main() {
	// 初始化 GeoIP 服务
	geoipService, err := services.NewGeoIPService("data/geoip/GeoLite2-City.mmdb")
	if err != nil {
		log.Fatal("Failed to initialize GeoIP service:", err)
	}
	defer geoipService.Close()

	// 测试不同的 IP 地址
	testIPs := []string{
		"127.0.0.1",     // 本地 IPv4
		"::1",           // 本地 IPv6
		"8.8.8.8",       // Google DNS (美国)
		"1.1.1.1",       // Cloudflare DNS (美国)
		"114.114.114.114", // 中国 DNS
		"208.67.222.222", // OpenDNS (美国)
		"185.228.168.9",  // 欧洲某个IP
		"203.208.60.1",   // 亚洲某个IP
	}

	fmt.Println("GeoIP 测试结果:")
	fmt.Println("================")

	for _, ip := range testIPs {
		locationInfo, err := geoipService.GetLocationInfo(ip)
		if err != nil {
			fmt.Printf("IP: %-15s | 错误: %v\n", ip, err)
			continue
		}

		fmt.Printf("IP: %-15s | 国家: %-15s | 城市: %-15s | 坐标: (%.4f, %.4f)\n",
			ip, locationInfo.Country, locationInfo.City, 
			locationInfo.Latitude, locationInfo.Longitude)
	}
}
