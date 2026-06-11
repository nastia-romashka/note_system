package searchsync

import (
	"fmt"
	fileclient "myproject/internal/client/file"
	"strings"

	categoryclient "myproject/internal/client/category"
	noteclient "myproject/internal/client/note"
	searchclient "myproject/internal/client/search"
)

func BuildIndexedNote(note noteclient.Note, categories []categoryclient.Category, tags []noteclient.Tag, files []fileclient.FileInfo) (searchclient.IndexedNote, error) {
	categoryName, ok := findCategoryName(categories, note.CategoryUuid)
	if !ok {
		return searchclient.IndexedNote{}, fmt.Errorf("category %s not found in workspace category tree", note.CategoryUuid)
	}

	tagNamesByID := make(map[string]string, len(tags))
	for _, tag := range tags {
		tagNamesByID[tag.Uuid] = tag.Name
	}

	tagNames := make([]string, 0, len(note.Tags))
	for _, tagUUID := range note.Tags {
		name, ok := tagNamesByID[tagUUID]
		if !ok {
			continue
		}
		tagNames = append(tagNames, name)
	}

	fileNamesText := buildFileNamesText(files)

	return searchclient.IndexedNote{
		ID:            note.Uuid,
		WorkspaceID:   note.WorkspaceID,
		Header:        note.Header,
		Body:          note.Body,
		ShortBody:     makeShortBody(note.Body),
		CategoryUUID:  note.CategoryUuid,
		CategoryName:  categoryName,
		TagUUIDs:      note.Tags,
		TagNames:      tagNames,
		FileNamesText: fileNamesText,
		CreatedAt:     note.CreatedDate,
		UpdatedAt:     note.UpdatedAt,
	}, nil
}

func BuildIndexedNotes(notes []noteclient.Note, categories []categoryclient.Category, tags []noteclient.Tag, filesByNote map[string][]fileclient.FileInfo) ([]searchclient.IndexedNote, error) {
	result := make([]searchclient.IndexedNote, 0, len(notes))
	for _, note := range notes {
		document, err := BuildIndexedNote(note, categories, tags, filesByNote[note.Uuid])
		if err != nil {
			return nil, err
		}
		result = append(result, document)
	}

	return result, nil
}

func CollectTagUUIDs(notes []noteclient.Note) []string {
	seen := make(map[string]struct{})
	result := make([]string, 0)
	for _, note := range notes {
		for _, tagUUID := range note.Tags {
			if tagUUID == "" {
				continue
			}
			if _, ok := seen[tagUUID]; ok {
				continue
			}
			seen[tagUUID] = struct{}{}
			result = append(result, tagUUID)
		}
	}

	return result
}

func findCategoryName(categories []categoryclient.Category, categoryUUID string) (string, bool) {
	for _, category := range categories {
		if category.Uuid == categoryUUID {
			return category.Name, true
		}
		if name, ok := findCategoryName(category.Children, categoryUUID); ok {
			return name, true
		}
	}

	return "", false
}

func makeShortBody(body string) string {
	body = strings.TrimSpace(body)
	if body == "" {
		return ""
	}

	body = strings.Join(strings.Fields(body), " ")
	if len(body) <= 120 {
		return body
	}

	return body[:117] + "..."
}

func buildFileNamesText(files []fileclient.FileInfo) string {
	if len(files) == 0 {
		return ""
	}

	seen := make(map[string]struct{}, len(files))
	names := make([]string, 0, len(files))
	for _, file := range files {
		name := strings.TrimSpace(file.Name)
		if name == "" {
			continue
		}
		key := strings.ToLower(name)
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		names = append(names, name)
	}

	return strings.Join(names, " ")
}
