package handlers

import (
	"context"

	"github.com/Gabriel-Schiestl/AnchorUSD/backend/internal/model"
	"github.com/Gabriel-Schiestl/AnchorUSD/backend/internal/utils"
	"github.com/gin-gonic/gin"
)

type HealthFactorCalculator interface {
	CalculateMint(ctx context.Context, req model.CalculateMintRequest) (model.HealthFactorProjection, error)
	CalculateBurn(ctx context.Context, req model.CalculateBurnRequest) (model.HealthFactorProjection, error)
	CalculateDeposit(ctx context.Context, req model.CalculateDepositRequest) (model.HealthFactorProjection, error)
	CalculateRedeem(ctx context.Context, req model.CalculateRedeemRequest) (model.HealthFactorProjection, error)
}

func CalculateMintHandler(svc HealthFactorCalculator) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		logger := utils.GetLogger()
		var req model.CalculateMintRequest

		if err := ctx.ShouldBindJSON(&req); err != nil {
			logger.Warn().Err(err).Msg("Invalid request body for mint calculation")
			ctx.JSON(400, gin.H{"error": err.Error()})
			return
		}

		logger.Info().Str("user", req.Address).Str("amount", req.MintAmount).Msg("Calculating mint health factor")

		result, err := svc.CalculateMint(ctx.Request.Context(), req)
		if err != nil {
			logger.Error().Err(err).Str("user", req.Address).Msg("Failed to calculate mint health factor")
			ctx.JSON(500, gin.H{"error": err.Error()})
			return
		}

		logger.Info().Str("user", req.Address).Str("new_health_factor", result.HealthFactorAfter).Msg("Mint health factor calculated successfully")
		ctx.JSON(200, result)
	}
}

func CalculateBurnHandler(svc HealthFactorCalculator) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		logger := utils.GetLogger()
		var req model.CalculateBurnRequest

		if err := ctx.ShouldBindJSON(&req); err != nil {
			logger.Warn().Err(err).Msg("Invalid request body for burn calculation")
			ctx.JSON(400, gin.H{"error": err.Error()})
			return
		}

		logger.Info().Str("user", req.Address).Str("amount", req.BurnAmount).Msg("Calculating burn health factor")

		result, err := svc.CalculateBurn(ctx.Request.Context(), req)
		if err != nil {
			logger.Error().Err(err).Str("user", req.Address).Msg("Failed to calculate burn health factor")
			ctx.JSON(500, gin.H{"error": err.Error()})
			return
		}

		logger.Info().Str("user", req.Address).Str("new_health_factor", result.HealthFactorAfter).Msg("Burn health factor calculated successfully")
		ctx.JSON(200, result)
	}
}

func CalculateDepositHandler(svc HealthFactorCalculator) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		logger := utils.GetLogger()
		var req model.CalculateDepositRequest

		if err := ctx.ShouldBindJSON(&req); err != nil {
			logger.Warn().Err(err).Msg("Invalid request body for deposit calculation")
			ctx.JSON(400, gin.H{"error": err.Error()})
			return
		}

		logger.Info().Str("user", req.Address).Str("amount", req.DepositAmount).Str("token", req.TokenAddress).Msg("Calculating deposit health factor")

		result, err := svc.CalculateDeposit(ctx.Request.Context(), req)
		if err != nil {
			logger.Error().Err(err).Str("user", req.Address).Str("token", req.TokenAddress).Msg("Failed to calculate deposit health factor")
			ctx.JSON(500, gin.H{"error": err.Error()})
			return
		}

		logger.Info().Str("user", req.Address).Str("token", req.TokenAddress).Str("new_health_factor", result.HealthFactorAfter).Msg("Deposit health factor calculated successfully")
		ctx.JSON(200, result)
	}
}

func CalculateRedeemHandler(svc HealthFactorCalculator) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		logger := utils.GetLogger()
		var req model.CalculateRedeemRequest

		if err := ctx.ShouldBindJSON(&req); err != nil {
			logger.Warn().Err(err).Msg("Invalid request body for redeem calculation")
			ctx.JSON(400, gin.H{"error": err.Error()})
			return
		}

		logger.Info().Str("user", req.Address).Str("amount", req.RedeemAmount).Str("token", req.TokenAddress).Msg("Calculating redeem health factor")

		result, err := svc.CalculateRedeem(ctx.Request.Context(), req)
		if err != nil {
			logger.Error().Err(err).Str("user", req.Address).Str("token", req.TokenAddress).Msg("Failed to calculate redeem health factor")
			ctx.JSON(500, gin.H{"error": err.Error()})
			return
		}

		logger.Info().Str("user", req.Address).Str("token", req.TokenAddress).Str("new_health_factor", result.HealthFactorAfter).Msg("Redeem health factor calculated successfully")
		ctx.JSON(200, result)
	}
}
