package port

import "context"

type MessagePublisher interface {
	Publish(ctx context.Context, queueName string, message any) error
}

type AnalyzeQueues struct {
	Request          string
	MetadataAnalyze  string
	HeuristicsAnalyze string
	AiModelAnalyze   string
	MetadataDone     string
	HeuristicsDone   string
	AiModelDone      string
	MetadataFailed   string
	HeuristicsFailed string
	AiModelFailed    string
}
