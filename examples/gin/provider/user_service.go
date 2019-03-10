package provider

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/pact-foundation/pact-go/examples/types"
)

// Login object to be submitted via API POST.
type Login struct {
	User     string `form:"username" json:"username" binding:"required"`
	Password string `form:"password" json:"password" binding:"required"`
}

var userRepository = &types.UserRepository{
	Users: map[string]*types.User{
		"jmarie": &types.User{
			Name:     "Jean-Marie de La Beaujardière😀😍",
			Username: "jmarie",
			Password: "issilly",
			Type:     "admin",
			ID:       10,
		},
	},
}

// Crude time-bound "bearer" token
func getAuthToken() string {
	return time.Now().Format("2006-01-02")
}

// Simple authentication middleware
func isAuthenticated() gin.HandlerFunc {
	return func(c *gin.Context) {
		fmt.Println(c.GetHeader("Authorization"))

		if c.GetHeader("Authorization") == fmt.Sprintf("Bearer %s", getAuthToken()) {
			c.Next()
		} else {
			c.JSON(http.StatusUnauthorized, gin.H{"status": "unauthorized"})
		}
	}
}

// UserLogin is the login route.
func UserLogin(c *gin.Context) {
	c.Header("X-Api-Correlation-Id", "1234")

	var json Login
	if c.BindJSON(&json) == nil {
		user, err := userRepository.ByUsername(json.User)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"status": "file not found"})
		} else if user.Username != json.User || user.Password != json.Password {
			c.JSON(http.StatusUnauthorized, gin.H{"status": "unauthorized"})
		} else {
			c.Header("X-Auth-Token", getAuthToken())
			c.JSON(http.StatusOK, types.LoginResponse{User: user})
		}
	}
}

// GetUser fetches a user if authenticated and exists
func GetUser(c *gin.Context) {
	fmt.Println("GET USER!")
	c.Header("X-Api-Correlation-Id", "1234")

	id, _ := strconv.Atoi(c.Param("id"))
	user, err := userRepository.ByID(id)

	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"status": "file not found"})
	} else {
		c.JSON(http.StatusOK, user)
	}
}
