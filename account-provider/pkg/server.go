package pkg

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type HttpServer struct {
	Engine      *gin.Engine
	Credentials *CredentialClient
}

func (h *HttpServer) StartServer(errChan chan error, dirName string) (string, error) {
	h.Engine.GET("/provider", h.handle)

	// use a host allocated port
	ln, err := net.Listen("tcp", ":0")
	if err != nil {
		return "", fmt.Errorf("failed to create http listener for http server: %v", err)
	}

	go func() {

		pool := x509.NewCertPool()
		clientCa, err := os.ReadFile(fmt.Sprintf(containerClientCa, baseDir, dirName))
		pool.AppendCertsFromPEM(clientCa)

		aca, err := os.ReadFile(fmt.Sprintf(serverCa, baseDir, dirName))
		pool.AppendCertsFromPEM(aca)

		rca, err := os.ReadFile(fmt.Sprintf(containerRootCa, baseDir, dirName))
		pool.AppendCertsFromPEM(rca)

		s := http.Server{
			Handler: h.Engine,
			TLSConfig: &tls.Config{
				InsecureSkipVerify: true,
				ClientCAs:          pool,
				VerifyConnection: func(state tls.ConnectionState) error {
					fmt.Println("verify conn")
					fmt.Println(state)
					j, _ := json.MarshalIndent(state, "", " ")
					fmt.Println(string(j))
					return nil
				},
				VerifyPeerCertificate: func(rawCerts [][]byte, verifiedChains [][]*x509.Certificate) error {
					fmt.Println("verify peer")
					return nil
				},
				ClientAuth: tls.RequireAndVerifyClientCert,
				MinVersion: tls.VersionTLS12,
			},
		}

		err = s.ServeTLS(ln, fmt.Sprintf(serverCrt, baseDir, dirName), fmt.Sprintf(serverKey, baseDir, dirName))
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
		Username:   string(s.Data["username"]),
		Password:   string(s.Data["password"]),
		DomainName: string(s.Data["domainName"]),
	})
}
