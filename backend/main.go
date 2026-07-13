package main

import (
	"baby-care-tracker/database"
	"baby-care-tracker/handlers"
	"embed"
	"io/fs"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

//go:embed dist
var embeddedDist embed.FS

func getEnv(key, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultVal
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// SPAHandler serves the SPA index.html for all non-API, non-asset routes
func SPAHandler(staticDir string) gin.HandlerFunc {
	return func(c *gin.Context) {
		path := c.Request.URL.Path

		// Skip API and WebSocket
		if strings.HasPrefix(path, "/api") || path == "/ws" {
			c.Status(404)
			return
		}

		// If requesting a file that exists, serve it directly
		if staticDir != "" {
			filePath := filepath.Join(staticDir, filepath.Clean(strings.TrimPrefix(path, "/")))
			if info, err := os.Stat(filePath); err == nil && !info.IsDir() {
				http.ServeFile(c.Writer, c.Request, filePath)
				return
			}
		}

		// Serve index.html for SPA
		if staticDir != "" {
			http.ServeFile(c.Writer, c.Request, filepath.Join(staticDir, "index.html"))
		} else {
			// Embedded mode: serve from embed.FS
			subFS, _ := fs.Sub(embeddedDist, "dist")
			f, err := subFS.Open("index.html")
			if err != nil {
				c.Status(404)
				return
			}
			f.Close()
			http.FileServer(http.FS(subFS)).ServeHTTP(c.Writer, c.Request)
		}
	}
}

func main() {
	// Create data directory
	dataDir := getEnv("DATA_DIR", "/app/data")
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		log.Fatal("无法创建数据目录:", err)
	}

	// Init DB
	dbPath := filepath.Join(dataDir, "app.db")
	if err := database.InitDB(dbPath); err != nil {
		log.Fatal("数据库初始化失败:", err)
	}

	// Start WebSocket Hub
	go handlers.Hub.Run()

	// Config Gin
	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()

	// CORS
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		AllowCredentials: true,
	}))

	// Health check
	r.GET("/api/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok", "app": "Baby Care Tracker"})
	})

	// API routes
	api := r.Group("/api")
	{
		api.POST("/auth/register", handlers.Register)
		api.POST("/auth/login", handlers.Login)

		protected := api.Group("")
		protected.Use(JWTAuth())
		{
			protected.GET("/me", handlers.GetCurrentUser)

			protected.GET("/babies", handlers.GetBabies)
			protected.POST("/babies", handlers.CreateBaby)
			protected.PUT("/babies/:id", handlers.UpdateBaby)
			protected.DELETE("/babies/:id", handlers.DeleteBaby)
			protected.GET("/babies/:id/stats", handlers.GetStats)
			protected.GET("/babies/:id/trend", handlers.GetTrendStats)
			protected.GET("/babies/:id/latest-feeding", handlers.GetLatestFeeding)

			protected.GET("/babies/:id/records", handlers.GetRecords)
			protected.GET("/babies/:id/records/count", handlers.GetRecordsCount)
			protected.POST("/babies/:id/feeding", handlers.CreateFeeding)
			protected.POST("/babies/:id/diaper", handlers.CreateDiaper)
			protected.PUT("/records/:id", handlers.UpdateRecord)
			protected.DELETE("/records/:id", handlers.DeleteRecord)
		}
	}

	// WebSocket
	r.GET("/ws", handlers.HandleWebSocket)

	// SPA static files - use external dir if available, else embedded
	staticDir := getDistDir()
	if staticDir != "" {
		log.Printf("📁 使用前端资源: %s", staticDir)
		r.Static("/assets", filepath.Join(staticDir, "assets"))
	} else {
		log.Println("📦 使用内嵌前端资源")
	}

	// SPA catch-all: must be last route registered
	r.NoRoute(SPAHandler(staticDir))

	port := getEnv("PORT", "8080")
	log.Printf("🚀 Baby Care Tracker 启动中: http://0.0.0.0:%s", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatal("服务器启动失败:", err)
	}
}

func getDistDir() string {
	for _, dir := range []string{"frontend/dist", "dist"} {
		if fileExists(filepath.Join(dir, "index.html")) {
			return dir
		}
	}
	return ""
}

// JWT middleware
func JWTAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(401, gin.H{"error": "未登录"})
			c.Abort()
			return
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		if tokenString == authHeader {
			c.JSON(401, gin.H{"error": "Token格式错误"})
			c.Abort()
			return
		}

		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			return handlers.JWTSecret, nil
		})
		if err != nil || !token.Valid {
			c.JSON(401, gin.H{"error": "Token无效或已过期"})
			c.Abort()
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			c.JSON(401, gin.H{"error": "Token解析失败"})
			c.Abort()
			return
		}

		uid, ok := claims["uid"].(float64)
		if !ok {
			c.JSON(401, gin.H{"error": "Token缺少用户信息"})
			c.Abort()
			return
		}

		c.Set("user_id", int64(uid))
		c.Next()
	}
}
