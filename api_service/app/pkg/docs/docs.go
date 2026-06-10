package docs

import (
	"encoding/json"
	"net/http"
)

const (
	swaggerURL = "/swagger"
	openapiURL = "/openapi.json"
)

// Register exposes a minimal Swagger UI and an OpenAPI schema for api_service.
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
	bearerSecurity := []map[string][]string{{"BearerAuth": {}}}

	return map[string]any{
		"openapi": "3.0.3",
		"info": map[string]any{
			"title":       "Note System API Gateway",
			"version":     "1.0.0",
			"description": "Gateway API for authentication, notes, categories, files, graph, profile, and search operations.",
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
						"200": map[string]any{
							"description": "Service is healthy",
							"content": map[string]any{
								"application/json": map[string]any{
									"schema": map[string]any{
										"type": "object",
										"properties": map[string]any{
											"status":  map[string]any{"type": "string", "example": "ok"},
											"service": map[string]any{"type": "string", "example": "api_service"},
										},
									},
								},
							},
						},
					},
				},
			},
			"/api/signup": map[string]any{
				"post": map[string]any{
					"tags":        []string{"auth"},
					"summary":     "Register a user",
					"operationId": "signUp",
					"requestBody": jsonBody(schemaRef("SignupRequest"), true),
					"responses": map[string]any{
						"201": jsonResponse("Access and refresh tokens", schemaRef("TokenResponse")),
						"400": errorResponse(),
					},
				},
			},
			"/api/auth": map[string]any{
				"post": map[string]any{
					"tags":        []string{"auth"},
					"summary":     "Authenticate with username and password",
					"operationId": "login",
					"requestBody": jsonBody(schemaRef("AuthRequest"), true),
					"responses": map[string]any{
						"201": jsonResponse("Access and refresh tokens", schemaRef("TokenResponse")),
						"400": errorResponse(),
						"401": errorResponse(),
					},
				},
				"put": map[string]any{
					"tags":        []string{"auth"},
					"summary":     "Refresh access token",
					"operationId": "refreshToken",
					"requestBody": jsonBody(schemaRef("RefreshRequest"), true),
					"responses": map[string]any{
						"201": jsonResponse("Refreshed access and refresh tokens", schemaRef("TokenResponse")),
						"400": errorResponse(),
						"401": errorResponse(),
					},
				},
				"delete": map[string]any{
					"tags":        []string{"auth"},
					"summary":     "Logout and revoke current refresh session",
					"operationId": "logout",
					"requestBody": jsonBody(schemaRef("RefreshRequest"), true),
					"responses": map[string]any{
						"204": noContentResponse("Refresh session revoked"),
						"400": errorResponse(),
					},
				},
			},
			"/api/heartbeat": map[string]any{
				"get": map[string]any{
					"tags":        []string{"system"},
					"summary":     "Authorized heartbeat",
					"operationId": "heartbeat",
					"security":    bearerSecurity,
					"responses": map[string]any{
						"204": noContentResponse("Authorized request accepted"),
						"401": errorResponse(),
					},
				},
			},
			"/api/categories": map[string]any{
				"get": map[string]any{
					"tags":        []string{"categories"},
					"summary":     "List user categories",
					"operationId": "getCategories",
					"security":    bearerSecurity,
					"responses": map[string]any{
						"200": jsonResponse("Category tree", arraySchemaRef("Category")),
						"401": errorResponse(),
					},
				},
				"post": map[string]any{
					"tags":        []string{"categories"},
					"summary":     "Create category",
					"operationId": "createCategory",
					"security":    bearerSecurity,
					"requestBody": jsonBody(schemaRef("CreateCategoryRequest"), true),
					"responses": map[string]any{
						"201": map[string]any{
							"description": "Category created",
							"headers": map[string]any{
								"Location": map[string]any{
									"schema": map[string]any{"type": "string"},
								},
							},
						},
						"400": errorResponse(),
						"401": errorResponse(),
					},
				},
			},
			"/api/categories/{uuid}": map[string]any{
				"patch": map[string]any{
					"tags":        []string{"categories"},
					"summary":     "Update category",
					"operationId": "updateCategory",
					"security":    bearerSecurity,
					"parameters":  []map[string]any{pathParam("uuid", "Category UUID")},
					"requestBody": jsonBody(schemaRef("UpdateCategoryRequest"), true),
					"responses": map[string]any{
						"204": noContentResponse("Category updated"),
						"400": errorResponse(),
						"401": errorResponse(),
						"404": errorResponse(),
					},
				},
				"delete": map[string]any{
					"tags":        []string{"categories"},
					"summary":     "Delete category recursively",
					"operationId": "deleteCategory",
					"security":    bearerSecurity,
					"parameters":  []map[string]any{pathParam("uuid", "Category UUID")},
					"responses": map[string]any{
						"204": noContentResponse("Category and nested resources deleted"),
						"401": errorResponse(),
						"404": errorResponse(),
					},
				},
			},
			"/api/tags": map[string]any{
				"get": map[string]any{
					"tags":        []string{"tags"},
					"summary":     "List tags",
					"operationId": "getTags",
					"security":    bearerSecurity,
					"parameters": []map[string]any{
						queryArrayParam("id", "Filter by repeated tag UUID query parameters"),
					},
					"responses": map[string]any{
						"200": jsonResponse("Tags", arraySchemaRef("Tag")),
						"401": errorResponse(),
					},
				},
				"post": map[string]any{
					"tags":        []string{"tags"},
					"summary":     "Create tag",
					"operationId": "createTag",
					"security":    bearerSecurity,
					"requestBody": jsonBody(schemaRef("CreateTagRequest"), true),
					"responses": map[string]any{
						"201": map[string]any{
							"description": "Tag created",
							"headers": map[string]any{
								"Location": map[string]any{
									"schema": map[string]any{"type": "string"},
								},
							},
						},
						"400": errorResponse(),
						"401": errorResponse(),
					},
				},
			},
			"/api/tags/{uuid}": map[string]any{
				"delete": map[string]any{
					"tags":        []string{"tags"},
					"summary":     "Delete tag",
					"operationId": "deleteTag",
					"security":    bearerSecurity,
					"parameters":  []map[string]any{pathParam("uuid", "Tag UUID")},
					"responses": map[string]any{
						"204": noContentResponse("Tag deleted"),
						"400": errorResponse(),
						"401": errorResponse(),
					},
				},
			},
			"/api/notes": map[string]any{
				"get": map[string]any{
					"tags":        []string{"notes"},
					"summary":     "List notes in category",
					"operationId": "getNotes",
					"security":    bearerSecurity,
					"parameters": []map[string]any{
						queryParam("category_uuid", "string", "Category UUID", true),
					},
					"responses": map[string]any{
						"200": jsonResponse("Notes", arraySchemaRef("Note")),
						"400": errorResponse(),
						"401": errorResponse(),
					},
				},
				"post": map[string]any{
					"tags":        []string{"notes"},
					"summary":     "Create note",
					"operationId": "createNote",
					"security":    bearerSecurity,
					"requestBody": jsonBody(schemaRef("CreateNoteRequest"), true),
					"responses": map[string]any{
						"201": map[string]any{
							"description": "Note created",
							"headers": map[string]any{
								"Location": map[string]any{
									"schema": map[string]any{"type": "string"},
								},
							},
						},
						"400": errorResponse(),
						"401": errorResponse(),
					},
				},
			},
			"/api/calendar": map[string]any{
				"get": map[string]any{
					"tags":        []string{"notes"},
					"summary":     "Get notes with enabled events in time range",
					"operationId": "getCalendarNotes",
					"security":    bearerSecurity,
					"parameters": []map[string]any{
						queryParam("from", "integer", "Unix timestamp start", true),
						queryParam("to", "integer", "Unix timestamp end", true),
					},
					"responses": map[string]any{
						"200": jsonResponse("Calendar notes", arraySchemaRef("Note")),
						"400": errorResponse(),
						"401": errorResponse(),
					},
				},
			},
			"/api/notes/{uuid}": map[string]any{
				"get": map[string]any{
					"tags":        []string{"notes"},
					"summary":     "Get note by UUID",
					"operationId": "getNote",
					"security":    bearerSecurity,
					"parameters":  []map[string]any{pathParam("uuid", "Note UUID")},
					"responses": map[string]any{
						"200": jsonResponse("Note", schemaRef("Note")),
						"400": errorResponse(),
						"401": errorResponse(),
						"404": errorResponse(),
					},
				},
				"patch": map[string]any{
					"tags":        []string{"notes"},
					"summary":     "Update note partially",
					"operationId": "updateNote",
					"security":    bearerSecurity,
					"parameters":  []map[string]any{pathParam("uuid", "Note UUID")},
					"requestBody": jsonBody(schemaRef("UpdateNoteRequest"), true),
					"responses": map[string]any{
						"204": noContentResponse("Note updated"),
						"400": errorResponse(),
						"401": errorResponse(),
						"404": errorResponse(),
					},
				},
				"delete": map[string]any{
					"tags":        []string{"notes"},
					"summary":     "Delete note",
					"operationId": "deleteNote",
					"security":    bearerSecurity,
					"parameters":  []map[string]any{pathParam("uuid", "Note UUID")},
					"responses": map[string]any{
						"204": noContentResponse("Note deleted"),
						"400": errorResponse(),
						"401": errorResponse(),
						"404": errorResponse(),
					},
				},
			},
			"/api/notes/{uuid}/duplicate": map[string]any{
				"post": map[string]any{
					"tags":        []string{"notes"},
					"summary":     "Duplicate note into another category",
					"operationId": "duplicateNote",
					"security":    bearerSecurity,
					"parameters":  []map[string]any{pathParam("uuid", "Source note UUID")},
					"requestBody": jsonBody(schemaRef("DuplicateNoteRequest"), true),
					"responses": map[string]any{
						"201": jsonResponse("Duplicated note", schemaRef("Note")),
						"400": errorResponse(),
						"401": errorResponse(),
						"404": errorResponse(),
					},
				},
			},
			"/api/notes/{uuid}/files": map[string]any{
				"get": map[string]any{
					"tags":        []string{"files"},
					"summary":     "List files for note",
					"operationId": "getNoteFiles",
					"security":    bearerSecurity,
					"parameters":  []map[string]any{pathParam("uuid", "Note UUID")},
					"responses": map[string]any{
						"200": jsonResponse("Files", arraySchemaRef("NoteFile")),
						"401": errorResponse(),
						"404": errorResponse(),
					},
				},
				"post": map[string]any{
					"tags":        []string{"files"},
					"summary":     "Upload file to note",
					"operationId": "uploadNoteFile",
					"security":    bearerSecurity,
					"parameters":  []map[string]any{pathParam("uuid", "Note UUID")},
					"requestBody": map[string]any{
						"required": true,
						"content": map[string]any{
							"multipart/form-data": map[string]any{
								"schema": map[string]any{
									"type":     "object",
									"required": []string{"file"},
									"properties": map[string]any{
										"file": map[string]any{
											"type":   "string",
											"format": "binary",
										},
									},
								},
							},
						},
					},
					"responses": map[string]any{
						"201": jsonResponse("Uploaded file metadata", schemaRef("NoteFile")),
						"400": errorResponse(),
						"401": errorResponse(),
						"404": errorResponse(),
					},
				},
			},
			"/api/notes/{uuid}/files/{fileId}": map[string]any{
				"get": map[string]any{
					"tags":        []string{"files"},
					"summary":     "Download note file",
					"operationId": "downloadNoteFile",
					"security":    bearerSecurity,
					"parameters": []map[string]any{
						pathParam("uuid", "Note UUID"),
						pathParam("fileId", "File identifier"),
					},
					"responses": map[string]any{
						"200": map[string]any{
							"description": "Binary file stream",
							"content": map[string]any{
								"application/octet-stream": map[string]any{
									"schema": map[string]any{"type": "string", "format": "binary"},
								},
							},
						},
						"401": errorResponse(),
						"404": errorResponse(),
					},
				},
				"delete": map[string]any{
					"tags":        []string{"files"},
					"summary":     "Delete note file",
					"operationId": "deleteNoteFile",
					"security":    bearerSecurity,
					"parameters": []map[string]any{
						pathParam("uuid", "Note UUID"),
						pathParam("fileId", "File identifier"),
					},
					"responses": map[string]any{
						"204": noContentResponse("File deleted"),
						"401": errorResponse(),
						"404": errorResponse(),
					},
				},
			},
			"/api/me": map[string]any{
				"get": map[string]any{
					"tags":        []string{"profile"},
					"summary":     "Get current user profile",
					"operationId": "getProfile",
					"security":    bearerSecurity,
					"responses": map[string]any{
						"200": jsonResponse("User profile", schemaRef("UserProfile")),
						"401": errorResponse(),
					},
				},
				"patch": map[string]any{
					"tags":        []string{"profile"},
					"summary":     "Update current user profile",
					"operationId": "updateProfile",
					"security":    bearerSecurity,
					"requestBody": jsonBody(schemaRef("UpdateProfileRequest"), true),
					"responses": map[string]any{
						"204": noContentResponse("Profile updated"),
						"400": errorResponse(),
						"401": errorResponse(),
					},
				},
			},
			"/api/me/actions": map[string]any{
				"get": map[string]any{
					"tags":        []string{"profile"},
					"summary":     "List current user actions",
					"operationId": "getActions",
					"security":    bearerSecurity,
					"parameters": []map[string]any{
						queryParam("limit", "integer", "Max items, default 50, capped at 100", false),
						queryParam("offset", "integer", "Offset, default 0", false),
					},
					"responses": map[string]any{
						"200": jsonResponse("User actions", arraySchemaRef("UserAction")),
						"400": errorResponse(),
						"401": errorResponse(),
					},
				},
			},
			"/api/me/summary": map[string]any{
				"get": map[string]any{
					"tags":        []string{"profile"},
					"summary":     "Get current user summary",
					"operationId": "getSummary",
					"security":    bearerSecurity,
					"responses": map[string]any{
						"200": jsonResponse("Profile summary", schemaRef("Summary")),
						"401": errorResponse(),
					},
				},
			},
			"/api/graph": map[string]any{
				"get": map[string]any{
					"tags":        []string{"graph"},
					"summary":     "Get graph data",
					"operationId": "getGraph",
					"security":    bearerSecurity,
					"responses": map[string]any{
						"200": jsonResponse("Graph data", schemaRef("GraphData")),
						"401": errorResponse(),
					},
				},
			},
			"/api/graph/links": map[string]any{
				"post": map[string]any{
					"tags":        []string{"graph"},
					"summary":     "Create user graph link",
					"operationId": "createGraphLink",
					"security":    bearerSecurity,
					"requestBody": jsonBody(schemaRef("GraphLinkRequest"), true),
					"responses": map[string]any{
						"204": noContentResponse("Graph link created"),
						"400": errorResponse(),
						"401": errorResponse(),
					},
				},
				"delete": map[string]any{
					"tags":        []string{"graph"},
					"summary":     "Delete user graph link",
					"operationId": "deleteGraphLink",
					"security":    bearerSecurity,
					"requestBody": jsonBody(schemaRef("GraphLinkRequest"), true),
					"responses": map[string]any{
						"204": noContentResponse("Graph link deleted"),
						"400": errorResponse(),
						"401": errorResponse(),
					},
				},
			},
			"/api/search/notes": map[string]any{
				"get": map[string]any{
					"tags":        []string{"search"},
					"summary":     "Search notes",
					"operationId": "searchNotes",
					"security":    bearerSecurity,
					"parameters": []map[string]any{
						queryParam("q", "string", "Search query", false),
						queryParam("category_uuid", "string", "Category filter", false),
						queryArrayParam("tag_uuid", "Repeated tag UUID query parameters"),
					},
					"responses": map[string]any{
						"200": jsonResponse("Indexed documents", arraySchemaRef("SearchNote")),
						"401": errorResponse(),
					},
				},
			},
			"/api/search/reindex": map[string]any{
				"post": map[string]any{
					"tags":        []string{"search"},
					"summary":     "Rebuild the user's search index",
					"operationId": "reindexNotes",
					"security":    bearerSecurity,
					"responses": map[string]any{
						"200": jsonResponse("Reindex result", schemaRef("ReindexResponse")),
						"401": errorResponse(),
					},
				},
			},
		},
		"components": map[string]any{
			"securitySchemes": map[string]any{
				"BearerAuth": map[string]any{
					"type":         "http",
					"scheme":       "bearer",
					"bearerFormat": "JWT",
				},
			},
			"schemas": map[string]any{
				"APIError": map[string]any{
					"type": "object",
					"properties": map[string]any{
						"message":           map[string]any{"type": "string"},
						"code":              map[string]any{"type": "string"},
						"developer_message": map[string]any{"type": "string"},
					},
				},
				"AuthRequest": map[string]any{
					"type":     "object",
					"required": []string{"username", "password"},
					"properties": map[string]any{
						"username": map[string]any{"type": "string"},
						"password": map[string]any{"type": "string"},
					},
				},
				"SignupRequest": map[string]any{
					"type":     "object",
					"required": []string{"username", "password", "email"},
					"properties": map[string]any{
						"username": map[string]any{"type": "string"},
						"password": map[string]any{"type": "string"},
						"email":    map[string]any{"type": "string", "format": "email"},
					},
				},
				"RefreshRequest": map[string]any{
					"type":     "object",
					"required": []string{"refresh_token"},
					"properties": map[string]any{
						"refresh_token": map[string]any{"type": "string"},
					},
				},
				"TokenResponse": map[string]any{
					"type":     "object",
					"required": []string{"token", "refresh_token"},
					"properties": map[string]any{
						"token":         map[string]any{"type": "string"},
						"refresh_token": map[string]any{"type": "string"},
					},
				},
				"Category": map[string]any{
					"type": "object",
					"properties": map[string]any{
						"uuid":        map[string]any{"type": "string"},
						"name":        map[string]any{"type": "string"},
						"color":       map[string]any{"type": "string"},
						"created_at":  map[string]any{"type": "integer", "format": "int64"},
						"user_uuid":   map[string]any{"type": "string"},
						"parent_uuid": map[string]any{"type": "string"},
						"children": map[string]any{
							"type":  "array",
							"items": schemaRef("Category"),
						},
					},
				},
				"CreateCategoryRequest": map[string]any{
					"type":     "object",
					"required": []string{"name"},
					"properties": map[string]any{
						"name":        map[string]any{"type": "string"},
						"color":       map[string]any{"type": "string"},
						"parent_uuid": map[string]any{"type": "string"},
					},
				},
				"UpdateCategoryRequest": map[string]any{
					"type": "object",
					"properties": map[string]any{
						"name":        map[string]any{"type": "string"},
						"color":       map[string]any{"type": "string"},
						"parent_uuid": map[string]any{"type": "string"},
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
					"required": []string{"header", "body", "category_uuid"},
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
				"DuplicateNoteRequest": map[string]any{
					"type":     "object",
					"required": []string{"category_uuid", "header"},
					"properties": map[string]any{
						"category_uuid": map[string]any{"type": "string"},
						"header":        map[string]any{"type": "string"},
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
					"required": []string{"name"},
					"properties": map[string]any{
						"name": map[string]any{"type": "string"},
					},
				},
				"NoteFile": map[string]any{
					"type": "object",
					"properties": map[string]any{
						"id":           map[string]any{"type": "string"},
						"name":         map[string]any{"type": "string"},
						"size":         map[string]any{"type": "integer", "format": "int64"},
						"content_type": map[string]any{"type": "string"},
					},
				},
				"UserProfile": map[string]any{
					"type": "object",
					"properties": map[string]any{
						"uuid":          map[string]any{"type": "string"},
						"username":      map[string]any{"type": "string"},
						"email":         map[string]any{"type": "string", "format": "email"},
						"created_at":    map[string]any{"type": "integer", "format": "int64"},
						"last_login_at": map[string]any{"type": "integer", "format": "int64", "nullable": true},
					},
				},
				"UpdateProfileRequest": map[string]any{
					"type": "object",
					"properties": map[string]any{
						"username":         map[string]any{"type": "string"},
						"email":            map[string]any{"type": "string", "format": "email"},
						"current_password": map[string]any{"type": "string"},
						"new_password":     map[string]any{"type": "string"},
					},
				},
				"UserAction": map[string]any{
					"type": "object",
					"properties": map[string]any{
						"uuid":        map[string]any{"type": "string"},
						"action":      map[string]any{"type": "string"},
						"entity_type": map[string]any{"type": "string"},
						"entity_id":   map[string]any{"type": "string"},
						"status":      map[string]any{"type": "string"},
						"metadata":    map[string]any{"type": "object", "additionalProperties": true},
						"ip_address":  map[string]any{"type": "string"},
						"user_agent":  map[string]any{"type": "string"},
						"created_at":  map[string]any{"type": "integer", "format": "int64"},
					},
				},
				"SummaryStats": map[string]any{
					"type": "object",
					"properties": map[string]any{
						"categories_count": map[string]any{"type": "integer", "format": "int64"},
						"notes_count":      map[string]any{"type": "integer", "format": "int64"},
						"tags_count":       map[string]any{"type": "integer", "format": "int64"},
						"files_count":      map[string]any{"type": "integer", "format": "int64"},
						"last_activity_at": map[string]any{"type": "integer", "format": "int64", "nullable": true},
					},
				},
				"Summary": map[string]any{
					"type": "object",
					"properties": map[string]any{
						"profile":         schemaRef("UserProfile"),
						"stats":           schemaRef("SummaryStats"),
						"upcoming_events": arraySchemaRef("Note"),
					},
				},
				"GraphNode": map[string]any{
					"type": "object",
					"properties": map[string]any{
						"id":            map[string]any{"type": "string"},
						"type":          map[string]any{"type": "string"},
						"label":         map[string]any{"type": "string"},
						"color":         map[string]any{"type": "string"},
						"category_uuid": map[string]any{"type": "string"},
						"created_at":    map[string]any{"type": "integer", "format": "int64"},
					},
				},
				"GraphEdge": map[string]any{
					"type": "object",
					"properties": map[string]any{
						"source": map[string]any{"type": "string"},
						"target": map[string]any{"type": "string"},
						"type":   map[string]any{"type": "string"},
					},
				},
				"GraphData": map[string]any{
					"type": "object",
					"properties": map[string]any{
						"nodes": arraySchemaRef("GraphNode"),
						"edges": arraySchemaRef("GraphEdge"),
					},
				},
				"GraphLinkRequest": map[string]any{
					"type":     "object",
					"required": []string{"source_id", "target_id"},
					"properties": map[string]any{
						"source_id": map[string]any{"type": "string"},
						"target_id": map[string]any{"type": "string"},
					},
				},
				"SearchNote": map[string]any{
					"type": "object",
					"properties": map[string]any{
						"id":            map[string]any{"type": "string"},
						"user_uuid":     map[string]any{"type": "string"},
						"header":        map[string]any{"type": "string"},
						"body":          map[string]any{"type": "string"},
						"short_body":    map[string]any{"type": "string"},
						"category_uuid": map[string]any{"type": "string"},
						"category_name": map[string]any{"type": "string"},
						"tag_uuids": map[string]any{
							"type":  "array",
							"items": map[string]any{"type": "string"},
						},
						"tag_names": map[string]any{
							"type":  "array",
							"items": map[string]any{"type": "string"},
						},
						"created_date": map[string]any{"type": "integer", "format": "int64"},
					},
				},
				"ReindexResponse": map[string]any{
					"type": "object",
					"properties": map[string]any{
						"indexed_notes":      map[string]any{"type": "integer"},
						"scanned_categories": map[string]any{"type": "integer"},
						"category_uuids": map[string]any{
							"type":  "array",
							"items": map[string]any{"type": "string"},
						},
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
  <title>Note System API Docs</title>
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
      deepLinking: true,
      persistAuthorization: true
    });
  </script>
</body>
</html>
`
