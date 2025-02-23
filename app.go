package main

import (
	"context"
	"fmt"
	"os"

	server "eth-daq-software/internal"
)

// App struct
type App struct {
	ctx context.Context
}

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{}
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
		go server.StartListener(port)
	}

}

// Greet returns a greeting for the given name
func (a *App) Greet(name string) string {
	return fmt.Sprintf("Hello %s, It's show time!", name)
}
