package http

import (
	"github.com/Gabriel-Schiestl/AnchorUSD/backend/internal/utils"
	"github.com/gin-gonic/gin"
)

var server *gin.Engine

func init() {
	logger := utils.GetLogger()
	logger.Info().Msg("Initializing Gin HTTP server")
	server = gin.Default()
}

func Run(addr string) error {
	logger := utils.GetLogger()
	logger.Info().Str("address", addr).Msg("Starting HTTP server")
	err := server.Run(addr)
	if err != nil {
		logger.Fatal().Err(err).Str("address", addr).Msg("Failed to start HTTP server")
	}
	return err
}
