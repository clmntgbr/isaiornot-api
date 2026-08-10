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
	webhooks.Post(
		"/minio/object-created",
		container.MediaUploadWebhookMiddleware.Protected(),
		container.MediaUploadWebhookHandler.ObjectCreated,
	)
	webhooks.Post(
		"/stripe",
		container.BillingWebhookMiddleware.Protected(),
		container.BillingWebhookHandler.Execute,
	)
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
	setupMediaRoutes(api, container)
	setupPlanRoutes(api, container)
	setupSubscriptionRoutes(api, container)
	setupInvoiceRoutes(api, container)
	setupRealtimeRoutes(api, container)
}

func setupRealtimeRoutes(api fiber.Router, container *di.Container) {
	auth := container.AuthenticateMiddleware.Protected()
	api.Get("/realtime/connection", auth, container.RealtimeHandler.GetConnection)
}

func setupUserRoutes(api fiber.Router, container *di.Container) {
	users := api.Group("/users")
	users.Use(container.AuthenticateMiddleware.Protected())
	users.Get("/me", container.UserHandler.GetUser)
}

func setupScanRoutes(api fiber.Router, container *di.Container) {
	auth := container.AuthenticateMiddleware.Protected()
	api.Get("/scans", auth, container.ScanHandler.GetScans)
	api.Get("/scans/statistics", auth, container.ScanHandler.GetStatistics)
	api.Post("/scans/presign-upload-url", auth, container.ScanHandler.PresignUpload)
	api.Get("/scans/:id", auth, container.ScanHandler.GetScan)
}

func setupMediaRoutes(api fiber.Router, container *di.Container) {
	auth := container.AuthenticateMiddleware.Protected()
	api.Get("/medias/:id/thumbnail", auth, container.MediaHandler.GetThumbnail)
}

func setupPlanRoutes(api fiber.Router, container *di.Container) {
	api.Get("/plans", container.PlanHandler.GetPlans)
}

func setupSubscriptionRoutes(api fiber.Router, container *di.Container) {
	auth := container.AuthenticateMiddleware.Protected()
	api.Get("/subscription", auth, container.SubscriptionHandler.GetSubscription)
	api.Get("/quota", auth, container.SubscriptionHandler.GetQuota)
	api.Post("/subscriptions", auth, container.SubscriptionHandler.CreateSubscription)
	api.Post("/subscriptions/preview", auth, container.SubscriptionHandler.PreviewSubscription)
}

func setupInvoiceRoutes(api fiber.Router, container *di.Container) {
	auth := container.AuthenticateMiddleware.Protected()
	api.Get("/invoices", auth, container.InvoiceHandler.GetInvoices)
}
