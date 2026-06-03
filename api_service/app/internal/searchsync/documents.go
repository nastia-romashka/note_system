package searchsync

import (
	"fmt"
	"strings"

	categoryclient "myproject/internal/client/category"
	noteclient "myproject/internal/client/note"
	searchclient "myproject/internal/client/search"
)

func BuildIndexedNote(note noteclient.Note, categories []categoryclient.Category, tags []noteclient.Tag) (searchclient.IndexedNote, error) {
	categoryName, ok := findCategoryName(categories, note.CategoryUuid)
	if !ok {
		return searchclient.IndexedNote{}, fmt.Errorf("category %s not found in user category tree", note.CategoryUuid)
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

	return searchclient.IndexedNote{
		ID:           note.Uuid,
		UserUUID:     note.UserUuid,
		Header:       note.Header,
		Body:         note.Body,
		ShortBody:    makeShortBody(note.Body),
		CategoryUUID: note.CategoryUuid,
		CategoryName: categoryName,
		TagUUIDs:     note.Tags,
		TagNames:     tagNames,
		CreatedDate:  note.CreatedDate,
	}, nil
}

func BuildIndexedNotes(notes []noteclient.Note, categories []categoryclient.Category, tags []noteclient.Tag) ([]searchclient.IndexedNote, error) {
	result := make([]searchclient.IndexedNote, 0, len(notes))
	for _, note := range notes {
		document, err := BuildIndexedNote(note, categories, tags)
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
