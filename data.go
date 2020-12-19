package main

import (
	"log"
	"math"

	"github.com/dgrijalva/jwt-go"
	"github.com/gofiber/fiber/v2"
)

// Movie struct
type Movie struct {
	ID        int
	Title     string
	AvgRating float64
}

// Review struct
type Review struct {
	ID       int
	Rating   int
	Comment  string
	Username string
}

// NewMovie struct
type NewMovie struct {
	Title string `json:"title"`
}

// NewReview struct
type NewReview struct {
	Rating  int    `json:"rating"`
	Comment string `json:"comment"`
}

// Home shows message
func Home(ctx *fiber.Ctx) error {
	ctx.SendString("Welcome to movie_rater api")
	return nil
}

// GetMovies gets all movie data from database
func GetMovies(ctx *fiber.Ctx) error {
	result, err := db.Query("SELECT id, title, avgRating FROM movies;")
	if err != nil {
		log.Println(err.Error())
		return ctx.Status(500).JSON(fiber.Map{
			"error": "Internal server error",
		})
	}

	var movies []Movie
	for result.Next() {
		var movie Movie
		err = result.Scan(&movie.ID, &movie.Title, &movie.AvgRating)
		if err != nil {
			log.Println(err)
			return ctx.Status(500).JSON(map[string]string{
				"error": "Internal server error",
			})
		}

		movies = append(movies, movie)
	}

	return ctx.Status(200).JSON(movies)
}

// GetReviews gets all review data from database
func GetReviews(ctx *fiber.Ctx) error {
	movieID := ctx.Params("id")
	result, err := db.Query("SELECT reviews.id, rating, comment, username FROM reviews INNER JOIN users ON userId = users.id WHERE movieId = ?;", movieID)
	if err != nil {
		log.Println(err.Error())
		return ctx.Status(500).JSON(map[string]string{
			"error": "Internal server error",
		})
	}

	var reviews []Review
	for result.Next() {
		var review Review
		err = result.Scan(&review.ID, &review.Rating, &review.Comment, &review.Username)

		if err != nil {
			log.Println(err.Error())
			return ctx.Status(500).JSON(map[string]string{
				"error": "Internal server error",
			})
		}

		reviews = append(reviews, review)
	}

	return ctx.Status(200).JSON(reviews)
}

// AddMovie adds a new movie to database
func AddMovie(ctx *fiber.Ctx) error {
	newMovie := new(NewMovie)
	if err := ctx.BodyParser(newMovie); err != nil {
		log.Println(err.Error())
		return ctx.Status(500).JSON(map[string]string{
			"error": "Internal server error",
		})
	}

	// Validate movie title
	if len(newMovie.Title) < 3 {
		return ctx.Status(400).JSON(map[string]string{
			"error": "Movie title too short (must be at least 3 characters)",
		})
	}

	// Inserts new movie to database
	_, err := db.Exec("INSERT INTO movies (title) values (?);", newMovie.Title)
	if err != nil {
		log.Println(err.Error())
		return ctx.Status(500).JSON(map[string]string{
			"error": "Internal server error",
		})
	}
	return ctx.Status(201).JSON(map[string]string{
		"success": "Movie successfully inserted",
	})
}

// AddReview adds a new review to database
func AddReview(ctx *fiber.Ctx) error {
	user := ctx.Locals("user").(*jwt.Token)
	claims := user.Claims.(jwt.MapClaims)
	userID := claims["id"].(float64)
	movieID := ctx.Params("id")
	newReview := new(NewReview)
	if err := ctx.BodyParser(newReview); err != nil {
		log.Println(err.Error())
		return ctx.Status(400).JSON(err.Error())
	}

	// Validate rating
	if newReview.Rating < 0.0 || newReview.Rating > 5.0 {
		return ctx.Status(400).JSON(map[string]string{
			"error": "Rating out of range (should be 0 - 5)",
		})
	}

	// Validate comment
	if len(newReview.Comment) > 500 {
		return ctx.Status(400).JSON(map[string]string{
			"error": "Comment exceeded limit (500 characters)",
		})
	}

	// Get movie from the database
	result, err := db.Query("SELECT avgRating, raterNum FROM movies WHERE id = ?", movieID)
	if err != nil {
		log.Println(err.Error())
		return ctx.Status(500).JSON(map[string]string{
			"error": "Internal server error",
		})
	}

	var avgRating float64
	var raterNum int
	isMovieExist := result.Next() // Returns false if no data within result (which means movie is not found)
	// Check movie existance
	if !isMovieExist {
		return ctx.Status(500).JSON(map[string]string{
			"error": "Movie does not exist",
		})
	}

	if err = result.Scan(&avgRating, &raterNum); err != nil {
		log.Println(err.Error())
		return ctx.Status(500).JSON(map[string]string{
			"error": "Internal server error",
		})
	}

	// Calculate the new average rating
	newAvgRating := (avgRating*float64(raterNum) + float64(newReview.Rating)) / float64(raterNum+1)

	// Update the avgRating and raterNum in the database
	_, err = db.Exec("UPDATE movies SET avgRating = ?, raterNum = ? WHERE id = ?", math.Round(newAvgRating*10)/10, raterNum+1, movieID)
	if err != nil {
		log.Println(err.Error())
		return ctx.Status(500).JSON(map[string]string{
			"error": "Internal server error",
		})
	}

	// Insert new review to database
	_, err = db.Exec("INSERT INTO reviews (rating, comment, movieId, userId) VALUES (?, ?, ?, ?);", newReview.Rating, newReview.Comment, movieID, int(userID))
	if err != nil {
		log.Println(err.Error())
		return ctx.Status(500).JSON(map[string]string{
			"error": "Internal server error",
		})
	}

	return ctx.Status(201).JSON(map[string]string{
		"success": "Review successfully inserted",
	})
}
