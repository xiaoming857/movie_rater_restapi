package main

import (
	"database/sql"
	"log"
	"os"

	_ "github.com/go-sql-driver/mysql"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
)

// Database instance
var db *sql.DB

// Database setting
const (
	dbUser     = "ARV134"
	dbPassword = "123456"
	dbName     = "ARV134"
	dbProtocol = "tcp"
	dbAddress  = "34.101.252.33"
	dbPort     = "3306"
)

// Connect funtion
func connect() error {
	var err error

	// Use DSN string to open
	db, err = sql.Open("mysql", dbUser+":"+dbPassword+"@"+dbProtocol+"("+dbAddress+":"+dbPort+")/"+dbName)
	if err != nil {
		return err
	}

	if err = db.Ping(); err != nil {
		return err
	}

	return nil
}

// Routes function
func setupRoutes(app *fiber.App) {
	app.Use(logger.New())

	// Unrestricted routes
	app.Get("/", Home)
	app.Post("/login", Login)
	app.Post("/register", Register)

	// Restricted routes
	app.Use("/refresh", RefreshProtected())
	app.Get("/refresh", Refresh)

	app.Use(AccessProtected())
	app.Get("/movies", GetMovies)
	app.Get("/reviews/:id", GetReviews)
	app.Post("/movie", AddMovie)
	app.Post("/review/:id", AddReview)
}

func main() {
	// Connect with database
	if err := connect(); err != nil {
		panic(err.Error())
	}

	// Create a Fiber app
	app := fiber.New()

	// Routes
	setupRoutes(app)

	log.Fatalln(app.Listen(":" + os.Getenv("PORT")))
	defer db.Close()
}
