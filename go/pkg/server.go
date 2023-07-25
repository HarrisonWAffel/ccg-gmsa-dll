package pkg

import (
	"net"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type HttpServer struct {
	Engine      *gin.Engine
	Credentials *CredentialController
}

func (h *HttpServer) StartServer(errChan chan error) string {
	h.Engine.GET("/provider", h.handle)

	// use a host allocated port
	ln, _ := net.Listen("tcp", ":0")
	go func() {
		err := http.Serve(ln, h.Engine)
		errChan <- err
	}()

	// let the server come up and
	// be assigned a port
	time.Sleep(250 * time.Millisecond)
	_, port, _ := net.SplitHostPort(ln.Addr().String())
	return port
}

func (h *HttpServer) handle(c *gin.Context) {
	s, e := h.Credentials.Secrets.Get(c.GetHeader("object"), metav1.GetOptions{})
	c.JSON(http.StatusOK, gin.H{
		"secret": s,
		"error":  e,
	})
}
