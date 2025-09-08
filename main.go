package main

import (
	"user-jwt/controllers"
	"user-jwt/initializers"
	"user-jwt/middleware"

	"github.com/gin-gonic/gin"
)

func init() {
	initializers.LoadEnvVariables()
	initializers.ConnectDB()
	initializers.SyncDatabase()
}

func main() {
	router := gin.Default()
	router.POST("/signup", controllers.SignUp)
	router.POST("/login", controllers.Login)
	router.GET("/users", middleware.RequireAuth, controllers.FetchUsers)
	router.GET("/validate", middleware.RequireAuth, controllers.Validate)
	router.DELETE("/user", middleware.RequireAuth, controllers.DeleteUser)
	router.POST("/forget-password", controllers.ForgetPassword)
	router.PATCH("/reset-password", controllers.ResetPassword)
			
	blogRoutes := router.Group("/blogs")
	blogRoutes.Use(middleware.RequireAuth)
	{
		blogRoutes.POST("/", controllers.CreateBlog)         
		blogRoutes.GET("/", controllers.GetMyBlogs) 
		blogRoutes.GET("/:id", controllers.GetBlog)
		blogRoutes.DELETE("/:id", controllers.DeleteBlog)         
	}
	router.Run()
}