package config

import "go-api/domain/port"

func (c *Config) AnalyzeQueues() port.AnalyzeQueues {
	return port.AnalyzeQueues{
		Request:           c.AnalyzeRequestQueueName,
		MetadataAnalyze:   c.MetadataAnalyzeQueueName,
		HeuristicsAnalyze: c.HeuristicsAnalyzeQueueName,
		AiModelAnalyze:    c.AiModelAnalyzeQueueName,
		MetadataDone:      c.MetadataDoneQueueName,
		HeuristicsDone:    c.HeuristicsDoneQueueName,
		AiModelDone:       c.AiModelDoneQueueName,
		MetadataFailed:    c.MetadataFailedQueueName,
		HeuristicsFailed:  c.HeuristicsFailedQueueName,
		AiModelFailed:     c.AiModelFailedQueueName,
	}
}
