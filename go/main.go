package main

import (
	"fmt"
	"os"

	"github.com/gin-gonic/gin"
	"harrisonwaffel/ccg-gmsa-dll/pkg"
)

// todo; Finalizer to clean up the tmp dirs

func main() {

	controller, err := pkg.NewController(os.Getenv("RELEASE_NAMESPACE"))
	if err != nil {
		panic(fmt.Sprintf("failed to setup wrangler controller :%v", err))
	}

	server := pkg.HttpServer{
		Engine:      gin.Default(),
		Credentials: controller,
	}

	err = pkg.CreateDir(os.Getenv("ACTIVE_DIRECTORY"))
	if err != nil {
		panic(fmt.Sprintf("failed to create dynamic directory : %v", err))
	}

	errChan := make(chan error)
	err = pkg.WritePortFile(server.StartServer(errChan))
	if err != nil {
		panic(err)
	}

	// block on http server error
	select {
	case err := <-errChan:
		panic(err)
	}
}
