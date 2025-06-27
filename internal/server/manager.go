package server

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// Manager 服务器管理器，支持优雅重启
type Manager struct {
	server     *http.Server
	router     *gin.Engine
	mutex      sync.RWMutex
	isRunning  bool
	restartCh  chan RestartRequest
	shutdownCh chan struct{}
}

// RestartRequest 重启请求
type RestartRequest struct {
	NewAddr string
	Done    chan error
}

// NewManager 创建服务器管理器
func NewManager(router *gin.Engine) *Manager {
	return &Manager{
		router:     router,
		restartCh:  make(chan RestartRequest, 1),
		shutdownCh: make(chan struct{}),
	}
}

// Start 启动服务器
func (m *Manager) Start(addr string) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if m.isRunning {
		return fmt.Errorf("server is already running")
	}

	m.server = &http.Server{
		Addr:    addr,
		Handler: m.router,
	}

	m.isRunning = true

	// 启动服务器监听
	go func() {
		log.Printf("Server starting on %s", addr)
		if err := m.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Printf("Server error: %v", err)
		}
	}()

	// 启动重启监听器
	go m.restartListener()

	return nil
}

// Stop 停止服务器
func (m *Manager) Stop() error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if !m.isRunning {
		return nil
	}

	// 发送关闭信号
	close(m.shutdownCh)

	// 优雅关闭服务器
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := m.server.Shutdown(ctx); err != nil {
		log.Printf("Server forced to shutdown: %v", err)
		return err
	}

	m.isRunning = false
	log.Println("Server stopped gracefully")
	return nil
}

// Restart 重启服务器到新地址
func (m *Manager) Restart(newAddr string) error {
	done := make(chan error, 1)
	
	select {
	case m.restartCh <- RestartRequest{NewAddr: newAddr, Done: done}:
		return <-done
	default:
		return fmt.Errorf("restart already in progress")
	}
}

// IsRunning 检查服务器是否正在运行
func (m *Manager) IsRunning() bool {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	return m.isRunning
}

// GetCurrentAddr 获取当前监听地址
func (m *Manager) GetCurrentAddr() string {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	
	if m.server != nil {
		return m.server.Addr
	}
	return ""
}

// restartListener 重启监听器
func (m *Manager) restartListener() {
	for {
		select {
		case <-m.shutdownCh:
			return
		case req := <-m.restartCh:
			err := m.performRestart(req.NewAddr)
			req.Done <- err
		}
	}
}

// performRestart 执行重启
func (m *Manager) performRestart(newAddr string) error {
	log.Printf("Performing graceful restart from %s to %s", m.GetCurrentAddr(), newAddr)

	// 优雅关闭当前服务器
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := m.server.Shutdown(ctx); err != nil {
		log.Printf("Error during graceful shutdown: %v", err)
		return fmt.Errorf("failed to shutdown server: %w", err)
	}

	// 创建新服务器
	m.mutex.Lock()
	m.server = &http.Server{
		Addr:    newAddr,
		Handler: m.router,
	}
	m.mutex.Unlock()

	// 启动新服务器
	go func() {
		log.Printf("Server restarting on %s", newAddr)
		if err := m.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Printf("Server error after restart: %v", err)
		}
	}()

	log.Printf("Server successfully restarted on %s", newAddr)
	return nil
}

// WaitForShutdown 等待关闭信号
func (m *Manager) WaitForShutdown() {
	<-m.shutdownCh
}
