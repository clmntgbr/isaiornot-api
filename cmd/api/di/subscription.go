package di

import (
	httphandler "go-api/handler/http"
	infraStripe "go-api/infrastructure/stripe"
	"go-api/usecase/plan"
	"go-api/usecase/subscription"
)

type subscriptionBundle struct {
	planHandler           *httphandler.PlanHandler
	billingWebhookHandler *httphandler.BillingWebhookHandler
	subscriptionHandler   *httphandler.SubscriptionHandler
	invoiceHandler        *httphandler.InvoiceHandler
}

func subscriptionFree(d *apiDeps) *subscription.CreateFreeSubscriptionUseCase {
	return subscription.NewCreateFreeSubscriptionUseCase(d.planRepo, d.subscriptionRepo, d.userRepo)
}

func wireSubscription(d *apiDeps, auth authBundle) subscriptionBundle {
	resolveEffectivePlanUseCase := subscription.NewResolveEffectivePlanUseCase(d.planRepo)
	getQuotaUsageUseCase := subscription.NewGetQuotaUsageUseCase(d.mediaRepo)
	checkoutSessionGateway := infraStripe.NewCheckoutSessionGateway(d.env)
	subscriptionGateway := infraStripe.NewSubscriptionGateway(d.env)
	billingPortalGateway := infraStripe.NewBillingPortalGateway(d.env)
	subscriptionNotifier := subscription.NewNotifier(d.userRepo, d.subscriptionRepo, d.centrifugoPublisher)
	upsertInvoiceUseCase := subscription.NewUpsertInvoiceUseCase(
		d.invoiceRepo,
		d.subscriptionRepo,
		d.userRepo,
	)

	createSubscriptionUseCase := subscription.NewCreateSubscriptionUseCase(
		d.planRepo,
		auth.fetchUserUseCase,
		checkoutSessionGateway,
	)
	createBillingPortalUseCase := subscription.NewCreateBillingPortalUseCase(
		d.subscriptionRepo,
		billingPortalGateway,
	)
	getUserSubscriptionUseCase := subscription.NewGetUserSubscriptionUseCase(
		d.subscriptionRepo,
		resolveEffectivePlanUseCase,
	)
	getUserQuotaUsageUseCase := subscription.NewGetUserQuotaUsageUseCase(
		d.subscriptionRepo,
		resolveEffectivePlanUseCase,
		getQuotaUsageUseCase,
	)
	getInvoicesUseCase := subscription.NewGetInvoicesUseCase(d.invoiceRepo)

	return subscriptionBundle{
		planHandler: httphandler.NewPlanHandler(plan.NewGetPlansUseCase(d.planRepo)),
		billingWebhookHandler: httphandler.NewBillingWebhookHandler(
			subscription.NewCheckoutCompletedUseCase(
				d.userRepo,
				d.planRepo,
				d.subscriptionRepo,
				subscriptionGateway,
				subscriptionNotifier,
			),
			subscription.NewSubscriptionUpdatedUseCase(d.planRepo, d.subscriptionRepo, subscriptionNotifier),
			subscription.NewSubscriptionDeletedUseCase(d.planRepo, d.subscriptionRepo, subscriptionNotifier),
			subscription.NewInvoicePaymentSucceededUseCase(d.subscriptionRepo, subscriptionNotifier),
			subscription.NewInvoicePaymentFailedUseCase(d.subscriptionRepo, subscriptionNotifier),
			upsertInvoiceUseCase,
		),
		subscriptionHandler: httphandler.NewSubscriptionHandler(
			createSubscriptionUseCase,
			createBillingPortalUseCase,
			getUserSubscriptionUseCase,
			getUserQuotaUsageUseCase,
		),
		invoiceHandler: httphandler.NewInvoiceHandler(getInvoicesUseCase),
	}
}
