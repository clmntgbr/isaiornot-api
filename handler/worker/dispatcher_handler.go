package worker

import (
	"context"

	rabbitmqDTO "go-api/infrastructure/messaging/rabbitmq"
	"go-api/infrastructure/messaging/security"
	pipelineuc "go-api/usecase/pipeline"

	amqp "github.com/rabbitmq/amqp091-go"
)

type AnalyzeRequestHandler struct {
	parser            *security.WorkerParser
	securityValidator *security.WorkerSecurityValidator
	dispatcher        *pipelineuc.Dispatcher
}

func NewAnalyzeRequestHandler(
	parser *security.WorkerParser,
	securityValidator *security.WorkerSecurityValidator,
	dispatcher *pipelineuc.Dispatcher,
) *AnalyzeRequestHandler {
	return &AnalyzeRequestHandler{
		parser:            parser,
		securityValidator: securityValidator,
		dispatcher:        dispatcher,
	}
}

func (h *AnalyzeRequestHandler) HandleMessage(ctx context.Context, message *amqp.Delivery) error {
	var payload struct {
		SecretKey string                     `json:"secret_key"`
		Message   rabbitmqDTO.AnalyzeMessage `json:"message"`
	}
	if err := h.parser.ParseAndValidate(message.Body, &payload); err != nil {
		return err
	}

	if err := h.securityValidator.Validate(payload.SecretKey); err != nil {
		return err
	}

	return h.dispatcher.HandleAnalyzeRequest(ctx, payload.Message)
}

type StageDoneHandler struct {
	parser            *security.WorkerParser
	securityValidator *security.WorkerSecurityValidator
	dispatcher        *pipelineuc.Dispatcher
}

func NewStageDoneHandler(
	parser *security.WorkerParser,
	securityValidator *security.WorkerSecurityValidator,
	dispatcher *pipelineuc.Dispatcher,
) *StageDoneHandler {
	return &StageDoneHandler{
		parser:            parser,
		securityValidator: securityValidator,
		dispatcher:        dispatcher,
	}
}

func (h *StageDoneHandler) HandleMessage(ctx context.Context, message *amqp.Delivery) error {
	var payload struct {
		SecretKey string                       `json:"secret_key"`
		Message   rabbitmqDTO.StageDoneMessage `json:"message"`
	}
	if err := h.parser.ParseAndValidate(message.Body, &payload); err != nil {
		return err
	}

	if err := h.securityValidator.Validate(payload.SecretKey); err != nil {
		return err
	}

	return h.dispatcher.HandleStageDone(ctx, payload.Message)
}

type StageWorkerHandler struct {
	stage             string
	parser            *security.WorkerParser
	securityValidator *security.WorkerSecurityValidator
	dispatcher        *pipelineuc.Dispatcher
	process           func(ctx context.Context, message rabbitmqDTO.AnalyzeMessage) error
}

func NewStageWorkerHandler(
	stage string,
	parser *security.WorkerParser,
	securityValidator *security.WorkerSecurityValidator,
	dispatcher *pipelineuc.Dispatcher,
	process func(ctx context.Context, message rabbitmqDTO.AnalyzeMessage) error,
) *StageWorkerHandler {
	return &StageWorkerHandler{
		stage:             stage,
		parser:            parser,
		securityValidator: securityValidator,
		dispatcher:        dispatcher,
		process:           process,
	}
}

func (h *StageWorkerHandler) HandleMessage(ctx context.Context, message *amqp.Delivery) error {
	var payload struct {
		SecretKey string                     `json:"secret_key"`
		Message   rabbitmqDTO.AnalyzeMessage `json:"message"`
	}
	if err := h.parser.ParseAndValidate(message.Body, &payload); err != nil {
		return err
	}

	if err := h.securityValidator.Validate(payload.SecretKey); err != nil {
		return err
	}

	if err := h.process(ctx, payload.Message); err != nil {
		failedQueue := h.dispatcher.StageFailedQueue(h.stage)
		if failedQueue == "" {
			return err
		}

		publishErr := h.dispatcher.PublishFailed(ctx, failedQueue, rabbitmqDTO.FailedMessage{
			MediaID: payload.Message.MediaID,
			Stage:   h.stage,
			Error:   err.Error(),
		})
		if publishErr != nil {
			return publishErr
		}

		return nil
	}

	doneQueue := h.dispatcher.StageDoneQueue(h.stage)
	if doneQueue == "" {
		return nil
	}

	return h.dispatcher.PublishStageDone(ctx, doneQueue, rabbitmqDTO.StageDoneMessage{
		MediaID: payload.Message.MediaID,
		Stage:   h.stage,
	})
}
