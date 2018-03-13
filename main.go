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
		c.Header("Access-Control-Allow-Origin", "*")
		c.Status(200)
		return
	})

	router.GET("/raw/sessions", func(c *gin.Context) {
		var sessions []Session
		db.Preload("Datapoints").Find(&sessions)
		c.Header("Access-Control-Allow-Origin", "*")
		c.IndentedJSON(200, &sessions)
		return

	})

	router.GET("/raw/session/:id", func(c *gin.Context) {
		var session Session
		sessionid := c.Param("id")
		db.Preload("Datapoints").Find(&session, sessionid)
		c.Header("Access-Control-Allow-Origin", "*")
		c.IndentedJSON(200, &session)
		return

	})

	router.POST("/raw/sessions", func(c *gin.Context) {
		finished := c.PostForm("Finished")
		var sessions []Session
		db.Preload("Datapoints").Where("finished = ?", finished).Find(&sessions)
		c.Header("Access-Control-Allow-Origin", "*")
		c.IndentedJSON(200, &sessions)
		return

	})

	router.PUT("/raw/session/:id", func(c *gin.Context) {
		var session Session
		sessionid := c.Param("id")
		db.Find(&session, sessionid)
		var newSession Session
		c.Bind(&newSession)
		db.Model(&session).Updates(&newSession)
		db.Save(&session)
		c.Header("Access-Control-Allow-Origin", "*")
		c.Status(200)
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
