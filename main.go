package main

import (
	"io/ioutil"
	"fmt"
	"github.com/gin-gonic/gin"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
)

var router = gin.Default()

func main() {

	router.OPTIONS("/session", func(c *gin.Context) {
		var session Session
		c.Bind(&session)
		db.Create(&session)
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Headers", "*")
		c.Header("Access-Control-Allow-Methods", "*")
		c.Status(200)
		return
	})

	router.GET("/raw/sessions", func(c *gin.Context) {
		var sessions []Session
		db.Preload("Datapoints").Preload("Beacons").Find(&sessions)
		c.Header("Access-Control-Allow-Origin", "*")
		c.IndentedJSON(200, &sessions)
		return

	})

	router.GET("/raw/session/:id", func(c *gin.Context) {
		var session Session
		sessionid := c.Param("id")
		db.Preload("Datapoints").Preload("Beacons").Find(&session, sessionid)
		c.Header("Access-Control-Allow-Origin", "*")
		c.IndentedJSON(200, &session)
		return

	})

	router.POST("/raw/sessions", func(c *gin.Context) {
		finished := c.PostForm("Finished")
		var sessions []Session
		db.Preload("Datapoints").Preload("Beacons").Where("finished = ?", finished).Find(&sessions)
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

	router.POST("/beacon", func(c *gin.Context) {
		var beacon Beacon
		c.Bind(&beacon)
		db.Create(&beacon)
		c.Header("Access-Control-Allow-Origin", "*")
		c.Status(200)
		return
	})

	router.GET("/beacons", func(c *gin.Context) {
		var beacons []Beacon
		db.Find(&beacons)
		c.Header("Access-Control-Allow-Origin", "*")
		c.IndentedJSON(200, &beacons)
		return
	})

	router.POST("/sessionbeacon", func(c *gin.Context) {
		var sessionbeacon SessionBeacon
		c.Bind(&sessionbeacon)
		db.Create(&sessionbeacon)
		c.Header("Access-Control-Allow-Origin", "*")
		c.Status(200)
		return
	})

	router.POST("/map", func(c *gin.Context) {
		file, err := c.FormFile("Map")
		if err != nil {
			c.String(400, fmt.Sprintf("get form err: %s", err.Error()))
			return
		}

		files,_ := ioutil.ReadDir("./maps/")
		filecount := len(files)
		if err := c.SaveUploadedFile(file, "./maps/"+fmt.Sprint(filecount)+".png"); err != nil {
			c.String(400, fmt.Sprintf("upload file err: %s", err.Error()))
			return
		}
		var url URL
		url.Url = "firetracker.freheims.xyz:8000/maps/"+fmt.Sprint(filecount)+".png"

		c.IndentedJSON(200, url)
		return
	})

	router.GET("/debug/drop", func(c *gin.Context) {
		db.DropTableIfExists(&Session{})
		db.DropTableIfExists(&Datapoint{})
		db.DropTableIfExists(&Beacon{})
		db.DropTableIfExists(&SessionBeacon{})
		db.AutoMigrate(&Session{})
		db.AutoMigrate(&Datapoint{})
		db.AutoMigrate(&Beacon{})
		db.AutoMigrate(&SessionBeacon{})
	})

	router.GET("/debug/drop/sessions", func(c *gin.Context) {
		db.DropTableIfExists(&Session{})
		db.DropTableIfExists(&Datapoint{})
		db.DropTableIfExists(&SessionBeacon{})
		db.AutoMigrate(&Session{})
		db.AutoMigrate(&Datapoint{})
		db.AutoMigrate(&SessionBeacon{})
	})

	router.Static("/maps", "./maps")

	router.Run(":8000")
}
