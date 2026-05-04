package api

import (
	_ "embed"
	"net/http"

	"github.com/gin-gonic/gin"
)

//go:embed or-planner.openapi.yaml
var openapiSpec []byte

// HandleOpenApi serves the embedded OpenAPI specification.
func HandleOpenApi(ctx *gin.Context) {
	ctx.Data(http.StatusOK, "application/yaml", openapiSpec)
}
