package handlers

import (
	"baby-care-tracker/database"
	"baby-care-tracker/models"
	"crypto/rand"
	"math/big"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

const inviteChars = "ABCDEFGHJKLMNPQRSTUVWXYZ23456789"

func generateInviteCode() string {
	code := make([]byte, 6)
	for i := range code {
		n, _ := rand.Int(rand.Reader, big.NewInt(int64(len(inviteChars))))
		code[i] = inviteChars[n.Int64()]
	}
	return string(code)
}

// EnsureUserHasFamily 确保用户有所属家庭，没有则自动创建
func EnsureUserHasFamily(userID int64) (int64, error) {
	var familyID int64
	err := database.DB.QueryRow("SELECT family_id FROM users WHERE id = ?", userID).Scan(&familyID)
	if err != nil || familyID == 0 {
		// 创建新家庭
		code := generateInviteCode()
		res, err := database.DB.Exec("INSERT INTO families (invite_code) VALUES (?)", code)
		if err != nil {
			return 0, err
		}
		familyID, _ = res.LastInsertId()
		database.DB.Exec("UPDATE users SET family_id = ? WHERE id = ?", familyID, userID)
	}
	return familyID, nil
}

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

	// 自动创建家庭
	var familyID int64
	code := generateInviteCode()
	familyRes, err := database.DB.Exec("INSERT INTO families (invite_code) VALUES (?)", code)
	if err == nil {
		familyID, _ = familyRes.LastInsertId()
		database.DB.Exec("UPDATE users SET family_id = ? WHERE id = ?", familyID, userID)
	}

	token := getJWTWithUserID(userID, 168) // 7 days

	c.JSON(http.StatusCreated, models.AuthResponse{
		Token: token,
		User: models.User{
			ID:       userID,
			Username: req.Username,
			FamilyID: &familyID,
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
	var familyIDRaw *int64
	err := database.DB.QueryRow(
		"SELECT id, username, password_hash, family_id, created_at FROM users WHERE username = ?",
		req.Username,
	).Scan(&user.ID, &user.Username, &passwordHash, &familyIDRaw, &user.CreatedAt)

	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "用户名或密码错误"})
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(req.Password)); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "用户名或密码错误"})
		return
	}

	// 确保用户有家庭（兼容旧数据）
	EnsureUserHasFamily(user.ID)

	// 重新查询完整用户信息
	database.DB.QueryRow(
		"SELECT id, username, family_id, created_at FROM users WHERE id = ?",
		user.ID,
	).Scan(&user.ID, &user.Username, &user.FamilyID, &user.CreatedAt)

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
		"SELECT id, username, family_id, created_at FROM users WHERE id = ?",
		userID,
	).Scan(&user.ID, &user.Username, &user.FamilyID, &user.CreatedAt)

	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "用户不存在"})
		return
	}

	c.JSON(http.StatusOK, user)
}
