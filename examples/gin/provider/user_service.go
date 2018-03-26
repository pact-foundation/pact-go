package provider

import (
	"net/http"

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
		"Jean-Marie de La Beaujardière😀😍": &types.User{
			Name:     "Jean-Marie de La Beaujardière😀😍",
			Username: "Jean-Marie de La Beaujardière😀😍",
			Password: "issilly",
			Type:     "admin",
		},
	},
}

// UserLogin is the login route.
func UserLogin(c *gin.Context) {
	var json Login
	if c.BindJSON(&json) == nil {
		user, err := userRepository.ByUsername(json.User)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"status": "file not found"})
		} else if user.Username != json.User || user.Password != json.Password {
			c.JSON(http.StatusUnauthorized, gin.H{"status": "unauthorized"})
		} else {
			c.JSON(http.StatusOK, types.LoginResponse{User: user})
		}
	}
}
