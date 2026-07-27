package worker

import (
	"context"

	rabbitmqDTO "go-api/infrastructure/messaging/rabbitmq"
	"go-api/infrastructure/messaging/security"
	aimodelUseCase "go-api/usecase/aimodel"
	pipelineUseCase "go-api/usecase/pipeline"
	"go-api/usecase/signal"

	amqp "github.com/rabbitmq/amqp091-go"
)

type AiModelHandler struct {
	parser                     *security.WorkerParser
	securityValidator          *security.WorkerSecurityValidator
	dispatcher                 *pipelineUseCase.Dispatcher
	analyzeMediaAiModelUseCase *aimodelUseCase.AnalyzeMediaAiModelUseCase
	createSignalUseCase        *signal.CreateSignalUseCase
}

func NewAiModelHandler(
	parser *security.WorkerParser,
	securityValidator *security.WorkerSecurityValidator,
	dispatcher *pipelineUseCase.Dispatcher,
	analyzeMediaAiModelUseCase *aimodelUseCase.AnalyzeMediaAiModelUseCase,
	createSignalUseCase *signal.CreateSignalUseCase,
) *AiModelHandler {
	return &AiModelHandler{
		parser:                     parser,
		securityValidator:          securityValidator,
		dispatcher:                 dispatcher,
		analyzeMediaAiModelUseCase: analyzeMediaAiModelUseCase,
		createSignalUseCase:        createSignalUseCase,
	}
}

func (h *AiModelHandler) process(ctx context.Context, message rabbitmqDTO.AnalyzeMessage) error {
	result, err := h.analyzeMediaAiModelUseCase.Execute(ctx, message.UserID, message.MediaKey)
	if err != nil {
		return err
	}

	_, err = h.createSignalUseCase.Execute(
		ctx,
		message.MediaID,
		"ai_model",
		result.Score,
		result.Confidence,
		result.Details,
	)
	return err
}

func (h *AiModelHandler) HandleMessage(ctx context.Context, message *amqp.Delivery) error {
	worker := NewStageWorkerHandler(
		"ai_model",
		h.parser,
		h.securityValidator,
		h.dispatcher,
		h.process,
	)

	return worker.HandleMessage(ctx, message)
}
