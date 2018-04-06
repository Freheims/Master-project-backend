package main

import (
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
)

type Session struct {
	gorm.Model
	Name		string
	User		string
	StartTime	int
	EndTime		int
	Datapoints	[]Datapoint `gorm:"foreignkey:SessionId"`
	Beacons		[]SessionBeacon `gorm:"foreignkey:SessionId"`
	Finished	bool `gorm:"default:false"`
	Map			string
}

type Datapoint struct {
	gorm.Model
	SessionId	int
	UUID		string
	Major		int
	Minor		int
	Timestamp	int
	RSSI		int
	Steps		int
	RotationX	float64
	RotationY	float64
	RotationZ	float64
}

type Beacon struct {
	gorm.Model
	UUID		string
	Major		string
	Minor		string
	Name		string
}

type SessionBeacon struct {
	gorm.Model
	SessionId	int
	UUID		string
	Major		string
	Minor		string
	Name		string
	XCoordinate	float64
	YCoordinate float64
}

type URL struct {
	Url string
}

var db *gorm.DB

func init() {
	init_db, err := gorm.Open("sqlite3", "firetracker.db")
	if err != nil {
		panic("failed to connect database")
	}
	db = init_db
	db.Set("gorm:auto_preload", true)

	db.AutoMigrate(&Session{})
	db.AutoMigrate(&Datapoint{})
	db.AutoMigrate(&Beacon{})
	db.AutoMigrate(&SessionBeacon{})
}
