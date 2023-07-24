package pkg

import (
	"net/http"

	"github.com/gin-gonic/gin"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type HttpServer struct {
	Engine      *gin.Engine
	Credentials *CredentialController
}

func (h *HttpServer) StartServer() string {
	h.Engine.GET("/provider", h.handle)

	h.Engine.Run("localhost:8080")
	return ""
}

func (h *HttpServer) handle(c *gin.Context) {
	s, e := h.Credentials.Secrets.Get(c.GetHeader("object"), metav1.GetOptions{})
	c.JSON(http.StatusOK, gin.H{
		"secret": s,
		"error":  e,
	})
}
