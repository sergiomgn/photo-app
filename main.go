package main

import (
  "fmt"
  "net/http"
  "photo-app/models"
  "github.com/gin-gonic/gin"
  "gorm.io/gorm"
)

func registerUser(c *gin.Context) {
  var user models.User
  if err := c.ShouldBindJSON(&user); err != nil {
    c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
    return
  }
  if err := models.DB.Where("username = ?", user.Username).First(&user).Error; err == gorm.ErrRecordNotFound {
    models.DB.Create(&user)
    c.JSON(http.StatusOK, gin.H{"message": "User Registered Successfully"})
  } else {
    c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
  }
}

func uploadPhoto(c *gin.Context) {
  username := c.Param("username")
  var user models.User
  if err := models.DB.Where("username = ?", username).First(&user).Error; err != nil {
    c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
    return
  }

  if len(user.Photos) >= 25 {
    c.JSON(http.StatusBadRequest, gin.H{"error": "Maximum number of photos taken for this user"})
    return
  }

  file, err := c.FormFile("photo")
  if err != nil {
    c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to upload photo"})
  }

  // Save the file to disk (or upload to cloud storage)
  filePath := "uploads/" + file.Filename
  if err := c.SaveUploadedFile(file, filePath); err != nil {
    c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save the photo in disk"})
    return
  }

  photo := models.Photo{UserID: user.ID, URL: filePath}
  models.DB.Create(&photo)

  c.JSON(http.StatusOK, gin.H{"message": "Photo uploaded successfully"})
}

func getUserPhotos(c *gin.Context) {
  username := c.Param("username")
  var user models.User
  if err := models.DB.Preload("Photos").Where("username = ?", username).First(&user).Error; err != nil {
    c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
    return
  }

  remaining := 25-len(user.Photos)
  c.JSON(http.StatusOK, gin.H{"photos": user.Photos, "remaining": remaining})
}

func main() {
  r := gin.Default()
  models.InitDB()

  // Define routes
  r.POST("/register", registerUser)
  r.POST("/upload/:username", uploadPhoto)
  r.GET("/photos/:username", getUserPhotos)

  // Serve the index.html file
  r.GET("/", func(c *gin.Context){
    c.File("./static/index.html")
  })
  fmt.Println("Starting Server")
  // Start the server
  r.Run()
}
