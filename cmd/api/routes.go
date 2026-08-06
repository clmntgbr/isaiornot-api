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
	webhooks.Post("/stripe", container.BillingWebhookMiddleware.Protected(), container.BillingWebhookHandler.Execute)
	webhooks.Post(
		"/minio/object-created",
		container.MediaUploadWebhookMiddleware.Protected(),
		container.MediaUploadWebhookHandler.ObjectCreated,
	)
}

func setupHealthChecks(app *fiber.App) {
	app.Get(healthcheck.LivenessEndpoint, healthcheck.New())
	app.Get(healthcheck.ReadinessEndpoint, healthcheck.New())
	app.Get(healthcheck.StartupEndpoint, healthcheck.New())
}

func setupAPIRoutes(app *fiber.App, container *di.Container) {
	api := app.Group("/api")

	setupPlanRoutes(api, container)

	api.Use(container.AuthenticateMiddleware.Protected())
	setupUsersRoutes(api, container)
	setupSubscriptionRoutes(api, container)
	setupInvoiceRoutes(api, container)
	setupScanRoutes(api, container)
	setupMediaRoutes(api, container)
	setupRealtimeRoutes(api, container)
}

func setupSubscriptionRoutes(api fiber.Router, container *di.Container) {
	api.Get("/subscription", container.SubscriptionHandler.GetSubscription)
	api.Get("/quota", container.SubscriptionHandler.GetQuota)
	api.Post("/subscriptions", container.SubscriptionHandler.CreateSubscription)
	api.Get("/subscriptions/portal", container.SubscriptionHandler.CreateBillingPortal)
}

func setupInvoiceRoutes(api fiber.Router, container *di.Container) {
	api.Get("/invoices", container.InvoiceHandler.GetInvoices)
}

func setupRealtimeRoutes(api fiber.Router, container *di.Container) {
	api.Get("/realtime/connection", container.RealtimeHandler.GetConnection)
}

func setupUsersRoutes(api fiber.Router, container *di.Container) {
	api.Get("/users/me", container.UserHandler.GetUser)
}

func setupScanRoutes(api fiber.Router, container *di.Container) {
	api.Post("/scans/presign-upload-url", container.ScanHandler.GeneratePresignedUploadUrl)
	api.Get("/scans/statistics", container.ScanHandler.GetStatistics)
	api.Get("/scans", container.ScanHandler.GetScans)
	api.Get("/scans/:id", container.ScanHandler.GetScan)
}

func setupMediaRoutes(api fiber.Router, container *di.Container) {
	api.Get("/medias/:id/thumbnail", container.MediaHandler.GetThumbnail)
}

func setupPlanRoutes(api fiber.Router, container *di.Container) {
	api.Get("/plans", container.PlanHandler.GetPlans)
}
