package handlers

import (
	"context"

	"github.com/gin-gonic/gin"
)

type MetricsReader interface {
	GetMetrics(ctx context.Context) (map[string]any, error)
}

func GetMetricsHandler(svc MetricsReader) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		metrics, err := svc.GetMetrics(ctx.Request.Context())
		if err != nil {
			ctx.JSON(500, gin.H{"error": err.Error()})
			return
		}
		ctx.JSON(200, metrics)
	}
}