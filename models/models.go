 package models

import (
  "gorm.io/driver/sqlite"
  "gorm.io/gorm"
)

var DB *gorm.DB


func InitDB() {
  var err error
  DB, err = gorm.Open(sqlite.Open("photoapp.db"), &gorm.Config{})
  if err != nil {
    panic("Failed to connect to Database")
  }

  DB.AutoMigrate(&User{}, &Photo{})

}

type User struct {
  gorm.Model
  Username string `gorm:"unique"`
  Email string
  Photos []Photo
}

type Photo struct {
  gorm.Model
  UserID uint
  URL string
}

