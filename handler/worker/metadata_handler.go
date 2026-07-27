package worker

import (
	"context"

	rabbitmqDTO "go-api/infrastructure/messaging/rabbitmq"
	"go-api/infrastructure/messaging/security"
	metadataUseCase "go-api/usecase/metadata"
	pipelineUseCase "go-api/usecase/pipeline"
	"go-api/usecase/signal"

	amqp "github.com/rabbitmq/amqp091-go"
)

type MetadataHandler struct {
	parser                      *security.WorkerParser
	securityValidator           *security.WorkerSecurityValidator
	dispatcher                  *pipelineUseCase.Dispatcher
	analyzeMediaMetadataUseCase *metadataUseCase.AnalyzeMediaMetadataUseCase
	createSignalUseCase         *signal.CreateSignalUseCase
}

func NewMetadataHandler(
	parser *security.WorkerParser,
	securityValidator *security.WorkerSecurityValidator,
	dispatcher *pipelineUseCase.Dispatcher,
	analyzeMediaMetadataUseCase *metadataUseCase.AnalyzeMediaMetadataUseCase,
	createSignalUseCase *signal.CreateSignalUseCase,
) *MetadataHandler {
	return &MetadataHandler{
		parser:                      parser,
		securityValidator:           securityValidator,
		dispatcher:                  dispatcher,
		analyzeMediaMetadataUseCase: analyzeMediaMetadataUseCase,
		createSignalUseCase:         createSignalUseCase,
	}
}

func (h *MetadataHandler) process(ctx context.Context, message rabbitmqDTO.AnalyzeMessage) error {
	result, err := h.analyzeMediaMetadataUseCase.Execute(ctx, message.UserID, message.MediaKey)
	if err != nil {
		return err
	}

	_, err = h.createSignalUseCase.Execute(
		ctx,
		message.MediaID,
		"metadata",
		result.Score,
		result.Confidence,
		result.Details,
	)
	return err
}

func (h *MetadataHandler) HandleMessage(ctx context.Context, message *amqp.Delivery) error {
	worker := NewStageWorkerHandler(
		"metadata",
		h.parser,
		h.securityValidator,
		h.dispatcher,
		h.process,
	)

	return worker.HandleMessage(ctx, message)
}
