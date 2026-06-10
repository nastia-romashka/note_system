package docs

import (
	"encoding/json"
	"net/http"
)

const (
	swaggerURL = "/swagger"
	openapiURL = "/openapi.json"
)

// Register exposes Swagger UI and an OpenAPI schema for note_service.
func Register(mux *http.ServeMux) {
	mux.HandleFunc(http.MethodGet+" "+swaggerURL, serveSwaggerUI)
	mux.HandleFunc(http.MethodGet+" "+openapiURL, serveOpenAPI)
}

func serveSwaggerUI(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(swaggerHTML))
}

func serveOpenAPI(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	encoder := json.NewEncoder(w)
	encoder.SetEscapeHTML(false)
	_ = encoder.Encode(openAPISpec())
}

func openAPISpec() map[string]any {
	return map[string]any{
		"openapi": "3.0.3",
		"info": map[string]any{
			"title":       "Note Service API",
			"version":     "1.0.0",
			"description": "Internal API for notes, tags, calendar events, and note statistics.",
		},
		"servers": []map[string]string{
			{"url": "/"},
		},
		"paths": map[string]any{
			"/health": map[string]any{
				"get": map[string]any{
					"tags":        []string{"system"},
					"summary":     "Health check",
					"operationId": "healthCheck",
					"responses": map[string]any{
						"200": jsonResponse("Service is healthy", map[string]any{
							"type": "object",
							"properties": map[string]any{
								"status":  map[string]any{"type": "string", "example": "ok"},
								"service": map[string]any{"type": "string", "example": "note_service"},
							},
						}),
					},
				},
			},
			"/api/notes": map[string]any{
				"get": map[string]any{
					"tags":        []string{"notes"},
					"summary":     "List notes by category",
					"operationId": "getNotesByCategory",
					"parameters": []map[string]any{
						queryParam("category_uuid", "string", "Category UUID", true),
						queryParam("user_uuid", "string", "User UUID", true),
					},
					"responses": map[string]any{
						"200": jsonResponse("Notes", arraySchemaRef("Note")),
						"400": errorResponse(),
					},
				},
				"post": map[string]any{
					"tags":        []string{"notes"},
					"summary":     "Create note",
					"operationId": "createNote",
					"requestBody": jsonBody(schemaRef("CreateNoteRequest"), true),
					"responses": map[string]any{
						"201": locationResponse("Note created"),
						"400": errorResponse(),
					},
				},
			},
			"/api/notes/{uuid}": map[string]any{
				"get": map[string]any{
					"tags":        []string{"notes"},
					"summary":     "Get note by UUID",
					"operationId": "getNote",
					"parameters": []map[string]any{
						pathParam("uuid", "Note UUID"),
						queryParam("user_uuid", "string", "User UUID", true),
					},
					"responses": map[string]any{
						"200": jsonResponse("Note", schemaRef("Note")),
						"400": errorResponse(),
						"404": errorResponse(),
					},
				},
				"patch": map[string]any{
					"tags":        []string{"notes"},
					"summary":     "Update note partially",
					"operationId": "updateNote",
					"parameters": []map[string]any{
						pathParam("uuid", "Note UUID"),
						queryParam("user_uuid", "string", "User UUID", true),
					},
					"requestBody": jsonBody(schemaRef("UpdateNoteRequest"), true),
					"responses": map[string]any{
						"204": noContentResponse("Note updated"),
						"400": errorResponse(),
						"404": errorResponse(),
					},
				},
				"delete": map[string]any{
					"tags":        []string{"notes"},
					"summary":     "Delete note",
					"operationId": "deleteNote",
					"parameters": []map[string]any{
						pathParam("uuid", "Note UUID"),
						queryParam("user_uuid", "string", "User UUID", true),
					},
					"responses": map[string]any{
						"204": noContentResponse("Note deleted"),
						"400": errorResponse(),
						"404": errorResponse(),
					},
				},
			},
			"/api/calendar": map[string]any{
				"get": map[string]any{
					"tags":        []string{"calendar"},
					"summary":     "Get notes with enabled events in time range",
					"operationId": "getCalendarNotes",
					"parameters": []map[string]any{
						queryParam("user_uuid", "string", "User UUID", true),
						queryParam("from", "integer", "Unix timestamp start", true),
						queryParam("to", "integer", "Unix timestamp end", true),
					},
					"responses": map[string]any{
						"200": jsonResponse("Calendar notes", arraySchemaRef("Note")),
						"400": errorResponse(),
					},
				},
			},
			"/api/stats": map[string]any{
				"get": map[string]any{
					"tags":        []string{"stats"},
					"summary":     "Get note and tag statistics",
					"operationId": "getNoteStats",
					"parameters": []map[string]any{
						queryParam("user_uuid", "string", "User UUID", true),
					},
					"responses": map[string]any{
						"200": jsonResponse("Statistics", schemaRef("NoteStats")),
						"400": errorResponse(),
					},
				},
			},
			"/api/tags": map[string]any{
				"get": map[string]any{
					"tags":        []string{"tags"},
					"summary":     "List tags",
					"operationId": "getTags",
					"parameters": []map[string]any{
						queryParam("user_uuid", "string", "User UUID", true),
						queryArrayParam("id", "Repeated tag UUID query parameters"),
					},
					"responses": map[string]any{
						"200": jsonResponse("Tags", arraySchemaRef("Tag")),
						"400": errorResponse(),
					},
				},
				"post": map[string]any{
					"tags":        []string{"tags"},
					"summary":     "Create tag",
					"operationId": "createTag",
					"requestBody": jsonBody(schemaRef("CreateTagRequest"), true),
					"responses": map[string]any{
						"201": locationResponse("Tag created"),
						"400": errorResponse(),
					},
				},
			},
			"/api/tags/{uuid}": map[string]any{
				"delete": map[string]any{
					"tags":        []string{"tags"},
					"summary":     "Delete tag",
					"operationId": "deleteTag",
					"parameters": []map[string]any{
						pathParam("uuid", "Tag UUID"),
						queryParam("user_uuid", "string", "User UUID", true),
					},
					"responses": map[string]any{
						"204": noContentResponse("Tag deleted"),
						"400": errorResponse(),
						"404": errorResponse(),
					},
				},
			},
		},
		"components": map[string]any{
			"schemas": map[string]any{
				"APIError": map[string]any{
					"type": "object",
					"properties": map[string]any{
						"message":           map[string]any{"type": "string"},
						"code":              map[string]any{"type": "string"},
						"developer_message": map[string]any{"type": "string"},
					},
				},
				"NoteEvent": map[string]any{
					"type": "object",
					"properties": map[string]any{
						"enabled":  map[string]any{"type": "boolean"},
						"start_at": map[string]any{"type": "integer", "format": "int64"},
						"end_at":   map[string]any{"type": "integer", "format": "int64"},
					},
				},
				"Note": map[string]any{
					"type": "object",
					"properties": map[string]any{
						"uuid":          map[string]any{"type": "string"},
						"user_uuid":     map[string]any{"type": "string"},
						"header":        map[string]any{"type": "string"},
						"body":          map[string]any{"type": "string"},
						"short_body":    map[string]any{"type": "string"},
						"created_date":  map[string]any{"type": "integer", "format": "int64"},
						"updated_at":    map[string]any{"type": "integer", "format": "int64"},
						"category_uuid": map[string]any{"type": "string"},
						"tags": map[string]any{
							"type":  "array",
							"items": map[string]any{"type": "string"},
						},
						"event": schemaRef("NoteEvent"),
					},
				},
				"CreateNoteRequest": map[string]any{
					"type":     "object",
					"required": []string{"user_uuid", "header", "body", "category_uuid"},
					"properties": map[string]any{
						"user_uuid":     map[string]any{"type": "string"},
						"header":        map[string]any{"type": "string"},
						"body":          map[string]any{"type": "string"},
						"category_uuid": map[string]any{"type": "string"},
						"tags": map[string]any{
							"type":  "array",
							"items": map[string]any{"type": "string"},
						},
						"event": schemaRef("NoteEvent"),
					},
				},
				"UpdateNoteRequest": map[string]any{
					"type": "object",
					"properties": map[string]any{
						"header":        map[string]any{"type": "string"},
						"body":          map[string]any{"type": "string"},
						"category_uuid": map[string]any{"type": "string"},
						"tags": map[string]any{
							"type":  "array",
							"items": map[string]any{"type": "string"},
						},
						"event": schemaRef("NoteEvent"),
					},
				},
				"NoteStats": map[string]any{
					"type": "object",
					"properties": map[string]any{
						"notes_count": map[string]any{"type": "integer", "format": "int64"},
						"tags_count":  map[string]any{"type": "integer", "format": "int64"},
					},
				},
				"Tag": map[string]any{
					"type": "object",
					"properties": map[string]any{
						"uuid":      map[string]any{"type": "string"},
						"user_uuid": map[string]any{"type": "string"},
						"name":      map[string]any{"type": "string"},
					},
				},
				"CreateTagRequest": map[string]any{
					"type":     "object",
					"required": []string{"user_uuid", "name"},
					"properties": map[string]any{
						"user_uuid": map[string]any{"type": "string"},
						"name":      map[string]any{"type": "string"},
					},
				},
			},
		},
	}
}

func schemaRef(name string) map[string]any {
	return map[string]any{"$ref": "#/components/schemas/" + name}
}

func arraySchemaRef(name string) map[string]any {
	return map[string]any{
		"type":  "array",
		"items": schemaRef(name),
	}
}

func jsonBody(schema map[string]any, required bool) map[string]any {
	return map[string]any{
		"required": required,
		"content": map[string]any{
			"application/json": map[string]any{
				"schema": schema,
			},
		},
	}
}

func jsonResponse(description string, schema map[string]any) map[string]any {
	return map[string]any{
		"description": description,
		"content": map[string]any{
			"application/json": map[string]any{
				"schema": schema,
			},
		},
	}
}

func errorResponse() map[string]any {
	return jsonResponse("Error response", schemaRef("APIError"))
}

func noContentResponse(description string) map[string]any {
	return map[string]any{"description": description}
}

func locationResponse(description string) map[string]any {
	return map[string]any{
		"description": description,
		"headers": map[string]any{
			"Location": map[string]any{
				"schema": map[string]any{"type": "string"},
			},
		},
	}
}

func pathParam(name, description string) map[string]any {
	return map[string]any{
		"name":        name,
		"in":          "path",
		"required":    true,
		"description": description,
		"schema": map[string]any{
			"type": "string",
		},
	}
}

func queryParam(name, schemaType, description string, required bool) map[string]any {
	param := map[string]any{
		"name":        name,
		"in":          "query",
		"required":    required,
		"description": description,
		"schema": map[string]any{
			"type": schemaType,
		},
	}

	if schemaType == "integer" {
		param["schema"] = map[string]any{
			"type":   "integer",
			"format": "int64",
		}
	}

	return param
}

func queryArrayParam(name, description string) map[string]any {
	return map[string]any{
		"name":        name,
		"in":          "query",
		"required":    false,
		"description": description,
		"schema": map[string]any{
			"type": "array",
			"items": map[string]any{
				"type": "string",
			},
		},
		"style":   "form",
		"explode": true,
	}
}

const swaggerHTML = `<!doctype html>
<html lang="en">
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <title>Note Service API Docs</title>
  <link rel="stylesheet" href="https://unpkg.com/swagger-ui-dist@5/swagger-ui.css">
  <style>
    body {
      margin: 0;
      background: #f6f7fb;
    }
    .topbar {
      display: none;
    }
  </style>
</head>
<body>
  <div id="swagger-ui"></div>
  <script src="https://unpkg.com/swagger-ui-dist@5/swagger-ui-bundle.js"></script>
  <script>
    window.ui = SwaggerUIBundle({
      url: '/openapi.json',
      dom_id: '#swagger-ui',
      deepLinking: true
    });
  </script>
</body>
</html>
`
