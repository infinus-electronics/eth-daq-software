package main

import (
	"context"
	"eth-daq-software/logger"
	"eth-daq-software/server"
	"fmt"
	"os"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// App struct
type App struct {
	ctx    context.Context
	server *server.Server
}

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{
		server: server.NewServer(),
	}
}

// startup is called when the app starts. The context is saved
// so we can call the runtime methods
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	logger.Initialize(ctx)

	if err := os.MkdirAll("data", 0755); err != nil {
		runtime.LogErrorf(ctx, "Failed to create data directory: %v\n", err)
		return
	}

	ports := []int{5002, 5555, 5556, 5557}

	for _, port := range ports {
		go a.server.StartListener(port)
	}

}
func (a *App) shutdown(ctx context.Context) {
	a.server.Shutdown()
}

// Greet returns a greeting for the given name
func (a *App) Greet(name string) string {
	return fmt.Sprintf("Hello %s, It's show time!", name)
}

// GetPortRate returns the current transfer rate for a specific port
func (a *App) GetPortRate(key server.BufferKey) float64 {
	rate, exists := a.server.GetBufferRate(key)
	if !exists {
		return 0
	}
	return rate
}

// GetAllRates returns all current transfer rates
func (a *App) GetAllRates() map[string]float64 {
	rates := a.server.GetAllBufferRates()
	if rates == nil {
		return make(map[string]float64)
	}
	return rates
}

func (a *App) GetAllConnectedIPs() map[string]server.IPConnection {
	ips := a.server.GetAllConnectedIPs()
	if ips == nil {
		return make(map[string]server.IPConnection)
	}
	return ips
}

func (a *App) GetLogs(ip string) []string {
	logs := a.server.GetLastLogs(ip)
	return logs
}

func (a *App) GetPortAverage(key server.BufferKey) float64 {
	// fmt.Printf("Request: %s, %d\n", key.IP, key.Port)
	result, _ := a.server.GetPortAverage(key)
	// fmt.Printf("Result: %f", result)
	return result
}

func (a *App) GetPortAverageB(key server.BufferKey) float64 {
	result, _ := a.server.GetPortAverageB(key)
	return result
}

// Add this method to expose the type
func (a *App) DUMMYGetIPConnectionDetails(conn server.IPConnection) string {
	// Just a dummy method to expose the type
	return fmt.Sprintf("Connection details: %+v", conn)
}
