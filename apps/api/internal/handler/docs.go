package handler

import (
	"github.com/gofiber/fiber/v2"
)

// SetupDocs sets up Scalar API documentation
func SetupDocs(app *fiber.App) {
	// Serve Scalar UI at /docs
	app.Get("/docs", func(c *fiber.Ctx) error {
		c.Set("Content-Type", "text/html; charset=utf-8")
		c.Set("Cache-Control", "public, max-age=3600")
		return c.Status(fiber.StatusOK).SendString(scalarHTML)
	})

	// Serve OpenAPI spec at /openapi.json
	app.Get("/openapi.json", func(c *fiber.Ctx) error {
		c.Set("Content-Type", "application/json; charset=utf-8")
		c.Set("Cache-Control", "public, max-age=3600")
		return c.Status(fiber.StatusOK).SendString(openAPISpec)
	})

	// Redirect root to docs
	app.Get("/", func(c *fiber.Ctx) error {
		return c.Redirect("/docs")
	})
}

// Scalar HTML using CDN
const scalarHTML = `<!DOCTYPE html>
<html>
  <head>
    <title>NER Studio API Documentation</title>
    <meta charset="utf-8" />
    <meta name="viewport" content="width=device-width, initial-scale=1" />
    <link rel="icon" type="image/png" href="https://scalar.com/favicon.png" />
  </head>
  <body>
    <script
      id="api-reference"
      data-url="/openapi.json"
      data-proxy-url="https://api.scalar.com/request-proxy"
      data-theme="purple"></script>
    <script src="https://cdn.jsdelivr.net/npm/@scalar/api-reference@latest"></script>
  </body>
</html>`

// OpenAPI specification
const openAPISpec = `{
  "openapi": "3.1.0",
  "info": {
    "title": "NER Studio API",
    "description": "AI Image Generation Platform for Creative Teams",
    "version": "1.0.0",
    "contact": {
      "name": "NER Studio Team"
    }
  },
  "servers": [
    {
      "url": "http://localhost:5005",
      "description": "Development server"
    }
  ],
  "tags": [
    {
      "name": "Auth",
      "description": "Authentication endpoints"
    },
    {
      "name": "Generations",
      "description": "Image generation management"
    },
    {
      "name": "Gallery",
      "description": "User gallery and images"
    },
    {
      "name": "Uploads",
      "description": "File upload endpoints"
    },
    {
      "name": "Callbacks",
      "description": "Provider webhook callbacks"
    }
  ],
  "paths": {
    "/health": {
      "get": {
        "summary": "Health check",
        "description": "Check API and database status",
        "tags": ["Health"],
        "responses": {
          "200": {
            "description": "API is healthy",
            "content": {
              "application/json": {
                "schema": {
                  "type": "object",
                  "properties": {
                    "status": { "type": "string", "example": "ok" },
                    "version": { "type": "string", "example": "1.0.0" },
                    "database": { "type": "string", "example": "connected" }
                  }
                }
              }
            }
          }
        }
      }
    },
    "/api/v1/auth/register": {
      "post": {
        "summary": "Register new user",
        "description": "Create a new user account with organization",
        "tags": ["Auth"],
        "requestBody": {
          "required": true,
          "content": {
            "application/json": {
              "schema": {
                "type": "object",
                "required": ["email", "password", "full_name", "org_name"],
                "properties": {
                  "email": { "type": "string", "format": "email", "example": "user@example.com" },
                  "password": { "type": "string", "minLength": 8, "example": "password123" },
                  "full_name": { "type": "string", "example": "John Doe" },
                  "org_name": { "type": "string", "example": "Acme Corp" }
                }
              }
            }
          }
        },
        "responses": {
          "201": {
            "description": "User created successfully",
            "content": {
              "application/json": {
                "schema": {
                  "type": "object",
                  "properties": {
                    "token": { "type": "string" },
                    "user": {
                      "type": "object",
                      "properties": {
                        "id": { "type": "string", "format": "uuid" },
                        "email": { "type": "string" },
                        "name": { "type": "string" },
                        "role": { "type": "string", "enum": ["admin", "member"] }
                      }
                    },
                    "organization": {
                      "type": "object",
                      "properties": {
                        "id": { "type": "string", "format": "uuid" },
                        "name": { "type": "string" },
                        "slug": { "type": "string" }
                      }
                    }
                  }
                }
              }
            }
          },
          "400": {
            "description": "Invalid input"
          }
        }
      }
    },
    "/api/v1/auth/login": {
      "post": {
        "summary": "User login",
        "description": "Authenticate user and get JWT token",
        "tags": ["Auth"],
        "requestBody": {
          "required": true,
          "content": {
            "application/json": {
              "schema": {
                "type": "object",
                "required": ["email", "password"],
                "properties": {
                  "email": { "type": "string", "format": "email" },
                  "password": { "type": "string" }
                }
              }
            }
          }
        },
        "responses": {
          "200": {
            "description": "Login successful",
            "content": {
              "application/json": {
                "schema": {
                  "type": "object",
                  "properties": {
                    "token": { "type": "string" },
                    "user": {
                      "type": "object",
                      "properties": {
                        "id": { "type": "string" },
                        "email": { "type": "string" },
                        "name": { "type": "string" },
                        "role": { "type": "string" }
                      }
                    }
                  }
                }
              }
            }
          },
          "401": {
            "description": "Invalid credentials"
          }
        }
      }
    },
    "/api/v1/auth/refresh": {
      "post": {
        "summary": "Refresh JWT token",
        "description": "Get a new JWT token using current token",
        "tags": ["Auth"],
        "security": [{ "bearerAuth": [] }],
        "responses": {
          "200": {
            "description": "New token generated",
            "content": {
              "application/json": {
                "schema": {
                  "type": "object",
                  "properties": {
                    "token": { "type": "string" }
                  }
                }
              }
            }
          }
        }
      }
    },
    "/api/v1/generations": {
      "get": {
        "summary": "List generations",
        "description": "Get all generations for the current user",
        "tags": ["Generations"],
        "security": [{ "bearerAuth": [] }],
        "parameters": [
          {
            "name": "limit",
            "in": "query",
            "schema": { "type": "integer", "default": 20 }
          },
          {
            "name": "offset",
            "in": "query",
            "schema": { "type": "integer", "default": 0 }
          }
        ],
        "responses": {
          "200": {
            "description": "List of generations"
          }
        }
      },
      "post": {
        "summary": "Create generation",
        "description": "Start a new image generation job",
        "tags": ["Generations"],
        "security": [{ "bearerAuth": [] }],
        "requestBody": {
          "required": true,
          "content": {
            "application/json": {
              "schema": {
                "type": "object",
                "required": ["base_prompt", "provider_id"],
                "properties": {
                  "base_prompt": { "type": "string", "description": "Main prompt for image generation" },
                  "reference_images": { "type": "array", "items": { "type": "string" } },
                  "product_images": { "type": "array", "items": { "type": "string" } },
                  "provider_id": { "type": "string", "format": "uuid" },
                  "num_variations": { "type": "integer", "minimum": 1, "maximum": 10, "default": 4 }
                }
              }
            }
          }
        },
        "responses": {
          "202": {
            "description": "Generation started"
          }
        }
      }
    },
    "/api/v1/generations/{id}": {
      "get": {
        "summary": "Get generation",
        "description": "Get a specific generation by ID",
        "tags": ["Generations"],
        "security": [{ "bearerAuth": [] }],
        "parameters": [
          {
            "name": "id",
            "in": "path",
            "required": true,
            "schema": { "type": "string", "format": "uuid" }
          }
        ],
        "responses": {
          "200": {
            "description": "Generation details"
          }
        }
      }
    },
    "/api/v1/gallery": {
      "get": {
        "summary": "Get gallery",
        "description": "Get user's generated images",
        "tags": ["Gallery"],
        "security": [{ "bearerAuth": [] }],
        "responses": {
          "200": {
            "description": "List of images"
          }
        }
      }
    },
    "/api/v1/uploads": {
      "post": {
        "summary": "Upload image",
        "description": "Upload an image to R2 storage",
        "tags": ["Uploads"],
        "security": [{ "bearerAuth": [] }],
        "requestBody": {
          "content": {
            "multipart/form-data": {
              "schema": {
                "type": "object",
                "properties": {
                  "image": { "type": "string", "format": "binary" },
                  "folder": { "type": "string", "enum": ["references", "products", "uploads"] }
                }
              }
            }
          }
        },
        "responses": {
          "200": {
            "description": "Upload successful"
          }
        }
      }
    },
    "/api/v1/callbacks/{provider}": {
      "post": {
        "summary": "Provider callback",
        "description": "Webhook endpoint for AI providers",
        "tags": ["Callbacks"],
        "parameters": [
          {
            "name": "provider",
            "in": "path",
            "required": true,
            "schema": { "type": "string" }
          }
        ],
        "requestBody": {
          "content": {
            "application/json": {
              "schema": {
                "type": "object",
                "properties": {
                  "task_id": { "type": "string" },
                  "status": { "type": "string" },
                  "image_url": { "type": "string" }
                }
              }
            }
          }
        },
        "responses": {
          "200": {
            "description": "Callback processed"
          }
        }
      }
    }
  },
  "components": {
    "securitySchemes": {
      "bearerAuth": {
        "type": "http",
        "scheme": "bearer",
        "bearerFormat": "JWT",
        "description": "JWT token obtained from /auth/login or /auth/register"
      }
    }
  }
}`
