package handlers

import (
	"baby-care-tracker/database"
	"baby-care-tracker/models"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

var JWTSecret = []byte("baby-care-secret-key-2024")

func getJWTWithUserID(userID int64, expirationHours int) string {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"exp": time.Now().Add(time.Duration(expirationHours) * time.Hour).Unix(),
		"iat": time.Now().Unix(),
		"uid": userID,
	})
	tokenString, _ := token.SignedString(JWTSecret)
	return tokenString
}

// ParseToken 解析JWT并返回userID
func ParseToken(tokenString string) (int64, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return JWTSecret, nil
	})
	if err != nil || !token.Valid {
		return 0, err
	}
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return 0, jwt.ErrTokenInvalidClaims
	}
	uid, ok := claims["uid"].(float64)
	if !ok {
		return 0, jwt.ErrTokenInvalidClaims
	}
	return int64(uid), nil
}

func Register(c *gin.Context) {
	var req models.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "用户名和密码必填，密码至少6位"})
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "密码加密失败"})
		return
	}

	result, err := database.DB.Exec(
		"INSERT INTO users (username, password_hash) VALUES (?, ?)",
		req.Username, string(hashedPassword),
	)
	if err != nil {
		c.JSON(http.StatusConflict, gin.H{"error": "用户名已存在"})
		return
	}

	userID, _ := result.LastInsertId()
	token := getJWTWithUserID(userID, 168) // 7 days

	c.JSON(http.StatusCreated, models.AuthResponse{
		Token: token,
		User: models.User{
			ID:        userID,
			Username:  req.Username,
			CreatedAt: time.Now(),
		},
	})
}

func Login(c *gin.Context) {
	var req models.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "用户名和密码必填"})
		return
	}

	var user models.User
	var passwordHash string
	err := database.DB.QueryRow(
		"SELECT id, username, password_hash, created_at FROM users WHERE username = ?",
		req.Username,
	).Scan(&user.ID, &user.Username, &passwordHash, &user.CreatedAt)

	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "用户名或密码错误"})
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(req.Password)); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "用户名或密码错误"})
		return
	}

	token := getJWTWithUserID(user.ID, 168)

	c.JSON(http.StatusOK, models.AuthResponse{
		Token: token,
		User:  user,
	})
}

// GetCurrentUser 获取当前用户信息
func GetCurrentUser(c *gin.Context) {
	userID := c.GetInt64("user_id")

	var user models.User
	err := database.DB.QueryRow(
		"SELECT id, username, created_at FROM users WHERE id = ?",
		userID,
	).Scan(&user.ID, &user.Username, &user.CreatedAt)

	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "用户不存在"})
		return
	}

	c.JSON(http.StatusOK, user)
}
