package signal

import (
	"context"
	"fmt"

	"go-api/internal/domain/port"
	domainsignal "go-api/internal/domain/signal"

	"github.com/google/uuid"
)

type UpsertSignalCommand struct {
	MediaID    uuid.UUID
	Name       domainsignal.Name
	Score      int
	Confidence domainsignal.ConfidenceLevel
	Details    []string
}

type UpsertSignalHandler struct {
	signalRepo domainsignal.SignalWriteRepository
	outbox     port.OutboxRepository
}

func NewUpsertSignalHandler(
	signalRepo domainsignal.SignalWriteRepository,
	outbox port.OutboxRepository,
) *UpsertSignalHandler {
	return &UpsertSignalHandler{
		signalRepo: signalRepo,
		outbox:     outbox,
	}
}

func (h *UpsertSignalHandler) Handle(ctx context.Context, cmd UpsertSignalCommand) (*domainsignal.Signal, error) {
	existing, err := h.signalRepo.GetByMediaIDAndName(ctx, cmd.MediaID, string(cmd.Name))
	if err != nil {
		return nil, err
	}

	err = h.signalRepo.WithTransaction(ctx, func(txCtx context.Context) error {
		if existing != nil {
			existing.Update(cmd.Score, cmd.Confidence, cmd.Details)
			if err := h.signalRepo.Update(txCtx, existing); err != nil {
				return err
			}
			return h.outbox.StoreEvents(txCtx, existing.PullEvents())
		}

		created := domainsignal.NewSignal(
			cmd.MediaID,
			string(cmd.Name),
			cmd.Score,
			cmd.Confidence,
			cmd.Details,
		)
		if err := h.signalRepo.Save(txCtx, created); err != nil {
			return err
		}
		existing = created
		return h.outbox.StoreEvents(txCtx, created.PullEvents())
	})
	if err != nil {
		return nil, fmt.Errorf("failed to upsert signal: %w", err)
	}

	return existing, nil
}
