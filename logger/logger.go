// logger/logger.go
package logger

import (
	"context"
	"fmt"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

var appContext context.Context

// Initialize stores the application context for logging
func Initialize(ctx context.Context) {
	appContext = ctx
}

// Info logs an informational message
func Info(message string) {
	if appContext != nil {
		runtime.LogInfo(appContext, message)
	}
}

// Error logs an error message
func Error(message string) {
	if appContext != nil {
		runtime.LogError(appContext, message)
	}
}

// Printf logs a formatted message
func Infof(format string, args ...interface{}) {
	if appContext != nil {
		runtime.LogInfo(appContext, fmt.Sprintf(format, args...))
	}
}

// Printf logs a formatted message
func Errorf(format string, args ...interface{}) {
	if appContext != nil {
		runtime.LogError(appContext, fmt.Sprintf(format, args...))
	}
}

// Printf logs a formatted message
func Debugf(format string, args ...interface{}) {
	if appContext != nil {
		runtime.LogDebug(appContext, fmt.Sprintf(format, args...))
	}
}
