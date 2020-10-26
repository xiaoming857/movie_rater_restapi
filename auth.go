package main

import (
	"log"
	"regexp"
	"time"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/gofiber/fiber/v2"
	jwtware "github.com/gofiber/jwt/v2"
	"golang.org/x/crypto/bcrypt"
)

var secretKey = []byte("Lorem ipsum dolor sit amet, consectetur adipiscing elit, sed do eiusmod tempor incididunt ut labore et dolore magna aliqua.")

// RegisterData struct
type RegisterData struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

// LoginData struct
type LoginData struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// Encrypts password
func hashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 11)
	return string(bytes), err
}

// Compares input pasword with encrypted passord
func compareHashAndPassword(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

// GenerateJWT generates Token
func generateJWT(id int, username string, email string) (string, error) {
	token := jwt.New(jwt.SigningMethodHS256)
	claims := token.Claims.(jwt.MapClaims)

	claims["id"] = id
	claims["username"] = username
	claims["email"] = email
	claims["exp"] = time.Now().Add(time.Hour + 24).Unix() // A day

	tokenString, err := token.SignedString(secretKey)

	if err != nil {
		return "", err
	}

	return tokenString, nil
}

// Protected protects routes
func Protected() func(*fiber.Ctx) error {
	return jwtware.New(jwtware.Config{
		SigningKey: secretKey,
		ErrorHandler: func(ctx *fiber.Ctx, err error) error {
			ctx.Status(401)
			return ctx.JSON(map[string]string{
				"error": "Unauthorized",
			})
		},
	})
}

// Login logins to the app with email and password
func Login(ctx *fiber.Ctx) error {
	loginData := new(LoginData)

	// Parse body (input data)
	if err := ctx.BodyParser(loginData); err != nil {
		log.Println(err.Error())
		return ctx.Status(400).JSON(map[string]string{
			"error": "Cannot parse JSON",
		})
	}

	// Validate email & password
	if len(loginData.Email) == 0 || len(loginData.Password) == 0 {
		return ctx.Status(400).JSON(map[string]string{
			"error": "Empty email or password",
		})
	}

	// Get id, email, password from database
	result, err := db.Query("SELECT id, username, email, password FROM users WHERE email = ?", loginData.Email)
	if err != nil {
		log.Println(err.Error())
		return ctx.Status(500).JSON(map[string]string{
			"error": "Internal server error",
		})
	}

	// Process data from database
	isUserExist := false
	var userID int
	var username string
	var userEmail string
	var userPassword string
	for result.Next() {
		if err = result.Scan(&userID, &username, &userEmail, &userPassword); err != nil {
			log.Println(err)
			return ctx.Status(500).JSON(map[string]string{
				"error": "Internal server error",
			})
		}

		isUserExist = true
	}

	// Authenticate email and password
	if !isUserExist || !compareHashAndPassword(loginData.Password, userPassword) {
		return ctx.Status(401).JSON(map[string]string{
			"error": "Bad credentials",
		})
	}

	// Create JWT token
	tokenString, err := generateJWT(userID, username, userEmail)
	if err != nil {
		log.Println(err.Error())
		return ctx.SendStatus(500)
	}

	return ctx.Status(200).JSON(map[string]string{
		"token": tokenString,
	})
}

// Register registers new user to the app with email and password
func Register(ctx *fiber.Ctx) error {
	registerData := new(RegisterData)

	// Parse body (input data)
	if err := ctx.BodyParser(registerData); err != nil {
		return ctx.Status(400).JSON(map[string]string{
			"error": "Cannot parse JSON",
		})
	}

	// Validate email
	if len(registerData.Email) == 0 {
		return ctx.Status(400).JSON(map[string]string{
			"error": "Empty email",
		})
	} else if regex, _ := regexp.Compile("^[\\w-\\.]+@([\\w-]+\\.)+[\\w-]{2,4}$"); !regex.MatchString(registerData.Email) {
		return ctx.Status(400).JSON(map[string]string{
			"error": "Invalid email format",
		})
	}

	// Validate password
	if len(registerData.Password) < 8 {
		return ctx.Status(400).JSON(map[string]string{
			"error": "Password too short (must be at least 8 characters)",
		})
	}

	// Validate username
	if len(registerData.Username) < 3 {
		return ctx.Status(400).JSON(map[string]string{
			"error": "Username too short (must be at least 3 characters)",
		})
	}

	// Check if username exist
	result, err := db.Query("SELECT id FROM users WHERE username = ?", registerData.Username)
	if result.Next() {
		log.Println(err.Error())
		return ctx.Status(409).JSON(map[string]string{
			"error": "Username already exist",
		})
	}

	// Check if email has been used
	result, err = db.Query("SELECT id FROM users WHERE email = ?", registerData.Email)
	if result.Next() {
		log.Println(err.Error())
		return ctx.Status(409).JSON(map[string]string{
			"error": "Email has been used",
		})
	}

	// Encrypt password
	hashedPassword, err := hashPassword(registerData.Password)
	if err != nil {
		log.Println(err.Error())
		return ctx.Status(500).JSON(map[string]string{
			"error": "Internal Server Error",
		})
	}

	// Create user in database
	_, err = db.Exec("INSERT INTO users (username, email, password) values (?,?,?)", registerData.Username, registerData.Email, hashedPassword)
	if err != nil {
		log.Println(err.Error())
		return ctx.Status(500).JSON(map[string]string{
			"error": "Internal Sever Error",
		})
	}

	// Get user ID
	result, err = db.Query("SELECT id FROM users WHERE email = ?", registerData.Email)
	if err != nil {
		log.Println(err.Error())
		return ctx.Status(500).JSON(map[string]string{
			"error": "Internal Sever Error",
		})
	}
	var userID int
	result.Next()
	err = result.Scan(&userID)
	if err != nil {
		log.Println(err.Error())
		return ctx.Status(500).JSON(map[string]string{
			"error": "Internal Sever Error",
		})
	}

	// Create JWT token
	tokenString, err := generateJWT(userID, registerData.Username, registerData.Email)
	if err != nil {
		log.Println(err.Error())
		return ctx.SendStatus(500)
	}

	return ctx.Status(200).JSON(map[string]string{
		"token": tokenString,
	})
}
