package main

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"harrisonwaffel/ccg-gmsa-dll/pkg"
	"os"
)

// todo; Finalizer to clean up the tmp dirs

func main() {
	controller, err := pkg.NewClient(os.Getenv("RELEASE_NAMESPACE"))
	if err != nil {
		panic(fmt.Sprintf("failed to setup wrangler controller: %v", err))
	}

	server := pkg.HttpServer{
		Engine:      gin.Default(),
		Credentials: controller,
	}

	activeDirectoryName := os.Getenv("ACTIVE_DIRECTORY")

	err = pkg.CreateDir(activeDirectoryName)
	if err != nil {
		panic(fmt.Sprintf("failed to create dynamic directory: %v", err))
	}

	err = pkg.WriteCerts(activeDirectoryName)
	if err != nil {
		panic(fmt.Sprintf("failed to write mTLS certificates to host: %v", err))
	}

	errChan := make(chan error)
	port, err := server.StartServer(errChan, activeDirectoryName)
	if err != nil {
		panic(fmt.Sprintf("failed to start http server: %v", err))
	}

	err = pkg.WritePortFile(activeDirectoryName, port)
	if err != nil {
		panic(fmt.Sprintf("failed to create dynamic directory: %v", err))
	}

	// block on http server error
	select {
	case err = <-errChan:
		panic(err)
	}
}
