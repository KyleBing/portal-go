// Command ws is the Go port of portal/bin/ws.ts. It runs a WebSocket server on
// port 9999 that answers heart-beat pings and maintains the diary "thumbs_up"
// counters, broadcasting updates to every connected client.
package main

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/KyleBing/portal-go/internal/config"
	"github.com/KyleBing/portal-go/internal/db"
	"github.com/gorilla/websocket"
)

const (
	msgHeartBeat = "heart-beat"
	msgThumbsUp  = "thumbs-up"
	addr         = ":9999"
)

// wsMessage mirrors the { type, content } envelope used by the Node server.
type wsMessage struct {
	Type    string      `json:"type"`
	Content interface{} `json:"content"`
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

// hub keeps track of every connected client so thumbs-up updates can be
// broadcast to all of them.
type hub struct {
	mu      sync.RWMutex
	clients map[*client]struct{}
}

// client wraps a websocket connection with its own write mutex, because a single
// connection must not be written to concurrently.
type client struct {
	conn    *websocket.Conn
	writeMu sync.Mutex
}

func (c *client) send(msg wsMessage) {
	c.writeMu.Lock()
	defer c.writeMu.Unlock()
	_ = c.conn.WriteJSON(msg)
}

func newHub() *hub {
	return &hub{clients: map[*client]struct{}{}}
}

func (h *hub) add(c *client) {
	h.mu.Lock()
	h.clients[c] = struct{}{}
	h.mu.Unlock()
}

func (h *hub) remove(c *client) {
	h.mu.Lock()
	delete(h.clients, c)
	h.mu.Unlock()
}

func (h *hub) broadcast(msg wsMessage) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	for c := range h.clients {
		c.send(msg)
	}
}

func now() string {
	return time.Now().Format("2006-01-02 15:04:05")
}

func main() {
	if _, err := config.Load(); err != nil {
		log.Fatalf("加载数据库配置失败: %v", err)
	}
	if _, err := db.Open(db.Diary); err != nil {
		log.Printf("%s 警告: 无法连接 diary 数据库: %v", now(), err)
	}

	h := newHub()
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		serveWS(h, w, r)
	})

	log.Printf("%s websocket 服务已运行在端口 9999", now())
	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatalf("websocket 服务启动失败: %v", err)
	}
}

func serveWS(h *hub, w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("%s websocket server: 升级连接失败: %v", now(), err)
		return
	}
	cl := &client{conn: conn}
	h.add(cl)
	log.Printf("%s 新客户端已连接", now())

	defer func() {
		h.remove(cl)
		_ = conn.Close()
		log.Printf("%s websocket server: 客户端已关闭连接", now())
	}()

	for {
		_, data, err := conn.ReadMessage()
		if err != nil {
			return
		}
		var incoming struct {
			Type    string `json:"type"`
			Content struct {
				Key string `json:"key"`
			} `json:"content"`
		}
		if err := json.Unmarshal(data, &incoming); err != nil {
			continue
		}
		switch incoming.Type {
		case msgHeartBeat:
			cl.send(wsMessage{Type: msgHeartBeat, Content: "pong"})
		case msgThumbsUp:
			handleThumbsUp(h, incoming.Content.Key)
		}
	}
}

func handleThumbsUp(h *hub, key string) {
	if key == "" {
		return
	}
	diary, err := db.Open(db.Diary)
	if err != nil {
		log.Printf("%s thumbs-up: 数据库连接失败: %v", now(), err)
		return
	}
	if _, err := diary.Exec(`UPDATE thumbs_up SET count = count + 1 WHERE name = ?`, key); err != nil {
		log.Printf("%s thumbs-up: 更新失败: %v", now(), err)
		return
	}
	var count int64
	if err := diary.QueryRow(`SELECT count FROM thumbs_up WHERE name = ?`, key).Scan(&count); err != nil {
		log.Printf("%s thumbs-up: 查询失败: %v", now(), err)
		return
	}
	h.broadcast(wsMessage{
		Type:    msgThumbsUp,
		Content: map[string]interface{}{"key": key, "count": count},
	})
}
