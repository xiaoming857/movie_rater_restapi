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

// Secret key for access token
var secretKey = []byte("Lorem ipsum dolor sit amet, consectetur adipiscing elit, sed do eiusmod tempor incididunt ut labore et dolore magna aliqua.")

// Secret key for refresh token
var refreshKey = []byte("Quisque sagittis purus sit amet volutpat consequat. Duis at consectetur lorem donec massa. Diam vel quam elementum pulvinar etiam.")

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

//GenerateRefreshToken generates refresh token
func generateRefreshToken(id int, username string, email string) (string, error) {
	token := jwt.New(jwt.SigningMethodHS256)
	claims := token.Claims.(jwt.MapClaims)

	claims["id"] = id
	claims["username"] = username
	claims["email"] = email
	claims["exp"] = time.Now().Add(time.Hour * 24 * 7 * 4).Unix() // A Month

	tokenString, err := token.SignedString(refreshKey)

	if err != nil {
		return "", err
	}

	return tokenString, nil
}

// GenerateAccessToken generates Token
func generateAccessToken(id int, username string, email string) (string, error) {
	token := jwt.New(jwt.SigningMethodHS256)
	claims := token.Claims.(jwt.MapClaims)

	claims["id"] = id
	claims["username"] = username
	claims["email"] = email
	claims["exp"] = time.Now().Add(time.Minute * 5).Unix() // 5 Minutes

	tokenString, err := token.SignedString(secretKey)

	if err != nil {
		return "", err
	}

	return tokenString, nil
}

// RefreshProtected protects routes
func RefreshProtected() func(*fiber.Ctx) error {
	return jwtware.New(jwtware.Config{
		SigningKey: refreshKey,
		ErrorHandler: func(ctx *fiber.Ctx, err error) error {
			return ctx.Status(401).JSON(map[string]string{
				"error": "Unauthorized",
			})
		},
	})
}

// AccessProtected protects routes
func AccessProtected() func(*fiber.Ctx) error {
	return jwtware.New(jwtware.Config{
		SigningKey: secretKey,
		ErrorHandler: func(ctx *fiber.Ctx, err error) error {
			return ctx.Status(401).JSON(map[string]string{
				"error": "Unauthorized",
			})
		},
	})
}

// Refresh checks for authorization
func Refresh(ctx *fiber.Ctx) error {
	user := ctx.Locals("user").(*jwt.Token)
	claims := user.Claims.(jwt.MapClaims)
	userID := claims["id"].(float64)
	username := claims["username"].(string)
	userEmail := claims["email"].(string)

	accessTokenString, err := generateAccessToken(int(userID), username, userEmail)
	if err != nil {
		log.Println(err.Error())
		return ctx.SendStatus(500)
	}

	refreshTokenString, err := generateRefreshToken(int(userID), username, userEmail)
	if err != nil {
		log.Println(err.Error())
		return ctx.SendStatus(500)
	}

	return ctx.Status(200).JSON(map[string]string{
		"accessToken":  accessTokenString,
		"refreshToken": refreshTokenString,
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

	// Create access token
	accessTokenString, err := generateAccessToken(userID, username, userEmail)
	if err != nil {
		log.Println(err.Error())
		return ctx.SendStatus(500)
	}

	// Create refresh token
	refreshTokenString, err := generateRefreshToken(userID, username, userEmail)
	if err != nil {
		log.Println(err.Error())
		return ctx.SendStatus(500)
	}

	return ctx.Status(200).JSON(map[string]string{
		"accessToken":  accessTokenString,
		"refreshToken": refreshTokenString,
		"username":     username,
		"email":        userEmail,
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
	} else if regex, _ := regexp.Compile(`\b[\w\.-]+@[\w\.-]+\.\w{2,4}\b`); !regex.MatchString(registerData.Email) {
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
	if err != nil {
		log.Println(err.Error())
		return ctx.Status(500).JSON(map[string]string{
			"error": "Internal server error",
		})
	} else if result.Next() {
		return ctx.Status(409).JSON(map[string]string{
			"error": "Username already exist",
		})
	}

	// Check if email has been used
	result, err = db.Query("SELECT id FROM users WHERE email = ?", registerData.Email)
	if err != nil {
		log.Println(err.Error())
		return ctx.Status(500).JSON(map[string]string{
			"error": "Internal server error",
		})
	} else if result.Next() {
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

	// Create access token
	accessTokenString, err := generateAccessToken(userID, registerData.Username, registerData.Email)
	if err != nil {
		log.Println(err.Error())
		return ctx.SendStatus(500)
	}

	// Create refresh token
	refreshTokenString, err := generateRefreshToken(userID, registerData.Username, registerData.Email)
	if err != nil {
		log.Println(err.Error())
		return ctx.SendStatus(500)
	}

	return ctx.Status(200).JSON(map[string]string{
		"accessToken":  accessTokenString,
		"refreshToken": refreshTokenString,
		"username":     registerData.Username,
		"email":        registerData.Email,
	})
}
