package main

import (
	"go-api/cmd/api/di"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/healthcheck"
)

func setupRoutes(app *fiber.App, container *di.Container) {
	setupHealthChecks(app)
	setupWebhooks(app, container)
	setupAPIRoutes(app, container)
}

func setupWebhooks(app *fiber.App, container *di.Container) {
	webhooks := app.Group("/webhooks")
	webhooks.Post("/clerk", container.UserWebhookMiddleware.Protected(), container.UserWebhookHandler.Execute)
}

func setupHealthChecks(app *fiber.App) {
	app.Get(healthcheck.LivenessEndpoint, healthcheck.New())
	app.Get(healthcheck.ReadinessEndpoint, healthcheck.New())
	app.Get(healthcheck.StartupEndpoint, healthcheck.New())
}

func setupAPIRoutes(app *fiber.App, container *di.Container) {
	api := app.Group("/api")

	setupUserRoutes(api, container)
	setupScanRoutes(api, container)
}

func setupUserRoutes(api fiber.Router, container *di.Container) {
	users := api.Group("/users")
	users.Use(container.AuthenticateMiddleware.Protected())
	users.Get("/me", container.UserHandler.GetUser)
}

func setupScanRoutes(api fiber.Router, container *di.Container) {
	auth := container.AuthenticateMiddleware.Protected()
	api.Get("/scans", auth, container.ScanHandler.GetScans)
	api.Get("/scan/:id", auth, container.ScanHandler.GetScan)
}
