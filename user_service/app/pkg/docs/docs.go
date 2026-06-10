package docs

import (
	"encoding/json"
	"net/http"
)

const (
	swaggerURL = "/swagger"
	openapiURL = "/openapi.json"
)

// Register exposes Swagger UI and an OpenAPI schema for user_service.
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
			"title":       "User Service API",
			"version":     "1.1.0",
			"description": "Internal API for users, profiles, action history, authentication, and refresh sessions.",
		},
		"tags": []map[string]any{
			{"name": "system", "description": "Service health endpoints."},
			{"name": "users", "description": "User registration and lookup endpoints."},
			{"name": "profile", "description": "User profile read and update endpoints."},
			{"name": "actions", "description": "User activity audit endpoints."},
			{"name": "auth", "description": "Username and password authentication."},
			{"name": "sessions", "description": "Refresh-session lifecycle endpoints."},
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
								"service": map[string]any{"type": "string", "example": "user_service"},
							},
						}),
					},
				},
			},
			"/api/users": map[string]any{
				"post": map[string]any{
					"tags":        []string{"users"},
					"summary":     "Create user",
					"operationId": "createUser",
					"requestBody": jsonBody(schemaRef("CreateUserRequest"), true),
					"responses": map[string]any{
						"201": locationResponse("User created"),
						"400": errorResponse(),
					},
				},
			},
			"/api/users/{uuid}": map[string]any{
				"get": map[string]any{
					"tags":        []string{"users"},
					"summary":     "Get user by UUID",
					"operationId": "getUser",
					"parameters":  []map[string]any{pathParam("uuid", "User UUID")},
					"responses": map[string]any{
						"200": jsonResponse("User", schemaRef("User")),
						"400": errorResponse(),
						"404": errorResponse(),
					},
				},
			},
			"/api/users/{uuid}/profile": map[string]any{
				"get": map[string]any{
					"tags":        []string{"profile"},
					"summary":     "Get user profile",
					"operationId": "getUserProfile",
					"parameters":  []map[string]any{pathParam("uuid", "User UUID")},
					"responses": map[string]any{
						"200": jsonResponse("User profile", schemaRef("UserProfile")),
						"400": errorResponse(),
						"404": errorResponse(),
					},
				},
				"patch": map[string]any{
					"tags":        []string{"profile"},
					"summary":     "Update user profile",
					"operationId": "updateUserProfile",
					"parameters":  []map[string]any{pathParam("uuid", "User UUID")},
					"requestBody": jsonBody(schemaRef("UpdateUserProfileRequest"), true),
					"responses": map[string]any{
						"204": noContentResponse("Profile updated"),
						"400": errorResponse(),
						"404": errorResponse(),
					},
				},
			},
			"/api/users/{uuid}/actions": map[string]any{
				"get": map[string]any{
					"tags":        []string{"actions"},
					"summary":     "List user actions",
					"operationId": "getUserActions",
					"parameters": []map[string]any{
						pathParam("uuid", "User UUID"),
						queryParam("limit", "integer", "Max items, default 50", false),
						queryParam("offset", "integer", "Offset, default 0", false),
					},
					"responses": map[string]any{
						"200": jsonResponse("User actions", arraySchemaRef("UserAction")),
						"400": errorResponse(),
					},
				},
			},
			"/api/user-actions/{uuid}": map[string]any{
				"post": map[string]any{
					"tags":        []string{"actions"},
					"summary":     "Create user action",
					"operationId": "createUserAction",
					"parameters":  []map[string]any{pathParam("uuid", "User UUID")},
					"requestBody": jsonBody(schemaRef("CreateUserActionRequest"), true),
					"responses": map[string]any{
						"201": noContentResponse("Action created"),
						"400": errorResponse(),
					},
				},
			},
			"/api/users/authenticate": map[string]any{
				"post": map[string]any{
					"tags":        []string{"auth"},
					"summary":     "Authenticate user",
					"operationId": "authenticateUser",
					"requestBody": jsonBody(schemaRef("AuthUserRequest"), true),
					"responses": map[string]any{
						"200": jsonResponse("Authenticated user", schemaRef("User")),
						"400": errorResponse(),
						"401": errorResponse(),
					},
				},
			},
			"/api/user-sessions": map[string]any{
				"post": map[string]any{
					"tags":        []string{"sessions"},
					"summary":     "Create refresh session",
					"operationId": "createUserSession",
					"requestBody": jsonBody(schemaRef("CreateUserSessionRequest"), true),
					"responses": map[string]any{
						"201": noContentResponse("Session created"),
						"400": errorResponse(),
					},
				},
			},
			"/api/user-sessions/rotate": map[string]any{
				"post": map[string]any{
					"tags":        []string{"sessions"},
					"summary":     "Rotate refresh session",
					"operationId": "rotateUserSession",
					"requestBody": jsonBody(schemaRef("RotateUserSessionRequest"), true),
					"responses": map[string]any{
						"200": jsonResponse("Rotated session", schemaRef("UserSession")),
						"400": errorResponse(),
						"401": errorResponse(),
						"404": errorResponse(),
					},
				},
			},
			"/api/user-sessions/revoke": map[string]any{
				"post": map[string]any{
					"tags":        []string{"sessions"},
					"summary":     "Revoke refresh session",
					"operationId": "revokeUserSession",
					"requestBody": jsonBody(schemaRef("RevokeUserSessionRequest"), true),
					"responses": map[string]any{
						"204": noContentResponse("Session revoked"),
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
				"User": map[string]any{
					"type": "object",
					"properties": map[string]any{
						"uuid":         map[string]any{"type": "string", "format": "uuid", "example": "7f4a9b78-4e5f-4d58-8579-6d2588d63f6c"},
						"username":     map[string]any{"type": "string", "example": "tester01"},
						"email":        map[string]any{"type": "string", "format": "email", "example": "tester01@example.com"},
						"created_date": map[string]any{"type": "integer", "format": "int64", "example": 1760000000},
					},
				},
				"UserProfile": map[string]any{
					"type": "object",
					"properties": map[string]any{
						"uuid":          map[string]any{"type": "string", "format": "uuid", "example": "7f4a9b78-4e5f-4d58-8579-6d2588d63f6c"},
						"username":      map[string]any{"type": "string", "example": "tester01"},
						"email":         map[string]any{"type": "string", "format": "email", "example": "tester01@example.com"},
						"created_at":    map[string]any{"type": "integer", "format": "int64", "example": 1760000000},
						"last_login_at": map[string]any{"type": "integer", "format": "int64", "nullable": true},
					},
				},
				"UserAction": map[string]any{
					"type": "object",
					"properties": map[string]any{
						"uuid":        map[string]any{"type": "string", "format": "uuid"},
						"action":      map[string]any{"type": "string", "example": "auth.login"},
						"entity_type": map[string]any{"type": "string", "example": "user"},
						"entity_id":   map[string]any{"type": "string", "example": "7f4a9b78-4e5f-4d58-8579-6d2588d63f6c"},
						"status":      map[string]any{"type": "string", "example": "success"},
						"metadata":    map[string]any{"type": "object", "additionalProperties": true},
						"ip_address":  map[string]any{"type": "string", "example": "127.0.0.1"},
						"user_agent":  map[string]any{"type": "string", "example": "Mozilla/5.0"},
						"created_at":  map[string]any{"type": "integer", "format": "int64", "example": 1760000100},
					},
				},
				"UserSession": map[string]any{
					"type": "object",
					"properties": map[string]any{
						"uuid":         map[string]any{"type": "string", "format": "uuid"},
						"user_uuid":    map[string]any{"type": "string", "format": "uuid"},
						"user_agent":   map[string]any{"type": "string", "example": "ManualTest"},
						"ip_address":   map[string]any{"type": "string", "example": "127.0.0.1"},
						"created_at":   map[string]any{"type": "integer", "format": "int64", "example": 1760000200},
						"expires_at":   map[string]any{"type": "integer", "format": "int64", "example": 1760605000},
						"last_used_at": map[string]any{"type": "integer", "format": "int64", "nullable": true},
					},
				},
				"CreateUserRequest": map[string]any{
					"type":     "object",
					"required": []string{"username", "email", "password"},
					"properties": map[string]any{
						"username": map[string]any{
							"type":        "string",
							"minLength":   3,
							"maxLength":   32,
							"pattern":     "^\\S+$",
							"description": "Username without spaces.",
							"example":     "tester01",
						},
						"email": map[string]any{
							"type":        "string",
							"format":      "email",
							"description": "User email address.",
							"example":     "tester01@example.com",
						},
						"password": map[string]any{
							"type":        "string",
							"description": "Plain password. It will be hashed by the service.",
							"example":     "secret123",
						},
					},
				},
				"AuthUserRequest": map[string]any{
					"type":     "object",
					"required": []string{"username", "password"},
					"properties": map[string]any{
						"username": map[string]any{"type": "string", "example": "tester01"},
						"password": map[string]any{"type": "string", "example": "secret123"},
					},
				},
				"UpdateUserProfileRequest": map[string]any{
					"type": "object",
					"properties": map[string]any{
						"username":         map[string]any{"type": "string", "example": "tester01"},
						"email":            map[string]any{"type": "string", "format": "email", "example": "tester01@example.com"},
						"current_password": map[string]any{"type": "string", "example": "secret123"},
						"new_password":     map[string]any{"type": "string", "example": "newsecret123"},
					},
				},
				"CreateUserActionRequest": map[string]any{
					"type":     "object",
					"required": []string{"action", "entity_type"},
					"properties": map[string]any{
						"action":      map[string]any{"type": "string", "example": "profile.update"},
						"entity_type": map[string]any{"type": "string", "example": "user"},
						"entity_id":   map[string]any{"type": "string", "example": "7f4a9b78-4e5f-4d58-8579-6d2588d63f6c"},
						"status":      map[string]any{"type": "string", "example": "success"},
						"metadata":    map[string]any{"type": "object", "additionalProperties": true},
						"ip_address":  map[string]any{"type": "string", "example": "127.0.0.1"},
						"user_agent":  map[string]any{"type": "string", "example": "Mozilla/5.0"},
					},
				},
				"CreateUserSessionRequest": map[string]any{
					"type":     "object",
					"required": []string{"user_uuid", "refresh_token_hash", "expires_at"},
					"properties": map[string]any{
						"user_uuid":          map[string]any{"type": "string", "format": "uuid"},
						"refresh_token_hash": map[string]any{"type": "string", "example": "rt_hash_001"},
						"expires_at":         map[string]any{"type": "integer", "format": "int64", "example": 1760605000},
						"user_agent":         map[string]any{"type": "string", "example": "ManualTest"},
						"ip_address":         map[string]any{"type": "string", "example": "127.0.0.1"},
					},
				},
				"RotateUserSessionRequest": map[string]any{
					"type":     "object",
					"required": []string{"refresh_token_hash", "new_refresh_token_hash", "expires_at"},
					"properties": map[string]any{
						"refresh_token_hash":     map[string]any{"type": "string", "example": "rt_hash_001"},
						"new_refresh_token_hash": map[string]any{"type": "string", "example": "rt_hash_002"},
						"expires_at":             map[string]any{"type": "integer", "format": "int64", "example": 1760605000},
						"user_agent":             map[string]any{"type": "string", "example": "ManualTest"},
						"ip_address":             map[string]any{"type": "string", "example": "127.0.0.1"},
					},
				},
				"RevokeUserSessionRequest": map[string]any{
					"type":     "object",
					"required": []string{"refresh_token_hash"},
					"properties": map[string]any{
						"refresh_token_hash": map[string]any{"type": "string", "example": "rt_hash_002"},
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

const swaggerHTML = `<!doctype html>
<html lang="en">
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <title>User Service API Docs</title>
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
