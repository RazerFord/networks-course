package main

import (
	"rest/src/controllers"

	"github.com/gin-gonic/gin"
)

func main() {
	router := gin.Default()
	router.POST("/product", controllers.PostProduct)
	router.GET("/product/:id", controllers.GetProduct)
	router.PUT("/product/:id", controllers.PutProduct)
	router.DELETE("/product/:id", controllers.DeleteProduct)
	router.GET("/products", controllers.GetProducts)

	router.Run("localhost:8080")
}
