package realtime

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
