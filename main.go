package main

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type User struct {
	ID       uint   `json:"id" gorm:"primary_key"`
	Username string `json:"username"`
	Password string `json:"password"`
}

func (User) TableName() string {
	return "testing"
}

type RequiredUser struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

var DB *gorm.DB

const (
	db_host     = "localhost"
	db_port     = 5432
	db_user     = "tunggal"
	db_password = "laggnut"
	db_dbname   = "gaguna"
)

func ConnectDatabase() {
	dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable", db_host, db_port, db_user, db_password, db_dbname)
	database, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		panic("Failed to connect to database!")
	}

	database.AutoMigrate(&User{})

	DB = database
}

func CheckError(err error) {
	if err != nil {
		panic(err.Error())
	}
}

// GET - Find all users
func FindUsers(c *gin.Context) {
	var users []User
	DB.Find(&users)
	c.JSON(http.StatusOK, gin.H{"data": users})
}

// GET - Find 1 user
func FindUser(c *gin.Context) {
	var user User
	username := c.Query("username")
	DB.Where("username = ?", username).Find(&user)
	if user.ID == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"status": "username not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"Data": user})
}

// POST - Create new user
func CreateUser(c *gin.Context) {
	var user User
	err := json.NewDecoder(c.Request.Body).Decode(&user)
	CheckError(err)
	// Check if user exist
	DB.Where("username = ?", user.Username).Find(&user)
	if user.ID != 0 {
		c.JSON(http.StatusBadRequest, gin.H{"status": "username exist"})
		return
	}

	// Create new user
	DB.Select("username", "password").Create(&user)
	c.JSON(http.StatusOK, gin.H{"Status": "User " + user.Username + " sucessfully added"})
}

// PUT - Update user
func UpdateUser(c *gin.Context) {
	var user User
	username := c.Param("username")
	// Check if user exist
	DB.Where("username = ?", username).Find(&user)
	if user.ID == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"status": "username not found"})
		return
	}

	// Validate request body
	var updatedUser RequiredUser
	if err := c.ShouldBindJSON(&updatedUser); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Update the whole user
	fmt.Println("==", updatedUser, user)
	DB.Model(&user).Where("username = ?", username).Updates(User{Username: updatedUser.Username, Password: updatedUser.Password})
	c.JSON(http.StatusOK, gin.H{"status": "everything on user " + username + " changed"})
}

// PATCH - Change user's password
func ChangePassword(c *gin.Context) {
	var user User
	username := c.Param("username")
	// Check if user exist
	DB.Where("username = ?", username).Find(&user)
	if user.ID == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"status": "username not found"})
		return
	}

	// Update the password
	err := json.NewDecoder(c.Request.Body).Decode(&user)
	CheckError(err)
	password := user.Password
	DB.Model(&user).Where("username = ?", username).Update("password", password)
	c.JSON(http.StatusOK, gin.H{"status": "password of user " + username + " changed"})
}

// DELETE - Delete existing user
func DeleteUser(c *gin.Context) {
	var user User
	username := c.Param("username")
	// Check if user exist
	DB.Where("username = ?", username).Find(&user)
	if user.ID == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"status": "username not found"})
		return
	}

	// Delete user
	DB.Where("username = ?", username).Delete(&user)
	c.JSON(http.StatusOK, gin.H{"status": "user " + username + " deleted"})
}

func main() {
	// Connect to database
	ConnectDatabase()

	r := gin.Default()

	// Routes
	r.GET("/users", FindUsers)
	r.GET("/user", FindUser)
	r.POST("/createuser", CreateUser)
	r.PUT("/changeuser/:username", UpdateUser)
	r.PATCH("/changepassword/:username", ChangePassword)
	r.DELETE("/deleteuser/:username", DeleteUser)

	// Run the server
	r.Run()
}
