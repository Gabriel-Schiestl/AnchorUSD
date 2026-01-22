package handlers

import (
	"context"

	"github.com/Gabriel-Schiestl/AnchorUSD/backend/internal/model"
	"github.com/Gabriel-Schiestl/AnchorUSD/backend/internal/utils"
	"github.com/gin-gonic/gin"
)

type HistoryReader interface {
	GetUserHistory(ctx context.Context, userAddress string) (model.HistoryData, error)
}

func GetHistoryHandler(svc HistoryReader) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		logger := utils.GetLogger()
		user := ctx.Param("user")

		logger.Info().Str("user", user).Str("endpoint", "/history/:user").Msg("Request received for user history")

		history, err := svc.GetUserHistory(ctx.Request.Context(), user)
		if err != nil {
			logger.Error().Err(err).Str("user", user).Msg("Failed to get user history")
			ctx.JSON(500, gin.H{"error": err.Error()})
			return
		}

		logger.Info().Str("user", user).Msg("User history retrieved successfully")
		ctx.JSON(200, history)
	}
}
