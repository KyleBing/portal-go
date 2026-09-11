// Command ws runs WebSocket on :9999 for thumbs-up broadcasts and WebRTC
// signaling only. File bytes never transit this server (1Mbps host constraint).
package main

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"math/big"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/KyleBing/portal-go/internal/auth"
	"github.com/KyleBing/portal-go/internal/config"
	"github.com/KyleBing/portal-go/internal/db"
	"github.com/gorilla/websocket"
)

const (
	msgHeartBeat = "heart-beat"
	msgThumbsUp  = "thumbs-up"

	msgRTCCreate = "rtc-create"
	msgRTCJoin   = "rtc-join"
	msgRTCLeave  = "rtc-leave"
	msgRTCPeers  = "rtc-peers"
	msgRTCOffer  = "rtc-offer"
	msgRTCAnswer = "rtc-answer"
	msgRTCIce    = "rtc-ice"
	msgRTCError  = "rtc-error"

	addr = ":9999"

	// Public rooms expire after this much idle signaling time (no create/join/offer/answer/ice).
	roomIdleTTL      = 30 * time.Minute
	roomCleanupEvery = time.Minute
)

type wsMessage struct {
	Type    string      `json:"type"`
	Content interface{} `json:"content"`
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

type hub struct {
	mu      sync.RWMutex
	clients map[*client]struct{}
	rooms   map[string]*room
}

type room struct {
	code       string
	clients    map[string]*client // peerID -> client
	lastActive time.Time
}

type client struct {
	conn    *websocket.Conn
	writeMu sync.Mutex
	uid     int64
	peerID  string
	room    string
}

func (c *client) send(msg wsMessage) {
	c.writeMu.Lock()
	defer c.writeMu.Unlock()
	_ = c.conn.WriteJSON(msg)
}

func newHub() *hub {
	return &hub{
		clients: map[*client]struct{}{},
		rooms:   map[string]*room{},
	}
}

func (h *hub) add(c *client) {
	h.mu.Lock()
	h.clients[c] = struct{}{}
	h.mu.Unlock()
}

func (h *hub) remove(c *client) {
	h.leaveRoom(c)
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
	if err := auth.Init(); err != nil {
		log.Fatalf("JWT 初始化失败: %v（请设置环境变量 JWT_SECRET）", err)
	}
	if _, err := config.Load(); err != nil {
		log.Fatalf("加载数据库配置失败: %v", err)
	}
	if _, err := db.Open(db.Diary); err != nil {
		log.Printf("%s 警告: 无法连接 diary 数据库: %v", now(), err)
	}

	h := newHub()
	go h.cleanupExpiredRooms()
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		serveWS(h, w, r)
	})

	srv := &http.Server{Addr: addr, Handler: mux}
	go func() {
		log.Printf("%s websocket 服务已运行在端口 9999（thumbs-up + 公共 WebRTC 信令，房间空闲 %s 过期）", now(), roomIdleTTL)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("websocket 服务启动失败: %v", err)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	log.Printf("%s 正在优雅关闭 websocket …", now())
	_ = srv.Shutdown(ctx)
}

func serveWS(h *hub, w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("%s websocket server: 升级连接失败: %v", now(), err)
		return
	}

	uid, ok := authenticate(r)
	cl := &client{conn: conn, uid: uid, peerID: randomPeerID()}
	h.add(cl)
	if ok {
		log.Printf("%s 客户端已连接 uid=%d peer=%s", now(), uid, cl.peerID)
	} else {
		log.Printf("%s 匿名客户端已连接 peer=%s", now(), cl.peerID)
	}

	defer func() {
		h.remove(cl)
		_ = conn.Close()
		log.Printf("%s websocket server: 客户端已关闭 peer=%s", now(), cl.peerID)
	}()

	for {
		_, data, err := conn.ReadMessage()
		if err != nil {
			return
		}
		var envelope struct {
			Type    string          `json:"type"`
			Content json.RawMessage `json:"content"`
		}
		if err := json.Unmarshal(data, &envelope); err != nil {
			continue
		}
		switch envelope.Type {
		case msgHeartBeat:
			cl.send(wsMessage{Type: msgHeartBeat, Content: "pong"})
		case msgThumbsUp:
			var content struct {
				Key string `json:"key"`
			}
			_ = json.Unmarshal(envelope.Content, &content)
			handleThumbsUp(h, content.Key)
		case msgRTCCreate, msgRTCJoin, msgRTCLeave, msgRTCOffer, msgRTCAnswer, msgRTCIce:
			handleRTC(h, cl, envelope.Type, envelope.Content)
		}
	}
}

func authenticate(r *http.Request) (int64, bool) {
	// 优先 Authorization: Bearer，其次 query ?token=
	raw := auth.BearerToken(r.Header.Get("Authorization"))
	if raw == "" {
		raw = strings.TrimSpace(r.URL.Query().Get("token"))
	}
	if raw == "" {
		return 0, false
	}
	claims, err := auth.Parse(raw)
	if err != nil || claims.UID <= 0 {
		return 0, false
	}
	diary, err := db.Open(db.Diary)
	if err != nil {
		return 0, false
	}
	var found int64
	err = diary.QueryRow(`SELECT uid FROM users WHERE uid = ? LIMIT 1`, claims.UID).Scan(&found)
	if err == sql.ErrNoRows || err != nil {
		return 0, false
	}
	return found, true
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

func handleRTC(h *hub, cl *client, typ string, raw json.RawMessage) {
	var content map[string]interface{}
	_ = json.Unmarshal(raw, &content)
	if content == nil {
		content = map[string]interface{}{}
	}

	switch typ {
	case msgRTCCreate:
		code := randomRoomCode()
		h.mu.Lock()
		for h.rooms[code] != nil {
			code = randomRoomCode()
		}
		rm := &room{
			code:       code,
			clients:    map[string]*client{cl.peerID: cl},
			lastActive: time.Now(),
		}
		h.rooms[code] = rm
		cl.room = code
		h.mu.Unlock()
		cl.send(wsMessage{Type: msgRTCCreate, Content: map[string]interface{}{
			"room":   code,
			"peerId": cl.peerID,
			"peers":  []string{},
		}})

	case msgRTCJoin:
		code := normalizeRoomCode(asString(content["room"]))
		if code == "" {
			cl.send(wsMessage{Type: msgRTCError, Content: map[string]string{"message": "请输入 6 位数字房间码"}})
			return
		}
		if cl.room != "" && cl.room != code {
			h.leaveRoom(cl)
		}
		h.mu.Lock()
		rm := h.rooms[code]
		if rm == nil {
			h.mu.Unlock()
			cl.send(wsMessage{Type: msgRTCError, Content: map[string]string{"message": "房间不存在或已过期"}})
			return
		}
		if time.Since(rm.lastActive) > roomIdleTTL {
			expired := h.takeRoomLocked(code)
			h.mu.Unlock()
			h.notifyRoomExpired(expired)
			cl.send(wsMessage{Type: msgRTCError, Content: map[string]string{"message": "房间已过期"}})
			return
		}
		if len(rm.clients) >= 2 && rm.clients[cl.peerID] == nil {
			h.mu.Unlock()
			cl.send(wsMessage{Type: msgRTCError, Content: map[string]string{"message": "房间已满（仅支持两人）"}})
			return
		}
		peers := peerIDs(rm, cl.peerID)
		rm.clients[cl.peerID] = cl
		rm.lastActive = time.Now()
		cl.room = code
		h.mu.Unlock()

		cl.send(wsMessage{Type: msgRTCJoin, Content: map[string]interface{}{
			"room":   code,
			"peerId": cl.peerID,
			"peers":  peers,
		}})
		h.roomBroadcast(code, cl.peerID, wsMessage{Type: msgRTCPeers, Content: map[string]interface{}{
			"room":   code,
			"peers":  append(append([]string{}, peers...), cl.peerID),
			"joined": cl.peerID,
		}})

	case msgRTCLeave:
		h.leaveRoom(cl)

	case msgRTCOffer, msgRTCAnswer, msgRTCIce:
		to := asString(content["to"])
		if to == "" || cl.room == "" {
			cl.send(wsMessage{Type: msgRTCError, Content: map[string]string{"message": "缺少 to 或未在房间内"}})
			return
		}
		h.touchRoom(cl.room)
		payload := map[string]interface{}{
			"from": cl.peerID,
			"to":   to,
			"room": cl.room,
		}
		for k, v := range content {
			if k == "to" || k == "from" {
				continue
			}
			payload[k] = v
		}
		if !h.sendToPeer(cl.room, to, wsMessage{Type: typ, Content: payload}) {
			cl.send(wsMessage{Type: msgRTCError, Content: map[string]string{"message": "对方不在线"}})
		}
	}
}

func (h *hub) leaveRoom(cl *client) {
	h.mu.Lock()
	code := cl.room
	if code == "" {
		h.mu.Unlock()
		return
	}
	rm := h.rooms[code]
	if rm != nil {
		delete(rm.clients, cl.peerID)
		rm.lastActive = time.Now()
		if len(rm.clients) == 0 {
			delete(h.rooms, code)
		}
	}
	cl.room = ""
	remaining := []string{}
	if rm != nil {
		for id := range rm.clients {
			remaining = append(remaining, id)
		}
	}
	h.mu.Unlock()
	if len(remaining) > 0 {
		h.roomBroadcast(code, "", wsMessage{Type: msgRTCPeers, Content: map[string]interface{}{
			"room":  code,
			"peers": remaining,
			"left":  cl.peerID,
		}})
	}
}

func (h *hub) touchRoom(code string) {
	h.mu.Lock()
	if rm := h.rooms[code]; rm != nil {
		rm.lastActive = time.Now()
	}
	h.mu.Unlock()
}

// takeRoomLocked removes a room while hub.mu is held and clears members' room field.
func (h *hub) takeRoomLocked(code string) *room {
	rm := h.rooms[code]
	if rm == nil {
		return nil
	}
	delete(h.rooms, code)
	for _, c := range rm.clients {
		c.room = ""
	}
	return rm
}

func (h *hub) notifyRoomExpired(rm *room) {
	if rm == nil {
		return
	}
	msg := wsMessage{Type: msgRTCError, Content: map[string]string{"message": "房间已过期"}}
	for _, c := range rm.clients {
		c.send(msg)
	}
}

func (h *hub) cleanupExpiredRooms() {
	ticker := time.NewTicker(roomCleanupEvery)
	defer ticker.Stop()
	for range ticker.C {
		nowT := time.Now()
		var expired []*room
		h.mu.Lock()
		for code, rm := range h.rooms {
			if nowT.Sub(rm.lastActive) > roomIdleTTL {
				expired = append(expired, h.takeRoomLocked(code))
			}
		}
		h.mu.Unlock()
		for _, rm := range expired {
			if rm == nil {
				continue
			}
			log.Printf("%s 房间过期已清理 room=%s peers=%d", now(), rm.code, len(rm.clients))
			h.notifyRoomExpired(rm)
		}
	}
}

func (h *hub) roomBroadcast(code, exceptPeer string, msg wsMessage) {
	h.mu.RLock()
	rm := h.rooms[code]
	if rm == nil {
		h.mu.RUnlock()
		return
	}
	targets := make([]*client, 0, len(rm.clients))
	for id, c := range rm.clients {
		if id == exceptPeer {
			continue
		}
		targets = append(targets, c)
	}
	h.mu.RUnlock()
	for _, c := range targets {
		c.send(msg)
	}
}

func (h *hub) sendToPeer(code, peerID string, msg wsMessage) bool {
	h.mu.RLock()
	rm := h.rooms[code]
	if rm == nil {
		h.mu.RUnlock()
		return false
	}
	c := rm.clients[peerID]
	h.mu.RUnlock()
	if c == nil {
		return false
	}
	c.send(msg)
	return true
}

func peerIDs(rm *room, except string) []string {
	out := make([]string, 0, len(rm.clients))
	for id := range rm.clients {
		if id == except {
			continue
		}
		out = append(out, id)
	}
	return out
}

func randomRoomCode() string {
	// 6-digit code, easier to share/remember than alphanumeric
	n, err := rand.Int(rand.Reader, big.NewInt(1_000_000))
	if err != nil {
		return "000000"
	}
	return fmt.Sprintf("%06d", n.Int64())
}

func normalizeRoomCode(raw string) string {
	var b strings.Builder
	for _, r := range strings.TrimSpace(raw) {
		if r >= '0' && r <= '9' {
			b.WriteRune(r)
		}
	}
	return b.String()
}

func randomPeerID() string {
	b := make([]byte, 8)
	_, _ = rand.Read(b)
	return strings.ToLower(hexEncode(b))
}

func hexEncode(b []byte) string {
	const hexdigits = "0123456789abcdef"
	out := make([]byte, len(b)*2)
	for i, v := range b {
		out[i*2] = hexdigits[v>>4]
		out[i*2+1] = hexdigits[v&0x0f]
	}
	return string(out)
}

func asString(v interface{}) string {
	if v == nil {
		return ""
	}
	switch s := v.(type) {
	case string:
		return s
	default:
		return ""
	}
}
