package main

import (
	"fmt"
	"os"

	"github.com/gin-gonic/gin"
	"harrisonwaffel/ccg-gmsa-dll/pkg"
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

	dirName := os.Getenv("ACTIVE_DIRECTORY")

	errChan := make(chan error)
	port, err := server.StartServer(errChan, dirName)
	if err != nil {
		panic(fmt.Sprintf("failed to start http server: %v", err))
	}

	err = pkg.CreateDir(dirName)
	if err != nil {
		panic(fmt.Sprintf("failed to create dynamic directory: %v", err))
	}

	err = pkg.WritePortFile(dirName, port)
	if err != nil {
		panic(fmt.Sprintf("failed to create dynamic directory: %v", err))
	}

	pkg.WriteClientCerts(dirName)
	// block on http server error
	select {
	case err = <-errChan:
		panic(err)
	}
}
