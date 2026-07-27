package di

import (
	httphandler "go-api/handler/http"
	"go-api/infrastructure/video"
	"go-api/usecase/media"
	"go-api/usecase/scan"
	"go-api/usecase/thumbnail"
)

type mediaBundle struct {
	mediaUploadWebhookHandler *httphandler.MediaUploadWebhookHandler
	mediaHandler              *httphandler.MediaHandler
}

func wireMedia(d *apiDeps, scanBundle scanBundle) mediaBundle {
	createMediaUseCase := media.NewCreateMediaUseCase(d.scanRepo, d.mediaRepo)
	generateImageThumbnailUseCase := thumbnail.NewGenerateImageThumbnailUseCase()
	generateThumbnailUseCase := media.NewGenerateThumbnailUseCase(d.storage, d.mediaRepo, generateImageThumbnailUseCase)
	publishMetadataUseCase := media.NewPublishMetadataUseCase(
		d.mediaRepo,
		d.publisher,
		d.centrifugoPublisher,
		d.env.AnalyzeQueues(),
	)
	updateScanStatusUseCase := scan.NewUpdateScanStatusUseCase(d.scanRepo)
	updateMediaStatusUseCase := media.NewUpdateMediaStatusUseCase(d.mediaRepo, updateScanStatusUseCase)
	processUploadedMediaUseCase := media.NewProcessUploadedMediaUseCase(
		d.storage,
		d.mediaRepo,
		createMediaUseCase,
		generateThumbnailUseCase,
		updateMediaStatusUseCase,
		publishMetadataUseCase,
		scanBundle.assertUploadAllowedUseCase,
		scanBundle.failScanUseCase,
		video.NewFrameExtractor(),
		generateImageThumbnailUseCase,
	)

	return mediaBundle{
		mediaUploadWebhookHandler: httphandler.NewMediaUploadWebhookHandler(d.env.StorageBucket, processUploadedMediaUseCase),
		mediaHandler: httphandler.NewMediaHandler(
			d.storage,
			media.NewGetMediaByIDUseCase(d.mediaRepo),
		),
	}
}
