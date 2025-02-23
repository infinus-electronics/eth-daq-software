package main

import (
	"context"
	"fmt"
	"os"

	server "eth-daq-software/internal"
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

	if err := os.MkdirAll("data", 0755); err != nil {
		fmt.Printf("Failed to create data directory: %v\n", err)
		return
	}

	ports := []int{5555, 5556, 5557}

	for _, port := range ports {
		go a.server.StartListener(port)
	}

}

// Greet returns a greeting for the given name
func (a *App) Greet(name string) string {
	return fmt.Sprintf("Hello %s, It's show time!", name)
}

// GetPortRate returns the current transfer rate for a specific port
func (a *App) GetPortRate(key server.BufferKey) float64 {
	rate, exists := a.server.GetBufferRate(key.IP, key.Port)
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

// Add this method to expose the type
func (a *App) DUMMYGetIPConnectionDetails(conn server.IPConnection) string {
	// Just a dummy method to expose the type
	return fmt.Sprintf("Connection details: %+v", conn)
}
