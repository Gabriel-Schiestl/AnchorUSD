package handlers

import (
	"context"

	"github.com/Gabriel-Schiestl/AnchorUSD/backend/internal/model"
	"github.com/gin-gonic/gin"
)

type HealthFactorCalculator interface {
	CalculateMint(ctx context.Context, req model.CalculateMintRequest) (model.HealthFactorProjection, error)
	CalculateBurn(ctx context.Context, req model.CalculateBurnRequest) (model.HealthFactorProjection, error)
	CalculateDeposit(ctx context.Context, req model.CalculateDepositRequest) (model.HealthFactorProjection, error)
}

func CalculateMintHandler(svc HealthFactorCalculator) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var req model.CalculateMintRequest
		if err := ctx.ShouldBindJSON(&req); err != nil {
			ctx.JSON(400, gin.H{"error": err.Error()})
			return
		}

		result, err := svc.CalculateMint(ctx.Request.Context(), req)
		if err != nil {
			ctx.JSON(500, gin.H{"error": err.Error()})
			return
		}

		ctx.JSON(200, result)
	}
}

func CalculateBurnHandler(svc HealthFactorCalculator) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var req model.CalculateBurnRequest
		if err := ctx.ShouldBindJSON(&req); err != nil {
			ctx.JSON(400, gin.H{"error": err.Error()})
			return
		}

		result, err := svc.CalculateBurn(ctx.Request.Context(), req)
		if err != nil {
			ctx.JSON(500, gin.H{"error": err.Error()})
			return
		}

		ctx.JSON(200, result)
	}
}

func CalculateDepositHandler(svc HealthFactorCalculator) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var req model.CalculateDepositRequest
		if err := ctx.ShouldBindJSON(&req); err != nil {
			ctx.JSON(400, gin.H{"error": err.Error()})
			return
		}

		result, err := svc.CalculateDeposit(ctx.Request.Context(), req)
		if err != nil {
			ctx.JSON(500, gin.H{"error": err.Error()})
			return
		}

		ctx.JSON(200, result)
	}
}
