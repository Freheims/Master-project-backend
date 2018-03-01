package main

import (
	"github.com/gin-gonic/gin"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
)

var router = gin.Default()

func main() {

	router.POST("/session", func(c *gin.Context) {
		var session Session
		c.Bind(&session)
		db.Create(&session)
		c.Status(200)
		return
	})

	router.GET("/debug/drop", func(c *gin.Context) {
		db.DropTableIfExists(&Session{})
		db.DropTableIfExists(&Datapoint{})
	})

	router.Run(":8000")
}
