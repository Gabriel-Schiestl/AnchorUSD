package http

import (
	"github.com/gin-gonic/gin"
)

var server *gin.Engine

func init() {
	server = gin.Default()
}

func Run(addr string) error {
	return server.Run(addr)
}
