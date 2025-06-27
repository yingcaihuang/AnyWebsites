package main

import (
	"fmt"
	"log"
	"time"

	"anywebsites/internal/services"
)

func main() {
	// 初始化 GeoIP 服务
	geoipService, err := services.NewGeoIPService("data/geoip/GeoLite2-City.mmdb")
	if err != nil {
		log.Fatal("Failed to initialize GeoIP service:", err)
	}
	defer geoipService.Close()

	// 测试IP地址
	testIP := "8.8.8.8"

	fmt.Println("GeoIP 缓存测试")
	fmt.Println("==============")

	// 第一次查询（应该从数据库查询）
	fmt.Printf("第一次查询 %s:\n", testIP)
	start := time.Now()
	locationInfo1, err := geoipService.GetLocationInfo(testIP)
	duration1 := time.Since(start)
	if err != nil {
		log.Printf("查询失败: %v", err)
	} else {
		fmt.Printf("结果: %s, %s (%.4f, %.4f)\n", 
			locationInfo1.Country, locationInfo1.City, 
			locationInfo1.Latitude, locationInfo1.Longitude)
	}
	fmt.Printf("查询时间: %v\n", duration1)

	// 显示缓存统计
	stats := geoipService.GetCacheStats()
	fmt.Printf("缓存统计: %+v\n", stats)

	fmt.Println()

	// 第二次查询（应该从缓存获取）
	fmt.Printf("第二次查询 %s (应该从缓存获取):\n", testIP)
	start = time.Now()
	locationInfo2, err := geoipService.GetLocationInfo(testIP)
	duration2 := time.Since(start)
	if err != nil {
		log.Printf("查询失败: %v", err)
	} else {
		fmt.Printf("结果: %s, %s (%.4f, %.4f)\n", 
			locationInfo2.Country, locationInfo2.City, 
			locationInfo2.Latitude, locationInfo2.Longitude)
	}
	fmt.Printf("查询时间: %v\n", duration2)

	// 显示缓存统计
	stats = geoipService.GetCacheStats()
	fmt.Printf("缓存统计: %+v\n", stats)

	// 比较性能
	fmt.Printf("\n性能对比:\n")
	fmt.Printf("第一次查询: %v\n", duration1)
	fmt.Printf("第二次查询: %v\n", duration2)
	if duration1 > duration2 {
		speedup := float64(duration1) / float64(duration2)
		fmt.Printf("缓存加速: %.2fx\n", speedup)
	}

	// 测试多个不同IP
	fmt.Println("\n测试多个IP地址:")
	testIPs := []string{"1.1.1.1", "114.114.114.114", "208.67.222.222"}
	
	for _, ip := range testIPs {
		start := time.Now()
		locationInfo, err := geoipService.GetLocationInfo(ip)
		duration := time.Since(start)
		if err != nil {
			fmt.Printf("IP: %-15s | 错误: %v\n", ip, err)
		} else {
			fmt.Printf("IP: %-15s | %s, %s | 时间: %v\n", 
				ip, locationInfo.Country, locationInfo.City, duration)
		}
	}

	// 最终缓存统计
	fmt.Println("\n最终缓存统计:")
	stats = geoipService.GetCacheStats()
	fmt.Printf("%+v\n", stats)
}
