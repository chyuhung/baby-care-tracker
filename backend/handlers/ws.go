package handlers

import (
	"baby-care-tracker/database"
	"baby-care-tracker/models"
	"encoding/json"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

// 保活参数：客户端是纯接收、不发任何数据，若不主动 ping，中间设备（运营商 NAT、
// 代理）的空闲回收会让连接被静默掐掉，而两侧都察觉不到——前端 wsConnected 会一直
// 谎报已连接。故服务端定期 ping，并给读侧设 pong 超时：收不到 pong 就主动关闭，
// 让客户端的 onclose 真的触发、重连逻辑跑起来。
const (
	wsPongWait   = 60 * time.Second
	wsPingPeriod = 30 * time.Second
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type WSHub struct {
	clients    map[int64]map[*models.WSClient]bool
	broadcast  chan wsFrame
	register   chan *models.WSClient
	unregister chan *models.WSClient
	mu         sync.RWMutex
}

// wsFrame 广播帧：data 是已序列化消息，familyID 决定投递范围（0 = 无家庭，投给任何人）
type wsFrame struct {
	data     []byte
	familyID int64
}

var Hub = &WSHub{
	clients:    make(map[int64]map[*models.WSClient]bool),
	broadcast:  make(chan wsFrame, 256),
	register:   make(chan *models.WSClient),
	unregister: make(chan *models.WSClient),
}

func (h *WSHub) Run() {
	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			if h.clients[client.UserID] == nil {
				h.clients[client.UserID] = make(map[*models.WSClient]bool)
			}
			h.clients[client.UserID][client] = true
			count := h.totalConnections()
			h.mu.Unlock()
			log.Printf("WS: 用户 %d 连接 (共 %d 个连接)", client.UserID, count)

		case client := <-h.unregister:
			h.mu.Lock()
			if set, ok := h.clients[client.UserID]; ok && set[client] {
				delete(set, client)
				client.CloseSend()
				if len(set) == 0 {
					delete(h.clients, client.UserID)
				}
			}
			h.mu.Unlock()
			log.Printf("WS: 用户 %d 断开", client.UserID)

		case frame := <-h.broadcast:
			h.mu.Lock()
			for userID, set := range h.clients {
				var dead []*models.WSClient
				for client := range set {
					// 按家庭过滤：不同家庭的连接不投递（familyID 0 不匹配任何帧）
					if client.FamilyID == 0 || client.FamilyID != frame.familyID {
						continue
					}
					select {
					case client.Send <- frame.data:
					default:
						client.CloseSend()
						dead = append(dead, client)
					}
				}
				for _, c := range dead {
					delete(set, c)
				}
				if len(set) == 0 {
					delete(h.clients, userID)
				}
			}
			h.mu.Unlock()
		}
	}
}

func (h *WSHub) totalConnections() int {
	n := 0
	for _, set := range h.clients {
		n += len(set)
	}
	return n
}

// BroadcastMessage 向 familyID 所属家庭的所有连接广播消息。
// familyID 必须由调用方显式给出（宝宝所属家庭），0 表示未知——直接丢弃，宁可不推也不越界。
func BroadcastMessage(msg models.WebSocketMessage, familyID int64) {
	if familyID == 0 {
		return
	}
	data, err := json.Marshal(msg)
	if err != nil {
		return
	}
	select {
	case Hub.broadcast <- wsFrame{data: data, familyID: familyID}:
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

	// 连接即带上所属家庭：注册/换家庭后仍以连接建立时的归属投递（重连即刷新）
	var familyID int64
	database.DB.QueryRow("SELECT family_id FROM users WHERE id = ?", userID).Scan(&familyID)

	client := &models.WSClient{
		UserID:   userID,
		FamilyID: familyID,
		Send:     make(chan []byte, 256),
	}

	Hub.register <- client

	// 读侧保活：ReadMessage 超时即代表对端/链路已死，break 后走 unregister 关闭连接。
	// Pong 由浏览器协议栈自动回应，这里只需把截止时间续上。
	_ = conn.SetReadDeadline(time.Now().Add(wsPongWait))
	conn.SetPongHandler(func(string) error {
		return conn.SetReadDeadline(time.Now().Add(wsPongWait))
	})

	// 写入协程（唯一调用 conn.WriteMessage 的地方 —— gorilla/websocket 不允许并发写）
	go func() {
		defer conn.Close()
		ticker := time.NewTicker(wsPingPeriod)
		defer ticker.Stop()
		for {
			select {
			case message, ok := <-client.Send:
				if !ok {
					conn.WriteMessage(websocket.CloseMessage, []byte{})
					return
				}
				if err := conn.WriteMessage(websocket.TextMessage, message); err != nil {
					return
				}
			case <-ticker.C:
				if err := conn.WriteMessage(websocket.PingMessage, nil); err != nil {
					return
				}
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
