package handlers

import (
	"baby-care-tracker/database"
	"baby-care-tracker/models"
	"bytes"
	"crypto/rand"
	"encoding/base64"
	"log"
	"math/big"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

const inviteChars = "ABCDEFGHJKLMNPQRSTUVWXYZ23456789"

func generateInviteCode() string {
	code := make([]byte, 6)
	for i := range code {
		n, err := rand.Int(rand.Reader, big.NewInt(int64(len(inviteChars))))
		if err != nil {
			buf := make([]byte, 1)
			rand.Read(buf)
			code[i] = inviteChars[int(buf[0])%len(inviteChars)]
			continue
		}
		code[i] = inviteChars[n.Int64()]
	}
	return string(code)
}

// EnsureUserHasFamily 确保用户有所属家庭，没有则自动创建
func EnsureUserHasFamily(userID int64) (int64, error) {
	var familyID int64
	err := database.DB.QueryRow("SELECT family_id FROM users WHERE id = ?", userID).Scan(&familyID)
	if err != nil || familyID == 0 {
		code := generateInviteCode()
		res, err := database.DB.Exec("INSERT INTO families (invite_code) VALUES (?)", code)
		if err != nil {
			return 0, err
		}
		familyID, err = res.LastInsertId()
		if err != nil {
			return 0, err
		}
		_, err = database.DB.Exec("UPDATE users SET family_id = ? WHERE id = ?", familyID, userID)
		if err != nil {
			return 0, err
		}
	}
	return familyID, nil
}

// getJWTSecret 取 JWT 签名密钥：优先 JWT_SECRET 环境变量，否则读取数据目录的
// 持密钥文件（首次启动随机生成并落盘，与 app.db 同目录，重启不变）。
// 不提供任何硬编码默认值——仓库内置的默认密钥等于公开密钥，任何拿到代码的人都能签出
// 有效 token（历史提交里就有一枚用默认密钥签的 demo_token.txt）。
// 首次部署到本实现后旧 token 全部失效，需重新登录一次。
func getJWTSecret() []byte {
	if secret := os.Getenv("JWT_SECRET"); secret != "" {
		return []byte(secret)
	}
	dataDir := os.Getenv("DATA_DIR")
	if dataDir == "" {
		dataDir = "/app/data" // 与 main.go getEnv("DATA_DIR", "/app/data") 保持一致
	}
	path := filepath.Join(dataDir, "jwt_secret")
	if b, err := os.ReadFile(path); err == nil {
		if s := bytes.TrimSpace(b); len(s) >= 32 {
			return s
		}
	}
	buf := make([]byte, 48)
	if _, err := rand.Read(buf); err != nil {
		log.Fatalf("无法生成 JWT 密钥: %v", err)
	}
	secret := []byte(base64.RawURLEncoding.EncodeToString(buf))
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		log.Printf("⚠️ 无法创建数据目录 %s，本次密钥重启后失效（需重新登录）: %v", dataDir, err)
		return secret
	}
	if err := os.WriteFile(path, secret, 0600); err != nil {
		log.Printf("⚠️ 无法写入 %s，本次密钥重启后失效（需重新登录）: %v", path, err)
		return secret
	}
	log.Printf("✅ 已生成 JWT 密钥: %s", path)
	return secret
}

var JWTSecret = getJWTSecret()

func getJWTWithUserID(userID int64, expirationHours int) string {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"exp": time.Now().Add(time.Duration(expirationHours) * time.Hour).Unix(),
		"iat": time.Now().Unix(),
		"uid": userID,
	})
	tokenString, err := token.SignedString(JWTSecret)
	if err != nil {
		return ""
	}
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

	userID, err := result.LastInsertId()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "注册失败"})
		return
	}

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
			ID:        userID,
			Username:  req.Username,
			FamilyID:  &familyID,
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

	EnsureUserHasFamily(user.ID)

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
