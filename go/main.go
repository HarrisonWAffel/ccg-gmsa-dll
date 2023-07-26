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
		panic(fmt.Sprintf("failed to setup wrangler controller :%v", err))
	}

	server := pkg.HttpServer{
		Engine:      gin.Default(),
		Credentials: controller,
	}

	err = pkg.CreateDir(os.Getenv("ACTIVE_DIRECTORY"), "")
	if err != nil {
		panic(fmt.Sprintf("failed to create dynamic directory : %v", err))
	}
	err = pkg.GenCerts(os.Getenv("ACTIVE_DIRECTORY"))
	if err != nil {
		panic(fmt.Errorf("%v : could not generate certificates", err))
	}

	errChan := make(chan error)
	port := server.StartServer(errChan, os.Getenv("ACTIVE_DIRECTORY"))
	err = pkg.CreateDir(os.Getenv("ACTIVE_DIRECTORY"), port)
	if err != nil {
		panic(fmt.Sprintf("failed to create dynamic directory : %v", err))
	}

	// block on http server error
	select {
	case err = <-errChan:
		panic(err)
	}
}
