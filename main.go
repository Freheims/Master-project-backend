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

	router.GET("/raw/sessions", func(c *gin.Context) {
		var sessions []Session
		db.Preload("Datapoints").Find(&sessions)
		c.IndentedJSON(200, &sessions)
		return

	})

	router.GET("/raw/session/:id", func(c *gin.Context) {
		var session Session
		sessionid := c.PostForm("id")
		db.Preload("Datapoints").Find(&session, sessionid)
		c.IndentedJSON(200, &session)
		return

	})

	router.GET("/debug/drop", func(c *gin.Context) {
		db.DropTableIfExists(&Session{})
		db.DropTableIfExists(&Datapoint{})
		db.AutoMigrate(&Session{})
		db.AutoMigrate(&Datapoint{})
	})

	router.Run(":8000")
}
