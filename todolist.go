package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"go.uber.org/fx"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type TodoItemModel struct {
	ID        int    `json:"id" gorm:"primaryKey"`
	Title     string `json:"title"`
	Completed bool   `json:"completed"`
}

func main() {
	app := fx.New(Module)
	app.Run()
}

var Module = fx.Module("server",
	fx.Provide(NewDB),
	fx.Invoke(RegisterRoutes),
)

func NewDB() (*gorm.DB, error) {
	db, err := gorm.Open(sqlite.Open("todo.db"), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}
	db.AutoMigrate(&TodoItemModel{})
	return db, nil
}

func RegisterRoutes(lc fx.Lifecycle, db *gorm.DB) {
	r := gin.Default()

	r.GET("/todos", GetTodoItems(db))
	r.POST("/todos", CreateTodoItem(db))
	r.PUT("/todos/:id", UpdateTodoItem(db))
	r.DELETE("/todos/:id", DeleteTodoItem(db))

	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			// Perform any initialization tasks here
			log.Println("Application started successfully")
			return nil
		},
		OnStop: func(ctx context.Context) error {
			log.Println("Connection error")
			return db.Error
		},
	})

	r.Run(":8080")
}

func GetTodoItems(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var todos []TodoItemModel
		db.Find(&todos)
		c.JSON(http.StatusOK, todos)
	}
}

func CreateTodoItem(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var input TodoItemModel
		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		todolist := TodoItemModel{
			ID:        input.ID,
			Title:     input.Title,
			Completed: input.Completed,
		}

		if err := db.Create(&todolist).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create todo item"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"data": todolist})
	}
}

func UpdateTodoItem(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, err := strconv.Atoi(c.Param("id"))

		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		var dataItem TodoItemModel

		if err := c.ShouldBind(&dataItem); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		if err := db.Where("id = ?", id).Updates(&dataItem).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"data": true})
	}
}

func DeleteTodoItem(db *gorm.DB) gin.HandlerFunc {

	return func(c *gin.Context) {

		id, err := strconv.Atoi(c.Param("id"))

		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		var item TodoItemModel

		if err := c.ShouldBindJSON(&item); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		if err := db.Where("id = ?", id).Delete(&item).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Todo item deleted successfully"})
	}

}
