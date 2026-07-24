package main

import (
	"go-api/cmd/api/wire"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/healthcheck"
)

func setupRoutes(app *fiber.App, container *wire.Container) {
	setupHealthChecks(app)
	setupWebhooks(app, container)
	setupAPIRoutes(app, container)
}

func setupWebhooks(app *fiber.App, container *wire.Container) {
	webhooks := app.Group("/webhook")

	webhooks.Post("/clerk", container.ClerkMiddleware.Protected(), container.ClerkHandler.Execute)
	webhooks.Post("/stripe", container.StripeMiddleware.Protected(), container.StripeHandler.Execute)

	minioWebhooks := app.Group("/webhooks/minio")
	minioWebhooks.Post(
		"/object-created",
		container.MinIOMiddleware.Protected(),
		container.MinIOHandler.ObjectCreated,
	)
}

func setupHealthChecks(app *fiber.App) {
	app.Get(healthcheck.LivenessEndpoint, healthcheck.New())
	app.Get(healthcheck.ReadinessEndpoint, healthcheck.New())
	app.Get(healthcheck.StartupEndpoint, healthcheck.New())
}

func setupAPIRoutes(app *fiber.App, container *wire.Container) {
	api := app.Group("/api")

	// Public routes
	setupPlanRoutes(api, container)

	// Protected routes
	api.Use(container.AuthenticateMiddleware.Protected())
	setupUsersRoutes(api, container)
	setupSubscriptionRoutes(api, container)
	setupAnalysisRoutes(api, container)
	setupMediaRoutes(api, container)
	setupRealtimeRoutes(api, container)
}

func setupSubscriptionRoutes(api fiber.Router, container *wire.Container) {
	api.Post("/subscriptions", container.SubscriptionHandler.CreateSubscription)
}

func setupRealtimeRoutes(api fiber.Router, container *wire.Container) {
	api.Get("/realtime/connection", container.RealtimeHandler.GetConnection)
}

func setupUsersRoutes(api fiber.Router, container *wire.Container) {
	api.Get("/users/me", container.UserHandler.GetUser)
}

func setupAnalysisRoutes(api fiber.Router, container *wire.Container) {
	api.Post("/analyses/presign-upload-url", container.AnalysisHandler.GeneratePresignedUploadUrl)
	api.Get("/analyses/statistics", container.AnalysisHandler.GetStatistics)
	api.Get("/analyses", container.AnalysisHandler.GetAnalyses)
	api.Get("/analyses/:id", container.AnalysisHandler.GetAnalysis)
}

func setupMediaRoutes(api fiber.Router, container *wire.Container) {
	api.Get("/medias/:id/thumbnail", container.MediaHandler.GetThumbnail)
}

func setupPlanRoutes(api fiber.Router, container *wire.Container) {
	api.Get("/plans", container.PlanHandler.GetPlans)
}
