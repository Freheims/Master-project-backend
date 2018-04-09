package main

import (
	"io/ioutil"
	"fmt"
	"math"
	"strings"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
)

var router = gin.Default()

func main() {

	router.Use(cors.New(cors.Config{
		AllowAllOrigins: true,
		AllowMethods:     []string{"PUT", "POST", "GET", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Authorization", "Content-Type", "Accept", "Content-Length"},
		ExposeHeaders:    []string{"Content-Length"},
	}))

	router.OPTIONS("/session", func(c *gin.Context) {
		var session Session
		c.Bind(&session)
		db.Create(&session)
		//c.Header("Access-Control-Allow-Origin", "*")
		//c.Header("Access-Control-Allow-Headers", "*")
		//c.Header("Access-Control-Allow-Methods", "*")
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
		db.Preload("Datapoints").Preload("Beacons").Preload("Locations").Where("finished = ?", finished).Find(&sessions)
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
		locations := ProcessSession(newSession)
		newSession.Locations = locations
		db.Model(&session).Updates(&newSession)
		db.Save(&session)
		c.Header("Access-Control-Allow-Origin", "*")
		c.Status(200)
		return

	})

	router.GET("/session/:id", func(c *gin.Context) {
		var session Session
		sessionid := c.Param("id")
		db.Preload("Datapoints").Preload("Beacons").Preload("Locations").Find(&session, sessionid)
		c.Header("Access-Control-Allow-Origin", "*")
		c.IndentedJSON(200, &session)
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
		db.DropTableIfExists(&Location{})
		db.AutoMigrate(&Session{})
		db.AutoMigrate(&Datapoint{})
		db.AutoMigrate(&Beacon{})
		db.AutoMigrate(&SessionBeacon{})
		db.AutoMigrate(&Location{})
	})

	router.GET("/debug/drop/sessions", func(c *gin.Context) {
		db.DropTableIfExists(&Session{})
		db.DropTableIfExists(&Datapoint{})
		db.DropTableIfExists(&SessionBeacon{})
		db.DropTableIfExists(&Location{})
		db.AutoMigrate(&Session{})
		db.AutoMigrate(&Datapoint{})
		db.AutoMigrate(&SessionBeacon{})
		db.AutoMigrate(&Location{})
	})

	router.Static("/maps", "./maps")

	router.Run(":8000")
}

func ProcessSession(session Session) []Location {
	var locations []Location
	datapoints := session.Datapoints
	prevDatapoint := datapoints[0]
	var location Location
	location.XCoordinate, location.YCoordinate = findCoordinates(prevDatapoint, session)
	location.Duration = 0
	for i := 1; i < len(datapoints); i++ {
		datapoint := datapoints[i]
		if isDatapointValid(datapoint, session) {
			if strings.ToLower(datapoint.UUID) == strings.ToLower(prevDatapoint.UUID) && datapoint.Major == prevDatapoint.Major && datapoint.Minor == prevDatapoint.Minor {
				location.Duration += (datapoint.Timestamp - prevDatapoint.Timestamp)
				if (datapoint.Steps - prevDatapoint.Steps) > 5 {
					location.Walking = true
				}
				if math.Abs(datapoint.RotationX - prevDatapoint.RotationX) > 1 || math.Abs(datapoint.RotationY - prevDatapoint.RotationY) > 1 || math.Abs(datapoint.RotationZ - prevDatapoint.RotationZ) > 1 {
					location.HeadMovement = true
				}
			prevDatapoint = datapoint
			} else if datapoint.RSSI > prevDatapoint.RSSI {
				locations = append(locations, location)
				location = Location{}
				location.XCoordinate, location.YCoordinate = findCoordinates(prevDatapoint, session)
				location.Duration = 0
				prevDatapoint = datapoint
			}
		}
	}
	locations = append(locations, location)
	return locations
}


func findCoordinates(datapoint Datapoint, session Session) (float64, float64) {
	for _, beacon := range session.Beacons {
		if strings.ToLower(beacon.UUID) == strings.ToLower(datapoint.UUID) && beacon.Major == datapoint.Major && beacon.Minor == datapoint.Minor {
			return beacon.XCoordinate, beacon.YCoordinate
		}
	}
	return 1000,1000
}

func isDatapointValid(datapoint Datapoint, session Session) bool {
	for _, beacon := range session.Beacons {
		if strings.ToLower(beacon.UUID) == strings.ToLower(datapoint.UUID) && beacon.Major == datapoint.Major && beacon.Minor == datapoint.Minor {
			return true
		}
	}
	return false
}
