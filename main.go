package main

import (
	"database/sql"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
	"log"
	"net/http"
)

var db *sql.DB

func init() {
	// Verbindung zur PostgreSQL-Datenbank herstellen
	var err error
	connStr := "postgres://bomb3000:vbn888@localhost:5432/hangman_game?sslmode=disable"
	db, err = sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal("Fehler bei der Verbindung zur DB:", err)
	}
}

type User struct {
	ID               int    `json:"id"`
	Name             string `json:"name"`
	Password         string `json:"password"`
	Points           int    `json:"points"`
	LongestWinStreak int    `json:"longest_win_streak"`
	Highscore        int    `json:"highscore"`
}

func getUsers(c *gin.Context) {
	rows, err := db.Query("SELECT id, name, password, points, longest_win_streak, highscore FROM \"user\"")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	var users []User
	for rows.Next() {
		var user User
		if err := rows.Scan(&user.ID, &user.Name, &user.Password, &user.Points, &user.LongestWinStreak, &user.Highscore); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		users = append(users, user)
	}
	c.JSON(http.StatusOK, users)
}

// API-Endpunkt, um einen neuen Benutzer zu erstellen
func createUser(c *gin.Context) {
	var newUser User
	if err := c.BindJSON(&newUser); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Benutzer in die DB einfügen
	query := `INSERT INTO "user" (name, password, points, longest_win_streak, highscore) 
			  VALUES ($1, $2, $3, $4, $5) RETURNING id`
	err := db.QueryRow(query, newUser.Name, newUser.Password, newUser.Points, newUser.LongestWinStreak, newUser.Highscore).Scan(&newUser.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, newUser)
}

// API-Endpunkt, um einen Benutzer nach Namen zu löschen
func deleteUser(c *gin.Context) {
	name := c.Param("name")

	// Löschen des Benutzers aus der DB
	_, err := db.Exec("DELETE FROM \"user\" WHERE name=$1", name)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "User deleted"})
}

func main() {
	// Gin-Router initialisieren
	r := gin.Default()

	// CORS-Middleware aktivieren
	r.Use(cors.Default())

	// API-Endpunkte
	r.GET("/users", getUsers)
	r.POST("/users", createUser)
	r.DELETE("/users/:name", deleteUser)

	// Server starten
	err := r.Run(":8080")
	if err != nil {
		log.Fatal("Fehler beim Starten des Servers:", err)
	}
}
