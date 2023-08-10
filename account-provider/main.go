package main

import (
	"fmt"
	"os"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	"harrisonwaffel/ccg-gmsa-dll/pkg"
)

// todo; Finalizer to clean up the tmp dirs

func main() {
	setLogLevel()

	activeDirectoryName := os.Getenv("ACTIVE_DIRECTORY")

	controller, err := pkg.NewClient(activeDirectoryName)
	if err != nil {
		panic(fmt.Sprintf("failed to setup wrangler controller: %v", err))
	}

	server := pkg.HttpServer{
		Engine:      gin.Default(),
		Credentials: controller,
	}

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

func setLogLevel() {
	switch os.Getenv("LOG_LEVEL") {
	case "info":
		log.SetLevel(log.InfoLevel)
	case "debug":
		log.SetLevel(log.DebugLevel)
	case "trace":
		log.SetLevel(log.TraceLevel)
	default:
		log.SetLevel(log.InfoLevel)
	}
}
