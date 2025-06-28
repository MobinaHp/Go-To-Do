package main

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type todo struct {
	ID        uint   `json:"id" gorm:"primaryKey;autoIncrement"`
	Item      string `json:"item"`
	Detail    string `json:"detail"`
	Completed bool   `json:"completed"`
}

var db *gorm.DB

func initDatabase() {
	dsn := "host=localhost user=postgres password=123456 dbname=todoapp port=5432 sslmode=disable"
	var err error
	db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		panic("Failed to connect to database!")
	}

	err = db.AutoMigrate(&todo{})
	if err != nil {
		panic("Failed to migrate database!")
	}
}

func getTodos(context *gin.Context) {
	var todos []todo
	db.Find(&todos)
	context.IndentedJSON(http.StatusOK, todos)
}

func AddTodos(context *gin.Context) {
	var newTodo todo
	if err := context.BindJSON(&newTodo); err != nil {
		context.IndentedJSON(http.StatusBadRequest, gin.H{"message": "Invalid input"})
		return
	}
	db.Create(&newTodo)
	context.IndentedJSON(http.StatusCreated, newTodo)
}

func getTodoById(id uint) (*todo, error) {
	var t todo
	result := db.First(&t, id)
	if result.Error != nil {
		return nil, errors.New("Todo not found")
	}
	return &t, nil
}

func getTodo(context *gin.Context) {
	idParam := context.Param("id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		context.IndentedJSON(http.StatusBadRequest, gin.H{"message": "Invalid ID"})
		return
	}
	todo, err := getTodoById(uint(id))
	if err != nil {
		context.IndentedJSON(http.StatusNotFound, gin.H{"message": "Todo not found"})
		return
	}
	context.IndentedJSON(http.StatusOK, todo)
}

func partialUpdateTodo(context *gin.Context) {
	idParam := context.Param("id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		context.IndentedJSON(http.StatusBadRequest, gin.H{"message": "Invalid ID"})
		return
	}

	var updatedTodo todo
	if err := context.BindJSON(&updatedTodo); err != nil {
		context.IndentedJSON(http.StatusBadRequest, gin.H{"message": "Invalid input"})
		return
	}

	todo, err := getTodoById(uint(id))
	if err != nil {
		context.IndentedJSON(http.StatusNotFound, gin.H{"message": "Todo not found"})
		return
	}

	if updatedTodo.Item != "" {
		todo.Item = updatedTodo.Item
	}
	if updatedTodo.Detail != "" {
		todo.Detail = updatedTodo.Detail
	}
	todo.Completed = updatedTodo.Completed

	db.Save(todo)
	context.IndentedJSON(http.StatusOK, todo)
}

func DeleteTodo(context *gin.Context) {
	idParam := context.Param("id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		context.IndentedJSON(http.StatusBadRequest, gin.H{"message": "Invalid ID"})
		return
	}

	var t todo
	result := db.Delete(&t, id)
	if result.RowsAffected == 0 {
		context.IndentedJSON(http.StatusNotFound, gin.H{"message": "Todo not found"})
		return
	}
	context.IndentedJSON(http.StatusOK, gin.H{"message": "Todo deleted successfully"})
}

func main() {
	initDatabase()

	router := gin.Default()
	router.GET("/todos", getTodos)
	router.GET("/todos/:id", getTodo)
	router.POST("/todos", AddTodos)
	router.PATCH("/todos/:id", partialUpdateTodo)
	router.DELETE("/todos/:id", DeleteTodo)

	router.Run("localhost:8080")
}
