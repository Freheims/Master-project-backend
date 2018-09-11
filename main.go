package main

import (
	"mime"
	"log"
	"io/ioutil"
	"io"
	"os"
	"fmt"
	"math"
	"strings"
	"github.com/gin-gonic/gin"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
)

func CORSMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {

        c.Writer.Header().Set("Content-Type", "application/json")
        c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "OPTIONS, POST")
        c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
        c.Next()
    }
}

func main() {
	f, _ := os.Create("gin.log")
    gin.DefaultWriter = io.MultiWriter(f, os.Stdout)
	log.SetOutput(f)

	var router = gin.Default()

	router.Use(CORSMiddleware())

	router.OPTIONS("/session", func(c *gin.Context) {
		var session Session
		c.Bind(&session)
		if session.Name =="" || session.User == "" || len(session.Beacons) <1 {
			return
		}
		db.Create(&session)
		return
	})

	router.GET("/sessions", func(c *gin.Context) {
		var sessions []Session
		db.Preload("Beacons").Find(&sessions)
		c.IndentedJSON(200, &sessions)
		return

	})

	router.POST("/sessions", func(c *gin.Context) {
		finished := c.PostForm("Finished")
		var sessions []Session
		db.Preload("Beacons").Where("finished = ?", finished).Find(&sessions)
		c.IndentedJSON(200, &sessions)
		return

	})

	router.GET("/session/:id", func(c *gin.Context) {
		var session Session
		sessionid := c.Param("id")
		db.Preload("Beacons").Preload("Locations").Find(&session, sessionid)
		c.IndentedJSON(200, &session)
		return

	})

	router.GET("/fullsession/:id", func(c *gin.Context) {
		var session Session
		sessionid := c.Param("id")
		db.Preload("Datapoints").Preload("Beacons").Preload("Locations").Find(&session, sessionid)
		c.IndentedJSON(200, &session)
		return

	})

	router.PUT("/session/:id", func(c *gin.Context) {
		var session Session
		sessionid := c.Param("id")
		db.Find(&session, sessionid)
		var newSession Session
		c.Bind(&newSession)
		locations := ProcessSession(newSession)
		newSession.Locations = locations
		db.Model(&session).Updates(&newSession)
		db.Save(&session)
		return

	})

	router.OPTIONS("/beacon", func(c *gin.Context) {
		var beacon Beacon
		c.Bind(&beacon)
		if beacon.UUID == "" || beacon.Major == "" || beacon.Minor == "" {
			return
		}
		db.Create(&beacon)
		c.Status(200)
		return
	})


	router.OPTIONS("/beacon/delete", func(c *gin.Context) {
		var beacon Beacon
		id := c.PostForm("Id")
		if id == "" {
			c.IndentedJSON(400, gin.H{"message": "No Id provided", "status": "failure"})
			return
		}

		if db.First(&beacon, id).RecordNotFound() {
			c.IndentedJSON(500, gin.H{"message": "Didn't find any beacons", "status": "failure"})
			return
		}
		db.Delete(&beacon)
		c.IndentedJSON(200, gin.H{"message": "Deleted", "status": "succsess"})
	})

	router.GET("/beacons", func(c *gin.Context) {
		var beacons []Beacon
		db.Find(&beacons)
		c.IndentedJSON(200, &beacons)
		return
	})

	router.POST("/sessionbeacon", func(c *gin.Context) {
		var sessionbeacon SessionBeacon
		c.Bind(&sessionbeacon)
		db.Create(&sessionbeacon)
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
		fileExtensions, err := mime.ExtensionsByType(file.Header.Get("Content-Type"))
		fileExtension := fileExtensions[0]

		if err := c.SaveUploadedFile(file, "./maps/"+fmt.Sprint(filecount) + fileExtension); err != nil {
			c.String(400, fmt.Sprintf("upload file err: %s", err.Error()))
			return
		}
		var url URL
		url.Url = "firetracker.freheims.xyz:8000/maps/"+fmt.Sprint(filecount) + fileExtension

		c.IndentedJSON(200, url)
		return
	})

	//router.GET("/debug/drop", func(c *gin.Context) {
	//	db.DropTableIfExists(&Session{})
	//	db.DropTableIfExists(&Datapoint{})
	//	db.DropTableIfExists(&Beacon{})
	//	db.DropTableIfExists(&SessionBeacon{})
	//	db.DropTableIfExists(&Location{})
	//	db.AutoMigrate(&Session{})
	//	db.AutoMigrate(&Datapoint{})
	//	db.AutoMigrate(&Beacon{})
	//	db.AutoMigrate(&SessionBeacon{})
	//	db.AutoMigrate(&Location{})
	//})

	//router.GET("/debug/drop/sessions", func(c *gin.Context) {
	//	db.DropTableIfExists(&Session{})
	//	db.DropTableIfExists(&Datapoint{})
	//	db.DropTableIfExists(&SessionBeacon{})
	//	db.DropTableIfExists(&Location{})
	//	db.AutoMigrate(&Session{})
	//	db.AutoMigrate(&Datapoint{})
	//	db.AutoMigrate(&SessionBeacon{})
	//	db.AutoMigrate(&Location{})
	//})

	router.Static("/maps", "./maps")

	router.Run(":8000")
}

func ProcessSession(session Session) []Location {
	var locations []Location
	datapoints := session.Datapoints
	numDatapoints := len(datapoints)
	log.Println("Number of datapoints: " + fmt.Sprint(numDatapoints))
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
				//location = Location{}
				//prevX, prevY := findCoordinates(prevDatapoint, session)
				//newX, newY := findCoordinates(datapoint, session)
				//location.XCoordinate, location.YCoordinate = findMidpoint(prevX, prevY, newX, newY)
				//locations = append(locations, location)
				location = Location{}
				location.XCoordinate, location.YCoordinate = findCoordinates(prevDatapoint, session)
				location.Duration = 0
				prevDatapoint = datapoint
			}
		}
		numDatapoints -= 1
	}
	locations = append(locations, location)
	cleanedLocations := CleanLocations2(locations)
	return cleanedLocations
	//fmt.Println(len(cleanedLocations))
	//for {
	//	if len(cleanedLocations) != len(locations) {
	//		locations = cleanedLocations
	//		cleanedLocations = CleanLocations(locations)
	//		fmt.Println(len(cleanedLocations))
	//	} else {
	//		return cleanedLocations
	//	}
	//}
	//return locations
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

func findMidpoint(x1 float64, y1 float64, x2 float64, y2 float64) (float64, float64) {
	x := (x1 + x2)/2
	y := (y1 + y2)/2
	return x, y
}

func CleanLocations2(locations []Location) []Location {
	var legitLocations []Location
	for _, location := range locations {
		if location.Duration > 10000 {
			legitLocations = append(legitLocations, location)
		}
	}
	var cleanLocations []Location
	for i := 0; i < len(legitLocations)-1; i++ {
		currentLocation := legitLocations[i]
		nextLocation := legitLocations[i+1]
		if currentLocation.XCoordinate == nextLocation.XCoordinate && currentLocation.YCoordinate == nextLocation.YCoordinate {
			fmt.Println(currentLocation)
			fmt.Println(nextLocation)
			fmt.Println("")
			newLocation := MergeLocations(currentLocation, nextLocation)
			cleanLocations = append(cleanLocations, newLocation)
			i++
		} else {
			cleanLocations = append(cleanLocations, currentLocation)
		}
	}
	if legitLocations[len(legitLocations)-1].XCoordinate != legitLocations[len(legitLocations)-2].XCoordinate && legitLocations[len(legitLocations)-1].YCoordinate != legitLocations[len(legitLocations)-2].YCoordinate {
		cleanLocations = append(cleanLocations, legitLocations[len(legitLocations)-1])
	}
	return cleanLocations
}


//func CleanLocations(locations []Location) []Location {
//	var cleanLocations []Location
//	//TODO Fix walking and head movement
//	for i := 0; i < len(locations)-2; i++ {
//		currentLocation := locations[i]
//		nextLocation := locations[i+1]
//		nextNextLocation := locations[i+2]
//		if currentLocation.XCoordinate == nextLocation.XCoordinate && currentLocation.YCoordinate == nextLocation.YCoordinate {
//			currentLocation.Duration += nextLocation.Duration
//			cleanLocations = append(cleanLocations, currentLocation)
//			i++
//		} else if currentLocation.XCoordinate == nextNextLocation.XCoordinate && currentLocation.YCoordinate == nextNextLocation.YCoordinate  && nextLocation.Duration < 1 {
//			currentLocation.Duration += nextNextLocation.Duration
//			cleanLocations = append(cleanLocations, currentLocation)
//			i += 2
//		} else {
//			cleanLocations = append(cleanLocations, currentLocation)
//		}
//	}
//	return cleanLocations
//
//}

func MergeLocations(firstLocation Location, secondLocation Location) Location{
	var newLocation Location
	newLocation.XCoordinate = firstLocation.XCoordinate
	newLocation.YCoordinate = firstLocation.YCoordinate
	newLocation.Duration = firstLocation.Duration + secondLocation.Duration
	if firstLocation.Walking || secondLocation.Walking {
		newLocation.Walking = true
	}
	if firstLocation.HeadMovement || secondLocation.HeadMovement {
		newLocation.HeadMovement = true
	}
	return newLocation
}
