package events

import handlermodel "note_service/internal/handlers/notes"

const (
	NoteCreatedEventType = "note.created"
	NoteUpdatedEventType = "note.updated"
	NoteDeletedEventType = "note.deleted"
)

type Envelope[T any] struct {
	EventID      string `json:"event_id"`
	EventType    string `json:"event_type"`
	EventVersion int    `json:"event_version"`
	OccurredAt   int64  `json:"occurred_at"`
	Producer     string `json:"producer"`
	Payload      T      `json:"payload"`
}

type NoteUpdatedPayload struct {
	NoteUUID       string                  `json:"note_uuid"`
	UserUUID       string                  `json:"user_uuid,omitempty"`
	WorkspaceID    string                  `json:"workspace_id"`
	AuthorUserUUID string                  `json:"author_user_uuid,omitempty"`
	CategoryUUID   string                  `json:"category_uuid"`
	Header         string                  `json:"header"`
	Body           string                  `json:"body,omitempty"`
	ShortBody      string                  `json:"short_body,omitempty"`
	Tags           []string                `json:"tags,omitempty"`
	Event          *handlermodel.NoteEvent `json:"event,omitempty"`
	CreatedAt      int64                   `json:"created_at"`
	UpdatedAt      int64                   `json:"updated_at"`
}

type NoteDeletedPayload struct {
	NoteUUID    string `json:"note_uuid"`
	UserUUID    string `json:"user_uuid,omitempty"`
	WorkspaceID string `json:"workspace_id"`
}
