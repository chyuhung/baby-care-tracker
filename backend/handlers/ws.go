package handlers

import (
	"baby-care-tracker/models"
	"encoding/json"
	"log"
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type WSHub struct {
	clients    map[int64]*models.WSClient
	broadcast  chan []byte
	register   chan *models.WSClient
	unregister chan *models.WSClient
	mu         sync.RWMutex
}

var Hub = &WSHub{
	clients:    make(map[int64]*models.WSClient),
	broadcast:  make(chan []byte, 256),
	register:   make(chan *models.WSClient),
	unregister: make(chan *models.WSClient),
}

func (h *WSHub) Run() {
	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			if old, ok := h.clients[client.UserID]; ok {
				old.CloseSend()
				delete(h.clients, client.UserID)
			}
			h.clients[client.UserID] = client
			count := len(h.clients)
			h.mu.Unlock()
			log.Printf("WS: 用户 %d 连接 (共 %d 个连接)", client.UserID, count)

		case client := <-h.unregister:
			h.mu.Lock()
			if existing, ok := h.clients[client.UserID]; ok && existing == client {
				delete(h.clients, client.UserID)
				client.CloseSend()
			}
			h.mu.Unlock()
			log.Printf("WS: 用户 %d 断开", client.UserID)

		case message := <-h.broadcast:
			h.mu.Lock()
			var dead []int64
			for userID, client := range h.clients {
				select {
				case client.Send <- message:
				default:
					client.CloseSend()
					dead = append(dead, userID)
				}
			}
			for _, uid := range dead {
				delete(h.clients, uid)
			}
			h.mu.Unlock()
		}
	}
}

// BroadcastMessage 向所有连接的客户端广播消息
func BroadcastMessage(msg models.WebSocketMessage) {
	data, err := json.Marshal(msg)
	if err != nil {
		return
	}
	select {
	case Hub.broadcast <- data:
	default:
		log.Println("WS: 广播队列满，丢弃消息")
	}
}

func HandleWebSocket(c *gin.Context) {
	tokenString := c.Query("token")
	if tokenString == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "需要认证"})
		return
	}

	userID, err := ParseToken(tokenString)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Token无效或已过期"})
		return
	}

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Println("WebSocket 升级失败:", err)
		return
	}

	client := &models.WSClient{
		UserID: userID,
		Send:   make(chan []byte, 256),
	}

	Hub.register <- client

	// 写入协程（唯一调用 conn.WriteMessage 的地方）
	go func() {
		defer conn.Close()
		for {
			message, ok := <-client.Send
			if !ok {
				conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			if err := conn.WriteMessage(websocket.TextMessage, message); err != nil {
				return
			}
		}
	}()

	// 读取协程
	go func() {
		defer func() {
			Hub.unregister <- client
		}()
		for {
			_, _, err := conn.ReadMessage()
			if err != nil {
				break
			}
		}
	}()
}
