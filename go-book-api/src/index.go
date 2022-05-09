package main

import (
	"fmt"
	"strings"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt"
)

type Book struct {
	ID     string `json:"id"`
	Title  string `json:"title"`
	Author string `json:"author"`
}

var books = []Book{
	{ID: "1", Title: "Harry Potter", Author: "J. K. Rowling"},
	{ID: "2", Title: "The Lord of the Rings", Author: "J. R. R. Tolkien"},
	{ID: "3", Title: "The Wizard of Oz", Author: "L. Frank Baum"},
}

func loginHandler(c *gin.Context) {

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, &jwt.StandardClaims{
		ExpiresAt: time.Now().Add(5 * time.Minute).Unix(),
	})

	ss, err := token.SignedString([]byte("MySignature"))
	
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error(),})
	}

	c.JSON(http.StatusOK, gin.H{"token": ss,})
}

func validateToken(token string) error {
	_, err := jwt.Parse(token, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}

		return []byte("MySignature"), nil
	})

	return err
}

func authorizationMiddleware(c *gin.Context) {
	s := c.Request.Header.Get("Authorization")

	token := strings.TrimPrefix(s, "Bearer ")

	if err := validateToken(token); err != nil {
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}
}

func home(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "Hello Worold"})
}

func listBooksHandler(c *gin.Context) {
	c.JSON(http.StatusOK, books)
}

func listByID(c *gin.Context) {
	id := c.Param("id")

	for _, a := range books {
		if a.ID == id{
			c.JSON(http.StatusOK, a)
			return
		}
	}
	c.JSON(http.StatusNotFound, gin.H{"message": "book not found!"})
}

func createBookHandler(c *gin.Context) {
	var book Book

	if err := c.ShouldBindJSON(&book); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error(),})
		return
	}

	books = append(books, book)
	c.JSON(http.StatusCreated, book)
}

func updateBookHandler(c *gin.Context) {
	id := c.Param("id")
	
	var book Book

	if err := c.ShouldBindJSON(&book); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error(),})
		return
	}

	for i, a := range books {
		if a.ID == id{
			books = append(books[:i], book)
			c.JSON(http.StatusAccepted, gin.H{"message": "book has bee updated!"})
			return
		}
	}
	c.JSON(http.StatusNotFound, gin.H{"massage": "book not found"})
}

func deleteBookHandler(c *gin.Context) {
	id := c.Param("id")

	for i, a := range books {
		if a.ID == id {
			books = append(books[:i], books[i+1:]... )
			c.JSON(http.StatusOK, gin.H{"message": "book has been deleted!"})
			return
		}
	}
	c.JSON(http.StatusNotFound, gin.H{"massage": "book not found"})
}

func main() {
	r := gin.Default()

	r.POST("/login", loginHandler)
	r.GET("/", home)

	protected := r.Group("/", authorizationMiddleware)

	protected.GET("/books", listBooksHandler)
	protected.GET("/books/:id", listByID)
	protected.POST("/books", createBookHandler)
	protected.PUT("/books/:id", updateBookHandler)
	protected.DELETE("/books/:id", deleteBookHandler)

	r.Run("localhost:8080")
}