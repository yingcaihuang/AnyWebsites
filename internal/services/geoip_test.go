package services

import (
	"testing"
	"time"
)

// TestGeoIPService_GetLocationInfo 测试地理位置信息获取（仅测试本地IP）
func TestGeoIPService_GetLocationInfo(t *testing.T) {
	// 创建测试服务（使用模拟数据）
	service := &GeoIPService{
		cache:        make(map[string]*CacheEntry),
		cacheExpiry:  time.Hour,
		batchChannel: make(chan *BatchRequest, 100),
		batchSize:    10,
		batchTimeout: 50 * time.Millisecond,
		stats:        &ServiceStats{},
	}

	tests := []struct {
		name     string
		ip       string
		wantErr  bool
		expected *LocationInfo
	}{
		{
			name:    "本地IP地址",
			ip:      "127.0.0.1",
			wantErr: false,
			expected: &LocationInfo{
				Country: "Local",
				City:    "Local",
			},
		},
		{
			name:    "本地IPv6地址",
			ip:      "::1",
			wantErr: false,
			expected: &LocationInfo{
				Country: "Local",
				City:    "Local",
			},
		},
		{
			name:    "localhost字符串",
			ip:      "localhost",
			wantErr: false,
			expected: &LocationInfo{
				Country: "Local",
				City:    "Local",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 只测试本地IP，因为它们不需要GeoIP数据库
			if !service.isLocalIP(tt.ip) {
				t.Skip("跳过非本地IP测试，因为需要GeoIP数据库")
				return
			}

			result, err := service.GetLocationInfo(tt.ip)

			if tt.wantErr {
				if err == nil {
					t.Errorf("期望错误，但没有返回错误")
				}
				return
			}

			if err != nil {
				t.Errorf("不期望错误，但返回了错误: %v", err)
				return
			}

			if result == nil {
				t.Errorf("结果不应该为nil")
				return
			}

			if tt.expected != nil {
				if result.Country != tt.expected.Country {
					t.Errorf("Country = %v, want %v", result.Country, tt.expected.Country)
				}
				if result.City != tt.expected.City {
					t.Errorf("City = %v, want %v", result.City, tt.expected.City)
				}
			}
		})
	}
}

// TestGeoIPService_Cache 测试缓存功能
func TestGeoIPService_Cache(t *testing.T) {
	service := &GeoIPService{
		cache:       make(map[string]*CacheEntry),
		cacheExpiry: time.Hour,
		stats:       &ServiceStats{},
	}

	testIP := "127.0.0.1"
	testLocation := &LocationInfo{
		Country: "Test Country",
		City:    "Test City",
	}

	// 测试缓存存储
	service.setCache(testIP, testLocation)

	// 测试缓存检索
	cached := service.getFromCache(testIP)
	if cached == nil {
		t.Errorf("缓存检索失败，应该返回缓存的数据")
		return
	}

	if cached.Country != testLocation.Country {
		t.Errorf("缓存的Country = %v, want %v", cached.Country, testLocation.Country)
	}

	// 测试缓存过期
	service.cacheExpiry = time.Nanosecond
	time.Sleep(time.Millisecond) // 确保缓存过期

	expired := service.getFromCache(testIP)
	if expired != nil {
		t.Errorf("过期的缓存应该返回nil")
	}
}

// TestGeoIPService_Statistics 测试统计功能
func TestGeoIPService_Statistics(t *testing.T) {
	service := &GeoIPService{
		stats: &ServiceStats{},
	}

	// 测试统计计数
	service.incrementTotalRequests()
	service.incrementCacheHits()
	service.incrementCacheMisses()
	service.incrementBatchProcessed()
	service.incrementDirectProcessed()

	stats := service.GetServiceStats()

	if stats["total_requests"] != int64(1) {
		t.Errorf("total_requests = %v, want 1", stats["total_requests"])
	}

	if stats["cache_hits"] != int64(1) {
		t.Errorf("cache_hits = %v, want 1", stats["cache_hits"])
	}

	if stats["cache_misses"] != int64(1) {
		t.Errorf("cache_misses = %v, want 1", stats["cache_misses"])
	}

	// 测试错误记录
	testError := "test error message"
	service.recordError(testError)

	updatedStats := service.GetServiceStats()
	if updatedStats["errors"] != int64(1) {
		t.Errorf("errors = %v, want 1", updatedStats["errors"])
	}

	if updatedStats["last_error"] != testError {
		t.Errorf("last_error = %v, want %v", updatedStats["last_error"], testError)
	}
}

// TestGeoIPService_CacheStats 测试缓存统计
func TestGeoIPService_CacheStats(t *testing.T) {
	service := &GeoIPService{
		cache:       make(map[string]*CacheEntry),
		cacheExpiry: time.Hour,
	}

	// 添加一些缓存条目
	service.setCache("127.0.0.1", &LocationInfo{Country: "Test1"})
	service.setCache("192.168.1.1", &LocationInfo{Country: "Test2"})

	stats := service.GetCacheStats()

	if stats["cache_size"] != 2 {
		t.Errorf("cache_size = %v, want 2", stats["cache_size"])
	}

	if stats["cache_expiry"] != "1h0m0s" {
		t.Errorf("cache_expiry = %v, want 1h0m0s", stats["cache_expiry"])
	}
}

// TestIsLocalIP 测试本地IP检测
func TestIsLocalIP(t *testing.T) {
	service := &GeoIPService{}

	tests := []struct {
		name     string
		ip       string
		expected bool
	}{
		{"IPv4 localhost", "127.0.0.1", true},
		{"IPv6 localhost", "::1", true},
		{"localhost string", "localhost", true},
		{"IPv4 public", "8.8.8.8", false},
		{"IPv6 public", "2001:4860:4860::8888", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := service.isLocalIP(tt.ip)
			if result != tt.expected {
				t.Errorf("isLocalIP(%s) = %v, want %v", tt.ip, result, tt.expected)
			}
		})
	}
}

// TestBatchRequest 测试批量请求结构
func TestBatchRequest(t *testing.T) {
	responseChan := make(chan *BatchResponse, 1)

	request := &BatchRequest{
		IP:       "127.0.0.1",
		Response: responseChan,
	}

	if request.IP != "127.0.0.1" {
		t.Errorf("BatchRequest.IP = %v, want 127.0.0.1", request.IP)
	}

	if request.Response == nil {
		t.Errorf("BatchRequest.Response should not be nil")
	}

	// 测试响应通道
	testResponse := &BatchResponse{
		LocationInfo: &LocationInfo{Country: "Test"},
		Error:        nil,
	}

	go func() {
		request.Response <- testResponse
	}()

	received := <-request.Response
	if received.LocationInfo.Country != "Test" {
		t.Errorf("Received response country = %v, want Test", received.LocationInfo.Country)
	}
}

// BenchmarkGeoIPService_GetLocationInfo 性能基准测试（本地IP）
func BenchmarkGeoIPService_GetLocationInfo(b *testing.B) {
	service := &GeoIPService{
		cache:       make(map[string]*CacheEntry),
		cacheExpiry: time.Hour,
		stats:       &ServiceStats{},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = service.GetLocationInfo("127.0.0.1")
	}
}

// BenchmarkGeoIPService_Cache 缓存性能基准测试
func BenchmarkGeoIPService_Cache(b *testing.B) {
	service := &GeoIPService{
		cache:       make(map[string]*CacheEntry),
		cacheExpiry: time.Hour,
		stats:       &ServiceStats{},
	}

	// 预填充缓存
	testLocation := &LocationInfo{Country: "Test", City: "Test"}
	service.setCache("127.0.0.1", testLocation)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = service.getFromCache("127.0.0.1")
	}
}
