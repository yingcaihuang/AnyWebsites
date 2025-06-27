package services

import (
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/oschwald/geoip2-golang"
)

// CacheEntry 缓存条目
type CacheEntry struct {
	LocationInfo *LocationInfo
	Timestamp    time.Time
}

// BatchRequest 批量查询请求
type BatchRequest struct {
	IP       string
	Response chan *BatchResponse
}

// BatchResponse 批量查询响应
type BatchResponse struct {
	LocationInfo *LocationInfo
	Error        error
}

type GeoIPService struct {
	db    *geoip2.Reader
	cache map[string]*CacheEntry
	mutex sync.RWMutex
	// 缓存过期时间，默认1小时
	cacheExpiry time.Duration
	// 批量处理相关
	batchChannel chan *BatchRequest
	batchSize    int
	batchTimeout time.Duration
	// 监控和统计
	stats *ServiceStats
}

// ServiceStats 服务统计信息
type ServiceStats struct {
	mutex           sync.RWMutex
	TotalRequests   int64
	CacheHits       int64
	CacheMisses     int64
	BatchProcessed  int64
	DirectProcessed int64
	Errors          int64
	LastError       string
	LastErrorTime   time.Time
}

type LocationInfo struct {
	Country   string  `json:"country"`
	City      string  `json:"city"`
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

func NewGeoIPService(dbPath string) (*GeoIPService, error) {
	db, err := geoip2.Open(dbPath)
	if err != nil {
		return nil, err
	}

	service := &GeoIPService{
		db:           db,
		cache:        make(map[string]*CacheEntry),
		cacheExpiry:  time.Hour,                      // 默认缓存1小时
		batchChannel: make(chan *BatchRequest, 1000), // 批量处理通道
		batchSize:    10,                             // 批量大小
		batchTimeout: 50 * time.Millisecond,          // 批量超时时间
		stats:        &ServiceStats{},                // 初始化统计信息
	}

	// 启动定期清理过期缓存的 goroutine
	go service.startCacheCleanup()

	// 启动批量处理 goroutine
	go service.startBatchProcessor()

	return service, nil
}

func (g *GeoIPService) Close() error {
	return g.db.Close()
}

func (g *GeoIPService) GetLocationInfo(ipStr string) (*LocationInfo, error) {
	// 处理本地IP地址
	if g.isLocalIP(ipStr) {
		return &LocationInfo{
			Country:   "Local",
			City:      "Local",
			Latitude:  0,
			Longitude: 0,
		}, nil
	}

	// 检查缓存
	if cachedInfo := g.getFromCache(ipStr); cachedInfo != nil {
		return cachedInfo, nil
	}

	// 使用批量处理
	return g.getBatchLocationInfo(ipStr)
}

// getBatchLocationInfo 通过批量处理获取地理位置信息
func (g *GeoIPService) getBatchLocationInfo(ipStr string) (*LocationInfo, error) {
	g.incrementTotalRequests()

	responseChan := make(chan *BatchResponse, 1)

	request := &BatchRequest{
		IP:       ipStr,
		Response: responseChan,
	}

	// 发送到批量处理通道
	select {
	case g.batchChannel <- request:
		// 等待响应
		response := <-responseChan
		if response.Error != nil {
			g.recordError(response.Error.Error())
		}
		return response.LocationInfo, response.Error
	default:
		// 通道满了，直接处理
		g.incrementDirectProcessed()
		result, err := g.processLocationInfo(ipStr)
		if err != nil {
			g.recordError(err.Error())
		}
		return result, err
	}
}

// processLocationInfo 直接处理单个IP地址的地理位置查询
func (g *GeoIPService) processLocationInfo(ipStr string) (*LocationInfo, error) {
	ip := net.ParseIP(ipStr)
	if ip == nil {
		return nil, fmt.Errorf("invalid IP address: %s", ipStr)
	}

	record, err := g.db.City(ip)
	if err != nil {
		return nil, err
	}

	country := "Unknown"
	if len(record.Country.Names) > 0 {
		if name, ok := record.Country.Names["en"]; ok {
			country = name
		} else {
			// 取第一个可用的名称
			for _, name := range record.Country.Names {
				country = name
				break
			}
		}
	}

	city := "Unknown"
	if len(record.City.Names) > 0 {
		if name, ok := record.City.Names["en"]; ok {
			city = name
		} else {
			// 取第一个可用的名称
			for _, name := range record.City.Names {
				city = name
				break
			}
		}
	}

	locationInfo := &LocationInfo{
		Country:   country,
		City:      city,
		Latitude:  float64(record.Location.Latitude),
		Longitude: float64(record.Location.Longitude),
	}

	// 将结果存入缓存
	g.setCache(ipStr, locationInfo)

	return locationInfo, nil
}

func (g *GeoIPService) isLocalIP(ip string) bool {
	// 检查常见的本地IP地址
	localIPs := []string{
		"127.0.0.1",
		"::1",
		"localhost",
	}

	for _, localIP := range localIPs {
		if ip == localIP {
			return true
		}
	}

	// 检查私有IP地址范围
	parsedIP := net.ParseIP(ip)
	if parsedIP == nil {
		return false
	}

	// IPv4 私有地址范围
	private4 := []string{
		"10.0.0.0/8",
		"172.16.0.0/12",
		"192.168.0.0/16",
	}

	for _, cidr := range private4 {
		_, network, err := net.ParseCIDR(cidr)
		if err != nil {
			continue
		}
		if network.Contains(parsedIP) {
			return true
		}
	}

	return false
}

// getFromCache 从缓存中获取地理位置信息
func (g *GeoIPService) getFromCache(ip string) *LocationInfo {
	g.mutex.RLock()
	defer g.mutex.RUnlock()

	entry, exists := g.cache[ip]
	if !exists {
		g.incrementCacheMisses()
		return nil
	}

	// 检查缓存是否过期
	if time.Since(entry.Timestamp) > g.cacheExpiry {
		// 缓存过期，删除条目
		delete(g.cache, ip)
		g.incrementCacheMisses()
		return nil
	}

	g.incrementCacheHits()
	return entry.LocationInfo
}

// setCache 将地理位置信息存入缓存
func (g *GeoIPService) setCache(ip string, locationInfo *LocationInfo) {
	g.mutex.Lock()
	defer g.mutex.Unlock()

	g.cache[ip] = &CacheEntry{
		LocationInfo: locationInfo,
		Timestamp:    time.Now(),
	}
}

// ClearExpiredCache 清理过期的缓存条目
func (g *GeoIPService) ClearExpiredCache() {
	g.mutex.Lock()
	defer g.mutex.Unlock()

	now := time.Now()
	for ip, entry := range g.cache {
		if now.Sub(entry.Timestamp) > g.cacheExpiry {
			delete(g.cache, ip)
		}
	}
}

// GetCacheStats 获取缓存统计信息
func (g *GeoIPService) GetCacheStats() map[string]interface{} {
	g.mutex.RLock()
	defer g.mutex.RUnlock()

	return map[string]interface{}{
		"cache_size":    len(g.cache),
		"cache_expiry":  g.cacheExpiry.String(),
		"cache_entries": len(g.cache),
	}
}

// startCacheCleanup 启动定期清理过期缓存的后台任务
func (g *GeoIPService) startCacheCleanup() {
	ticker := time.NewTicker(30 * time.Minute) // 每30分钟清理一次
	defer ticker.Stop()

	for range ticker.C {
		g.ClearExpiredCache()
	}
}

// startBatchProcessor 启动批量处理器
func (g *GeoIPService) startBatchProcessor() {
	batch := make([]*BatchRequest, 0, g.batchSize)
	timer := time.NewTimer(g.batchTimeout)
	timer.Stop() // 初始状态停止计时器

	for {
		select {
		case request := <-g.batchChannel:
			batch = append(batch, request)

			// 如果这是第一个请求，启动计时器
			if len(batch) == 1 {
				timer.Reset(g.batchTimeout)
			}

			// 如果达到批量大小，立即处理
			if len(batch) >= g.batchSize {
				g.processBatch(batch)
				batch = batch[:0] // 清空批次
				timer.Stop()
			}

		case <-timer.C:
			// 超时，处理当前批次
			if len(batch) > 0 {
				g.processBatch(batch)
				batch = batch[:0] // 清空批次
			}
		}
	}
}

// processBatch 处理一批地理位置查询请求
func (g *GeoIPService) processBatch(batch []*BatchRequest) {
	g.incrementBatchProcessed()
	for _, request := range batch {
		go func(req *BatchRequest) {
			locationInfo, err := g.processLocationInfo(req.IP)
			if err != nil {
				g.recordError(err.Error())
			}
			req.Response <- &BatchResponse{
				LocationInfo: locationInfo,
				Error:        err,
			}
		}(request)
	}
}

// 统计方法
func (g *GeoIPService) incrementTotalRequests() {
	g.stats.mutex.Lock()
	defer g.stats.mutex.Unlock()
	g.stats.TotalRequests++
}

func (g *GeoIPService) incrementCacheHits() {
	g.stats.mutex.Lock()
	defer g.stats.mutex.Unlock()
	g.stats.CacheHits++
}

func (g *GeoIPService) incrementCacheMisses() {
	g.stats.mutex.Lock()
	defer g.stats.mutex.Unlock()
	g.stats.CacheMisses++
}

func (g *GeoIPService) incrementBatchProcessed() {
	g.stats.mutex.Lock()
	defer g.stats.mutex.Unlock()
	g.stats.BatchProcessed++
}

func (g *GeoIPService) incrementDirectProcessed() {
	g.stats.mutex.Lock()
	defer g.stats.mutex.Unlock()
	g.stats.DirectProcessed++
}

func (g *GeoIPService) recordError(errorMsg string) {
	g.stats.mutex.Lock()
	defer g.stats.mutex.Unlock()
	g.stats.Errors++
	g.stats.LastError = errorMsg
	g.stats.LastErrorTime = time.Now()
}

// GetServiceStats 获取服务统计信息
func (g *GeoIPService) GetServiceStats() map[string]interface{} {
	g.stats.mutex.RLock()
	defer g.stats.mutex.RUnlock()

	cacheHitRate := float64(0)
	if g.stats.TotalRequests > 0 {
		cacheHitRate = float64(g.stats.CacheHits) / float64(g.stats.TotalRequests) * 100
	}

	return map[string]interface{}{
		"total_requests":   g.stats.TotalRequests,
		"cache_hits":       g.stats.CacheHits,
		"cache_misses":     g.stats.CacheMisses,
		"cache_hit_rate":   fmt.Sprintf("%.2f%%", cacheHitRate),
		"batch_processed":  g.stats.BatchProcessed,
		"direct_processed": g.stats.DirectProcessed,
		"errors":           g.stats.Errors,
		"last_error":       g.stats.LastError,
		"last_error_time":  g.stats.LastErrorTime.Format("2006-01-02 15:04:05"),
	}
}
