package main

import "github.com/gin-gonic/gin"
import "github.com/pact-foundation/pact-go/examples/gin/provider"

func main() {
	router := gin.Default()
	router.POST("/users/login/:id", provider.UserLogin)
	router.Run(":8080")
}
