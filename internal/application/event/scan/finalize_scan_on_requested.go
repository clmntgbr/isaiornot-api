package scan

import (
	"context"
	"encoding/json"
	"errors"
	"log"

	scancmd "go-api/internal/application/command/scan"
	"go-api/internal/application/messaging"
	domainscan "go-api/internal/domain/scan"

	"github.com/google/uuid"
)

type FinalizeScanOnRequestedHandler struct {
	finalizeScanHandler *scancmd.FinalizeScanHandler
}

func NewFinalizeScanOnRequestedHandler(
	finalizeScanHandler *scancmd.FinalizeScanHandler,
) *FinalizeScanOnRequestedHandler {
	return &FinalizeScanOnRequestedHandler{
		finalizeScanHandler: finalizeScanHandler,
	}
}

func (h *FinalizeScanOnRequestedHandler) Handle(ctx context.Context, payload []byte) error {
	var evt domainscan.ScanFinalizeRequested
	if err := json.Unmarshal(payload, &evt); err != nil {
		return messaging.NonRetryable(err)
	}

	scanID, err := uuid.Parse(evt.ScanID)
	if err != nil {
		return messaging.NonRetryable(err)
	}

	err = h.finalizeScanHandler.Handle(ctx, scancmd.FinalizeScanCommand{ScanID: scanID})
	if err != nil {
		if errors.Is(err, scancmd.ErrFinalizeScanNotFound) {
			return messaging.NonRetryable(err)
		}
		if errors.Is(err, scancmd.ErrFinalizeScanNotReady) {
			return nil
		}
		log.Printf(
			"finalize scan failed eventId=%s scanId=%s: %v",
			evt.ID,
			evt.ScanID,
			err,
		)
		return messaging.Retryable(err)
	}

	log.Printf(
		"event handled %s eventId=%s scanId=%s action=finalize_scan",
		domainscan.EventTypeScanFinalize,
		evt.ID,
		evt.ScanID,
	)
	return nil
}
