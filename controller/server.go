package controller

import "github.com/gin-gonic/gin"

func StartServer() {
	router := gin.Default()
	router.GET("/", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "Api is now working",
		})
	})
	// Include Demo Controller
	DemoController(router)
	router.Run()
}

func DemoController(router *gin.Engine) {
	panic("unimplemented")
}
