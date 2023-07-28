package pkg

import (
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type HttpServer struct {
	Engine      *gin.Engine
	Credentials *CredentialClient
}

func (h *HttpServer) StartServer(errChan chan error, dirName string) string {
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
	fmt.Println("Listening on port ", port)
	return port
}

func (h *HttpServer) handle(c *gin.Context) {
	secret := c.GetHeader("object")
	if secret == "" {
		c.Status(http.StatusBadRequest)
		fmt.Println("Received request with no object")
		return
	}

	s, err := h.Credentials.Secrets.Get(c.GetHeader("object"), metav1.GetOptions{})
	if errors.IsForbidden(err) || errors.IsNotFound(err) {
		c.Status(http.StatusNotFound)
		fmt.Println(err)
		return
	}

	c.JSON(http.StatusOK, Response{
		Username:   s.StringData["username"],
		Password:   s.StringData["password"],
		DomainName: s.StringData["domainName"],
	})
}
