package events

const WorkspaceInviteAcceptedEventType = "workspace.invite.accepted"

type Envelope[T any] struct {
	EventID      string `json:"event_id"`
	EventType    string `json:"event_type"`
	EventVersion int    `json:"event_version"`
	OccurredAt   int64  `json:"occurred_at"`
	Producer     string `json:"producer"`
	Payload      T      `json:"payload"`
}

type WorkspaceInviteAcceptedPayload struct {
	WorkspaceID   string `json:"workspace_id"`
	WorkspaceName string `json:"workspace_name,omitempty"`
	WorkspaceType string `json:"workspace_type,omitempty"`
	UserUUID      string `json:"user_uuid"`
	InviteUUID    string `json:"invite_uuid,omitempty"`
}
