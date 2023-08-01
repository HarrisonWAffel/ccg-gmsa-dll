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

const (
	serverCrt = "%s/%s/container/ssl/server/tls.crt"
	serverKey = "%s/%s/container/ssl/server/tls.key"

	clientCrt = "%s/%s/container/ssl/client/tls.crt"
	clientKey = "%s/%s/container/ssl/client/tls.key"

	hostClientCrt = "%s/%s/ssl/client/tls.crt"
	hostClientKey = "%s/%s/ssl/client/tls.key"
)

func (h *HttpServer) StartServer(errChan chan error, dirName string) (string, error) {
	h.Engine.GET("/rancher-ccg-gmsa-provider", h.handle)

	// use a host allocated port
	ln, err := net.Listen("tcp", ":0")
	if err != nil {
		return "", fmt.Errorf("failed to create http listener for http server: %v", err)
	}

	go func() {
		err = http.ServeTLS(ln, h.Engine, fmt.Sprintf("%s/%s/container/ssl/server/tls.crt", baseDir, dirName), fmt.Sprintf("%s/%s/container/ssl/server/tls.key", baseDir, dirName))
		errChan <- fmt.Errorf("HTTP server encountered a fatal error: %v", err.Error())
	}()

	// let the server come up and
	// be assigned a port
	time.Sleep(250 * time.Millisecond)
	_, port, err := net.SplitHostPort(ln.Addr().String())
	if err != nil {
		return "", fmt.Errorf("failed to split host port from net listener: %v", err)
	}
	fmt.Println("Listening on port ", port)
	return port, nil
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
