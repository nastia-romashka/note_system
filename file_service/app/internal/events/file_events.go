package events

import domainfile "file_service/internal/file"

const (
	FileUploadedEventType = "file.uploaded"
	FileDeletedEventType  = "file.deleted"
)

type Envelope[T any] struct {
	EventID      string `json:"event_id"`
	EventType    string `json:"event_type"`
	EventVersion int    `json:"event_version"`
	OccurredAt   int64  `json:"occurred_at"`
	Producer     string `json:"producer"`
	Payload      T      `json:"payload"`
}

type FileUploadedPayload struct {
	FileID       string `json:"file_id"`
	Name         string `json:"name"`
	UserUUID     string `json:"user_uuid"`
	WorkspaceID  string `json:"workspace_id"`
	NoteUUID     string `json:"note_uuid"`
	Size         int64  `json:"size"`
	ContentType  string `json:"content_type"`
}

type FileDeletedPayload struct {
	FileID      string `json:"file_id"`
	UserUUID    string `json:"user_uuid"`
	WorkspaceID string `json:"workspace_id"`
	NoteUUID    string `json:"note_uuid"`
}

func NewFileUploadedPayload(file domainfile.FileInfo) FileUploadedPayload {
	return FileUploadedPayload{
		FileID:      file.ID,
		Name:        file.Name,
		UserUUID:    file.UserUUID,
		WorkspaceID: file.WorkspaceID,
		NoteUUID:    file.NoteUUID,
		Size:        file.Size,
		ContentType: file.ContentType,
	}
}
