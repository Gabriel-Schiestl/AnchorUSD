package handlers

import (
	"context"

	"github.com/Gabriel-Schiestl/AnchorUSD/backend/internal/model"
	"github.com/gin-gonic/gin"
)

type UserReader interface {
	GetUserData(ctx context.Context, user string) (model.UserData, error)
}

func GetUserDataHandler(svc UserReader) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		user := ctx.Param("user")

		metrics, err := svc.GetUserData(ctx.Request.Context(), user)
		if err != nil {
			ctx.JSON(500, gin.H{"error": err.Error()})
			return
		}
		ctx.JSON(200, metrics)
	}
}