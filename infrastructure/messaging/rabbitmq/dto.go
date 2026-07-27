package rabbitmq

import "go-api/domain/messaging"

type AnalyzeMessage = messaging.AnalyzeMessage
type StageDoneMessage = messaging.StageDoneMessage
type FailedMessage = messaging.FailedMessage

type MessagePayload struct {
	SecretKey string `json:"secret_key"`
	Message   any    `json:"message"`
}
