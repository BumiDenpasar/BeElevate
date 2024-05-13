package main

import (
	"database/sql"
	"log"
	"net/http"
	"strconv"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
)

type Lesson struct {
	UserID      string `json:"userid"`
	ID          int    `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Point       int    `json:"point"`
}

func main() {
	db, err := sql.Open("mysql", "root@tcp(127.0.0.1:3306)/be_elevate")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	err = db.Ping()
	if err != nil {
		log.Fatal(err)
	}

	r := gin.Default()

	// Konfigurasi middleware CORS
	config := cors.DefaultConfig()
	config.AllowOrigins = []string{"http://localhost:3000"}
	config.AllowMethods = []string{"GET", "POST", "PUT", "DELETE"}
	r.Use(cors.New(config))

	// Get all lessons
	r.GET("/api/lessons/:userid", func(c *gin.Context) {
		userid := c.Param("userid")
		rows, err := db.Query("SELECT userid, id, name, description, point FROM lesson WHERE userid = ?", userid)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error querying database"})
			return
		}
		defer rows.Close()

		var lessons []Lesson

		for rows.Next() {
			var lesson Lesson
			err := rows.Scan(&lesson.UserID, &lesson.ID, &lesson.Name, &lesson.Description, &lesson.Point)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Error scanning rows"})
				return
			}
			lessons = append(lessons, lesson)
		}

		c.JSON(http.StatusOK, lessons)
	})

	// Get single lesson
	r.GET("/api/lessons/single/:id", func(c *gin.Context) {
		id := c.Param("id")
		lessonID, err := strconv.Atoi(id)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid lesson ID"})
			return
		}

		var lesson Lesson
		err = db.QueryRow("SELECT userid, id, name, description, point FROM lesson WHERE id = ?", lessonID).Scan(&lesson.UserID, &lesson.ID, &lesson.Name, &lesson.Description, &lesson.Point)
		if err != nil {
			if err == sql.ErrNoRows {
				c.JSON(http.StatusNotFound, gin.H{"error": "Lesson not found"})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error querying database"})
			return
		}

		c.JSON(http.StatusOK, lesson)
	})

	// Create lesson
	r.POST("/api/lessons", func(c *gin.Context) {
		var lesson Lesson
		if err := c.ShouldBindJSON(&lesson); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request data"})
			return
		}

		result, err := db.Exec("INSERT INTO lesson (userid, name, description, point) VALUES (?, ?, ?, ?)",
			lesson.UserID, lesson.Name, lesson.Description, lesson.Point)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error inserting data into database"})
			return
		}

		lessonID, _ := result.LastInsertId()
		lesson.ID = int(lessonID)

		c.JSON(http.StatusCreated, lesson)
	})

	// Update lesson
	r.PUT("/api/lessons/:id", func(c *gin.Context) {
		id := c.Param("id")
		lessonID, err := strconv.Atoi(id)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid lesson ID"})
			return
		}

		var updatedLesson Lesson
		if err := c.ShouldBindJSON(&updatedLesson); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request data"})
			return
		}

		_, err = db.Exec("UPDATE lesson SET userid=?, name=?, description=?, point=? WHERE id=?",
			updatedLesson.UserID, updatedLesson.Name, updatedLesson.Description, updatedLesson.Point, lessonID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error updating data in the database"})
			return
		}

		c.JSON(http.StatusOK, updatedLesson)
	})

	// Delete lesson
	r.DELETE("/api/lessons/:id", func(c *gin.Context) {
		id := c.Param("id")
		lessonID, err := strconv.Atoi(id)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid lesson ID"})
			return
		}

		_, err = db.Exec("DELETE FROM lesson WHERE id=?", lessonID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error deleting data from database"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Lesson deleted successfully"})
	})

	r.Run(":8080")
}
