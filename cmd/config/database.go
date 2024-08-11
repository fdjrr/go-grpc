package config

import (
	"log"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func ConnectDatabase() *gorm.DB {
	db, err := gorm.Open(mysql.Open("root:@tcp(127.0.0.1:3306)/go_grpc"), &gorm.Config{})
	if err != nil {
		log.Fatalf("Error connecting to database: %v", err.Error())
	}

	return db
}
