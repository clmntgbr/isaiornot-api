package realtime

// Client-facing CRUD actions published to Centrifugo.
// Domain events may be richer (uploaded, completed, failed…); they map to these verbs.
const (
	ActionCreated = "created"
	ActionUpdated = "updated"
	ActionDeleted = "deleted"

	EntityUser  = "user"
	EntityScan  = "scan"
	EntityMedia = "media"
)

func EventType(entity, action string) string {
	return entity + "." + action
}
