package handlers

import (
	"context"

	"github.com/Gabriel-Schiestl/AnchorUSD/backend/internal/model"
	"github.com/Gabriel-Schiestl/AnchorUSD/backend/internal/utils"
	"github.com/gin-gonic/gin"
)

type UserReader interface {
	GetUserData(ctx context.Context, user string) (model.UserData, error)
}

func GetUserDataHandler(svc UserReader) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		logger := utils.GetLogger()
		user := ctx.Param("user")

		logger.Info().Str("user", user).Str("endpoint", "/user/:user").Msg("Request received for user data")

		metrics, err := svc.GetUserData(ctx.Request.Context(), user)
		if err != nil {
			logger.Error().Err(err).Str("user", user).Msg("Failed to get user data")
			ctx.JSON(500, gin.H{"error": err.Error()})
			return
		}

		logger.Info().Str("user", user).Interface("data", metrics).Msg("User data retrieved successfully")
		ctx.JSON(200, metrics)
	}
}