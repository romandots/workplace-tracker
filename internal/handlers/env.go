package handlers

import "github.com/example/workplace-tracker/internal/app"

// Env bundles dependencies for HTTP handlers.
type Env struct {
	App *app.App
}

func NewEnv(a *app.App) *Env {
	return &Env{App: a}
}
