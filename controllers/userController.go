package controllers

import (
	"net/http"
	"net/smtp"
	"os"
	"time"
	"user-jwt/initializers"
	"user-jwt/models"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

func SignUp(c *gin.Context) {
	var body struct {
		Email    string
		Password string
	}

	if err := c.ShouldBindJSON(&body); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(body.Password), bcrypt.DefaultCost)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	user := models.User{Email: body.Email, Password: string(hash)}
	result := initializers.DB.Create(&user)
	if result.Error != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"error": result.Error.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "User created successfully",
	})
}

func Login(c *gin.Context) {
	var body struct {
		Email    string
		Password string
	}

	if err := c.ShouldBindJSON(&body); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	var user models.User
	initializers.DB.First(&user, "email = ?", body.Email)
	if user.ID == 0 {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"error": "Invalid email",
		})
		return
	}

	err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(body.Password))
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"error": "Invalid password",
		})
		return
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": user.ID,
		"exp": time.Now().Add(time.Hour * 72).Unix(),
	})

	// Sign and get the complete encoded token as a string using the secret
	tokenString, err := token.SignedString([]byte(os.Getenv("JWT_SECRET")))

	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie("Authorization", tokenString, 3600*24*3, "", "", false, true)

	c.JSON(http.StatusOK, gin.H{})
}


func FetchUsers(c *gin.Context) {
	var users []models.User
	initializers.DB.Find(&users)
	c.JSON(http.StatusOK, gin.H{
		"data": users,
	})
}

func Validate(c *gin.Context) {
	user, _ := c.Get("user")

	c.JSON(http.StatusOK, gin.H{
		"data": user.(models.User).Email,
	})
}	


func DeleteUser(c* gin.Context) {
	user,_ := c.Get("user")

	if user.(models.User).ID == 0 {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"error": "User not found",
		})
		return
	}

	
	initializers.DB.Delete(&models.User{}, user.(models.User).ID)
	
	c.SetCookie("Authorization", "", -1, "/", "", false, true)
	c.JSON(http.StatusOK, gin.H{
		"data": "User deleted successfully",
	})	
}


func ForgetPassword(c *gin.Context) {
	var body struct {
		Email string
	}

	if err := c.ShouldBindJSON(&body); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	var user models.User
	initializers.DB.First(&user, "email = ?", body.Email)
	if user.ID == 0 {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"error": "Invalid email",
		})
		return
	}

	auth := smtp.PlainAuth("", "samsonvjulius@gmail.com", os.Getenv("EMAIL_PASSWORD"), "smtp.gmail.com")

	msg  := "Subject: Password Reset\n\n" +
		"Click the link to reset your password: http://localhost:8080/reset-password?email=" + user.Email
	err := smtp.SendMail("smtp.gmail.com:587", auth, "samsonvjulius@gmail.com", []string{user.Email}, []byte(msg))
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Password reset email sent",
	})
}


func ResetPassword(c *gin.Context) {
	var body struct {
		NewPassword string `json:"new_password"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}
	
	var user models.User

	initializers.DB.First(&user, "email = ?", c.Query("email"))
	if user.ID == 0 {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"error": "Invalid email",
		})
		return
	}
	
	hash, err := bcrypt.GenerateFromPassword([]byte(body.NewPassword), bcrypt.DefaultCost)

	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	user.Password = string(hash)

	res := initializers.DB.Save(&user)	

	if res.Error != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"error": res.Error.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Password reset successfully",
	})
}