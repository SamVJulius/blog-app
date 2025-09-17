package controllers

import (
	"net/http"
	"user-jwt/initializers"
	"user-jwt/models"

	"github.com/gin-gonic/gin"
)

func CreateBlog(c *gin.Context) {

	var blog struct {
		Title   string   `json:"title"`
		Content string   `json:"content"`
		Tags    []string `json:"tags"`
	}

	if err := c.ShouldBindJSON(&blog); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "User ID not found in context"})
		return
	}

	var tags []models.Tag

	for _, tagName := range blog.Tags {
		tag := models.Tag{Name: tagName}
		initializers.DB.FirstOrCreate(&tag, models.Tag{Name: tagName})
		tags = append(tags, tag)
	}

	post := models.Post{
        Title:   blog.Title,
        Body:    blog.Content,
        UserID:  user.(models.User).ID, // assuming JWT middleware sets this
        Tags:    tags,
    }

	if err := initializers.DB.Create(&post).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create blog post"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Blog post created successfully", "data": blog})
}

func GetMyBlogs(c *gin.Context) {
	user, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "User ID not found in context"})
		return
	}

	var blogs []models.Post
	if err := initializers.DB.Where("user_id = ?", user.(models.User).ID).Find(&blogs).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve blog posts"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": blogs})
}

func GetBlog(c *gin.Context) {
	user, exists := c.Get("user")

	if !exists {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "User not found in context"})
		return
	}

	var blog models.Post
	if err := initializers.DB.Where("id = ? AND user_id = ?", c.Param("id"), user.(models.User).ID).First(&blog).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Blog post not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": blog})
}

func DeleteBlog(c *gin.Context) {
	user, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "User not found in context"})
		return
	}

	initializers.DB.Delete(&models.Post{}, "id = ? AND user_id = ?", c.Param("id"), user.(models.User).ID)

	c.JSON(http.StatusOK, gin.H{"message": "Blog post deleted successfully"})
}

func FindPostsByTag(c *gin.Context) {
	tag := c.Param("tag")

	var post []models.Post
	result := initializers.DB.
		Joins("JOIN post_tags ON post_tags.post_id = posts.id").
		Joins("JOIN tags ON tags.id = post_tags.tag_id").
		Where("tags.name =?", tag).
		Preload("Tags").
		Find(&post)

	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": result.Error.Error()})
		return
	}

	c.JSON((http.StatusOK), gin.H{"data": post})
}
