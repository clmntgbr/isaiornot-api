package di

import (
	httphandler "go-api/handler/http"
	"go-api/usecase/scan"
	"go-api/usecase/subscription"
)

type scanBundle struct {
	failScanUseCase            *scan.FailScanUseCase
	assertUploadAllowedUseCase *subscription.AssertUploadAllowedUseCase
	scanHandler                *httphandler.ScanHandler
}

func wireScan(d *apiDeps) scanBundle {
	resolveEffectivePlanUseCase := subscription.NewResolveEffectivePlanUseCase(d.planRepo)
	getQuotaUsageUseCase := subscription.NewGetQuotaUsageUseCase(d.mediaRepo)
	assertUploadAllowedUseCase := subscription.NewAssertUploadAllowedUseCase(
		d.userRepo,
		d.subscriptionRepo,
		resolveEffectivePlanUseCase,
		getQuotaUsageUseCase,
	)
	failScanUseCase := scan.NewFailScanUseCase(d.scanRepo, d.centrifugoPublisher)
	generatePresignedUploadUrlUseCase := scan.NewGeneratePresignedUploadUrlUseCase(
		d.storage,
		d.scanRepo,
		d.mediaRepo,
	)
	historyCutoff := scan.NewHistoryCutoffResolver(
		d.userRepo,
		d.subscriptionRepo,
		resolveEffectivePlanUseCase,
	)

	return scanBundle{
		failScanUseCase:            failScanUseCase,
		assertUploadAllowedUseCase: assertUploadAllowedUseCase,
		scanHandler: httphandler.NewScanHandler(
			generatePresignedUploadUrlUseCase,
			scan.NewGetScanUseCase(d.scanRepo, historyCutoff),
			scan.NewGetScansUseCase(d.scanRepo, historyCutoff),
			scan.NewGetMediaStatisticsUseCase(d.scanRepo, historyCutoff),
		),
	}
}
