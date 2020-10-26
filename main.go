package main

import (
	"database/sql"
	"log"

	_ "github.com/go-sql-driver/mysql"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
)

// Database instance
var db *sql.DB

// Database setting
const (
	dbUser     = "test"
	dbPassword = "test@1234"
	dbName     = "movie_rater"
)

// Connect funtion
func connect() error {
	var err error

	// Use DSN string to open
	db, err = sql.Open("mysql", dbUser+":"+dbPassword+"@/"+dbName)
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
	app.Get("/movies", Protected(), GetMovies)
	app.Get("/reviews/:id", Protected(), GetReviews)
	app.Post("/movie", Protected(), AddMovie)
	app.Post("/review/:id", Protected(), AddReview)
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

	log.Fatalln(app.Listen(":8080"))
	defer db.Close()
}
