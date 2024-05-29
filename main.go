package main

import (
  "net/http"
  "time"
  "fmt"

  "github.com/dgrijalva/jwt-go"
  "photo-app/models"
  "github.com/gin-gonic/gin"
)

var jwtSecret = []byte("1OY73Lez*DeNq3fvJ*CeN#^&yWB%@F6e")

func generateJWT(username string) (string, error) {
  token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
    "username": username,
    "exp": time.Now().Add(24 * time.Hour).Unix(),
  })
  return token.SignedString(jwtSecret)
}

func authenticate(c *gin.Context) {
  tokenString, err := c.Cookie("jwt")
  if err != nil {
    c.JSON(http.StatusUnauthorized, gin.H{"erro": "Unauthorized"})
    c.Abort()
    return
  }

 token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
        if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
            return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
        }
        return jwtSecret, nil
    })

    if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
        c.Set("username", claims["username"])
    } else {
        c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
        c.Abort()
  }
}

func registerUser(c *gin.Context) {
    var user models.User
    if err := c.ShouldBindJSON(&user); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    if _, err := models.FindUserByUsername(user.Username); err == nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Username already exists"})
        return
    }

    if err := models.CreateUser(&user); err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }

    token, err := generateJWT(user.Username)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
        return
    }

    c.SetCookie("jwt", token, 3600*24, "/", "", false, true)
    c.JSON(http.StatusOK, gin.H{"message": "User registered successfully"})
}

func uploadPhoto(c *gin.Context) {
    username := c.GetString("username")
    var user models.User
    if err := models.DB.Where("username = ?", username).First(&user).Error; err != nil {
        c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
        return
    }

    if len(user.Photos) >= 25 {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Maximum number of photos reached"})
        return
    }

    file, err := c.FormFile("photo")
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to upload photo"})
        return
    }

    // Save the file to disk (or upload to cloud storage)
    filePath := "uploads/" + file.Filename
    if err := c.SaveUploadedFile(file, filePath); err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save photo"})
        return
    }

    photo := models.Photo{UserID: user.ID, URL: filePath}
    models.DB.Create(&photo)

    c.JSON(http.StatusOK, gin.H{"message": "Photo uploaded successfully"})
}

func getUserPhotos(c *gin.Context) {
    username := c.GetString("username")
    var user models.User
    if err := models.DB.Preload("Photos").Where("username = ?", username).First(&user).Error; err != nil {
        c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
        return
    }

    remaining := 25 - len(user.Photos)
    c.JSON(http.StatusOK, gin.H{"photos": user.Photos, "remaining": remaining})
}

func main() {
    r := gin.Default()

    models.InitDB()

    // Serve static files from the "static" directory
    r.Static("/static", "./static")

    // Serve uploads directory
    r.Static("/uploads", "./uploads")

    r.POST("/register", registerUser)
    r.POST("/upload", authenticate, uploadPhoto)
    r.GET("/photos", authenticate, getUserPhotos)

    // Serve the index.html file
    r.GET("/", func(c *gin.Context) {
        c.File("./static/index.html")
    })

    r.Run(":8080")
}
