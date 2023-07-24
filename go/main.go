package main

import (
	"fmt"
	"os"

	"github.com/gin-gonic/gin"
	"harrisonwaffel/ccg-gmsa-dll/pkg"
)

// todo; Finalizer to clean up the tmp dirs

func main() {
	engine := gin.Default()
	server := pkg.HttpServer{
		Engine:      engine,
		Credentials: pkg.NewController("cattle-windows-gmsa-system"),
	}

	port := server.StartServer()
	fmt.Println(port)

	err := pkg.CreateDir(os.Getenv("ACTIVE_DIRECTORY"))
	if err != nil {
		panic(err)
	}

	select {}
}
