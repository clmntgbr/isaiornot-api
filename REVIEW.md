# Architecture Review — Conformité & Axes d'Amélioration

Date: 2026-08-10  
Évaluation vs. `.cursor/rules/architecture.mdc`

---

## Résumé Exécutif

**Conformité globale: 7/10** — Architecture CQRS/event-driven bien fondée. Domain-driven design en place. Lacunes: pas de tests, observabilité manquante, workers trop spécialisés, state machines faibles.

---

## ✅ Points Forts

### 1. Event Versioning
**État:** Conforme  
Tous événements nommés `event.type.v1`:
- `media.created.v1`, `media.uploaded.v1`, `media.analyze.metadata.v1`, etc.
- `scan.created.v1`, `user.created.v1`, `signal.created.v1`

Prêt pour migration v2 si contrats changent.

### 2. CQRS Séparation  
**État:** Conforme
- Commands (`internal/application/command/`): synchrones, retournent IDs
- Queries (`internal/application/query/`): lectures dédiées
- Events (`internal/application/event/`): handlers asynchrones via RabbitMQ

### 3. Domain Events & Aggregates
**État:** Conforme
- Agrégats (Media, Scan, User) accumulent events en mémoire
- `PullEvents()` / `recordEvent()` pattern implémenté
- `EventID` généré à la création (pas au publish)
- Domains libres de tags GORM

### 4. Dedup & Idempotence
**État:** Conforme  
```go
reg.Register(eventType, dedup.With(
  dedupRepo,
  "stable_handler_name",
  handler.Handle,
))
```
Clé: `(event_id, handler_name)` — permet retry sans duplication.

### 5. Error Classification
**État:** Conforme
- `RetryableError` vs `NonRetryableError` définis
- Utilisés dans handlers: `messaging.Retryable(err)`, `messaging.NonRetryable(err)`
- DLQ configuré pour poison messages

### 6. Migrations SQL
**État:** Conforme
- Goose utilisé: `migrations/00001_init.sql` → `00007_media_status_pending.sql`
- Schema source of truth (pas AutoMigrate GORM)
- Cohérence persistence models ↔ DB (présumée — pas d'assertions de boot)

### 7. DI Container
**État:** Conforme
- Wiring manuel en `cmd/*/di/container.go`
- Pas de `CommandBus` générique — handlers appelés directement
- Composition lisible et typée

---

## ⚠️ Axes d'Amélioration Critiques

### 1. **Pas de Tests** (CRITIQUE)
**Symptôme:** 0 fichiers `*_test.go`
**Impact:** Zéro couverture. Refactos risqués. Idempotence non vérifiée.

**Actions:**
- Ajouter tests unitaires pour domain aggregates (state transitions, event recording)
- Tests intégration pour event handlers (outbox, dedup, retry flow)
- Tests e2e pour pipelines media (uploaded → metadata → heuristics → ai model → completed)
- Tests de chaos pour processus workers (kill/restart, dups, out-of-order events)

**Exemple:** Vérifier que `Media.MarkProcessing()` sur une media déjà `StatusCompleted` = no-op idempotent.

---

### 2. **State Machines Faibles** (ÉLEVÉ)
**Symptôme:** Status transitions not validated.  
Media: `Pending → Uploaded → Processing → Completed | Failed`  
Scan: pas de graphe défini.

**Code actuel:**
```go
func (m *Media) MarkProcessing() {
  if m.Status == StatusProcessing {
    return  // idempotent mais pas de vraie validation
  }
  m.setStatus(StatusProcessing)
}
```

**Problème:** Pas de "source of truth" sur transitions légales. Permet transitions invalides si logique change.

**Actions:**
- Définir diagrammes état → transitions légales (ASCII ou mermaid dans README)
- Ajouter validateur: `func (m *Media) CanTransitionTo(newStatus Status) error`
- Rejeter violations dans domain, pas app layer

Exemple:
```go
func (m *Media) MarkProcessing() error {
  if !m.CanTransitionTo(StatusProcessing) {
    return fmt.Errorf("invalid transition from %s to %s", m.Status, StatusProcessing)
  }
  ...
}
```

---

### 3. **Multiple Workers Spécialisés** (MOYEN)
**Symptôme:** 5 processus séparés.
```
cmd/
  api/       HTTP server
  worker/    outbox relay + event handlers (RabbitMQ consumer)
  cli/       migrations
  heuristic/ dedicated worker (externe?)
  aimodel/   dedicated worker (externe?)
  metadata/  dedicated worker (externe?)
```

**Guideline:** "One scalable worker process; internal dispatch via HandlerRegistry."

**Impact:** Déploiement/scaling fragmenté. Heuristic + metadata + AI model semblent externalisés (pas d'event handlers internes visible).

**Questions:**
- Heuristic/metadata/aimodel: Comment ils reçoivent les événements? Direct RabbitMQ? HTTP?
- Pourquoi pas une configuration unique du worker avec feature flags?
- Couplage vers externe (SightEngine?) fait que analysis logic n'est pas versionnable via domain events.

**Actions:**
- Mapper comment heuristic/metadata/aimodel sont déclenchés (`EnqueueAnalyzeHandler` les appelle via publisher direct?)
- Si externes: documenter contrats (event shape, routing key, retry)
- Si internes: consolider dans `cmd/worker` avec dispatching par event type
- Ajouter config pour enable/disable stages (pour déploiements progressifs)

---

### 4. **Observabilité Manquante** (MOYEN)
**Symptôme:** Pas de logging/tracing/metrics visible.

**Manque:**
- Logs structurés pour event processing (qui → quoi → résultat)
- Traces distribuées (e.g., media upload → 3 stages → completed)
- Métriques: event throughput, handler latency, retry rates, DLQ depth
- Audit: qui a changé quoi, quand, pourquoi (event sourcing audit trail)

**Actions:**
- Ajouter structured logging (zap/logrus) dans handlers
- Inclure: eventID, aggregateID, handler name, latency, error stack
- Ajouter spans OpenTelemetry pour traces distribuées
- Exposer métriques Prometheus (handler_duration_seconds, events_processed_total, etc.)
- Documenter run-book pour DLQ investigation

**Exemple:**
```go
func (h *EnqueueAnalyzeHandler) OnMediaUploaded(ctx context.Context, payload []byte) error {
  log.Infow("handling media uploaded",
    "mediaID", evt.MediaID,
    "contentType", evt.ContentType,
  )
  ...
}
```

---

### 5. **Documentations Manquantes** (MOYEN)

**Manque:**
- Pipeline d'analyse media (7 stages: uploaded → metadata → heuristics → AI model → completed). Comment? Why?
- Contrats événements (payload schemas)
- Runbook opérational (déployer worker, configurer RabbitMQ topology, tester retry)
- Architecture decision records (pourquoi 5 workers? Pourquoi pas une read/write DB split?)

**Actions:**
- Ajouter `docs/` avec:
  - `EVENTS.md` — catalog événements, payloads, handlers
  - `PIPELINE.md` — diagramme mermaid + timings média upload → completion
  - `RUNBOOK.md` — opérations (restart worker, investigate DLQ, scaling)
  - `ADR/` — decisions: "why 5 workers", "why not event sourcing read models", "why Centrifugo for realtime"

---

### 6. **Transaction Boundaries Ambigus** (MOYEN)
**Symptôme:** Pas toujours clair where outbox entries are stored.

`EnqueueAnalyzeHandler.enqueue()`:
```go
if err := h.publisher.Publish(ctx, port.EventEnvelope{...}); err != nil {
  return messaging.Retryable(err)
}
```

**Question:** `Publish()` = direct RabbitMQ ou outbox? Code suggests direct publish (not transactional). ❌ Risk: media marked but event not published if crash.

**Actions:**
- Clarifier où outbox.StoreEvents() s'appelle (should be in command, before publish)
- Si EnqueueAnalyzeHandler = event handler (async), pourquoi rePublish? Shouldn't just update media state?
- Séparation: command = sync write + outbox. Event handler = async effect (no new outbox, just side effects).

Current pattern seems mixed:
- `MarkProcessing()` = OK (command, stores outbox)
- `enqueue()` = unclear (event handler publishing new events?) — is it transactional?

---

### 7. **Ports & Adapters Clarification** (MOYEN)
**Symptôme:** Interfaces dans `internal/domain/port/` not fully visible.

**Questions:**
- How many adapters for EventPublisher? (RabbitMQ only, or also HTTP?)
- OutboxRepository: single write DB or could be split?
- Where are read repos? (should be in `internal/application/query/`, not domain/port/)

**Actions:**
- Audit all interfaces in `domain/port/`
- Ensure no adapter-specific logic leaks into ports (e.g., RabbitMQ connection details)
- Document in `docs/PORTS.md` which adapters exist per port

---

### 8. **Schéma Assertion Manquante** (BAS)
**État:** Migrations existent + persistence models. Mais pas de vérification au boot.

Guidelines: "Add a SQL migration when changing tables; update persistence models accordingly. ... migrate check / boot-time `schema.AssertModelsMatchDB` **must fail** on drift."

**Actions:**
- Ajouter `schema.go` avec assertions:
  ```go
  func AssertModelsMatchDB(db *gorm.DB) error {
    // Check MediaPersistence fields ↔ media table columns
    // Check ScanPersistence fields ↔ scan table columns
    return nil or error
  }
  ```
- Appeler dans `cmd/api/main.go` et `cmd/worker/main.go` avant startup
- Fail fast si drift (e.g., dev ajoute colonne en migration mais oublie struct field)

---

## 🔄 Améliorations Mineures (À Considérer)

1. **Handler Registry Logging**
   - Ajouter `reg.Handlers()` → map for introspection
   - Log au boot: "Registered X handlers for Y event types"

2. **Processed Events Audit**
   - Ajouter endpoint/query: "Show all handlers that processed event X"
   - Utile pour debug idempotence issues

3. **Event Payload Validation**
   - JSONSchema per event type (in domain or port)
   - Validate on consume (before calling handler)
   - Non-retryable si schema invalid

4. **Graceful Shutdown**
   - Worker: wait for in-flight handlers before exit
   - Relay: flush pending outbox entries
   - Currently: présumé mais pas visible

5. **Realtime Channels Security**
   - Centrifugo publishers expose channel names
   - Verify JWT tokens per channel (no leaks across scans/users)

---

## 📋 Checklist Priorisation

| Priorité | Item | Effort | Impact |
|----------|------|--------|--------|
| 🔴 CRITIQUE | Tests (unit + integration) | 5j | Confiance → refactos safe |
| 🟠 ÉLEVÉ | State machine validators | 1j | Prevent invalid transitions |
| 🟠 ÉLEVÉ | Observability (logs + metrics) | 3j | Troubleshooting, scaling decisions |
| 🟡 MOYEN | Documentations (EVENTS.md, PIPELINE.md) | 2j | Onboarding, ops runbooks |
| 🟡 MOYEN | Worker consolidation (or clarify multi-worker) | 2j | Ops clarity |
| 🟡 MOYEN | Transaction boundary audit | 1j | Exactly-once guarantees |
| 🟢 BAS | Schema assertions | 0.5j | Drift prevention |
| 🟢 BAS | Handler registry introspection | 0.5j | Debugging |

---

## 🎯 Recommandations Immédiates

### Court terme (1-2 sprints)
1. **Tests** — Start with domain aggregate tests (state transitions, event recording)
2. **Docs** — Write `EVENTS.md` (event catalog), `PIPELINE.md` (media analysis flow)
3. **Observability** — Add structured logging to key handlers + DLQ alerting

### Moyen terme (2-4 sprints)
1. **State machines** — Add validators + tests
2. **Audit** — Schema assertions + processed events audit
3. **Clarify workers** — Map heuristic/metadata/aimodel flow (external? internal?)

### Long terme
1. **Event sourcing read models** — If scaling read path later
2. **CQRS projection** — Denormalize media state for fast queries (instead of joins)
3. **Circuit breakers** — External service calls (SightEngine, Centrifugo) with fallbacks

---

## 📚 Références

- Current style: conventions.mdc ✅
- Missing: EVENTS.md (event catalog + schemas)
- Missing: PIPELINE.md (media processing flow diagram)
- Missing: RUNBOOK.md (operations playbook)
- Missing: ADRs (architectural decision records)

---

**Status:** Review complete. Recommend scheduling planning session to prioritize test + docs work.
