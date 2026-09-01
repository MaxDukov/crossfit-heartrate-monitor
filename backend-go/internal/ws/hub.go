// Package ws — hub WebSocket-соединений с broadcast-рассылкой.
package ws

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"sync"

	"github.com/coder/websocket"
)

// Hub хранит активные соединения и рассылает JSON-сообщения всем клиентам.
// Заменяет Python-связку ConnectionManager + run_coroutine_threadsafe:
// любой горутине достаточно вызвать Broadcast.
type Hub struct {
	mu    sync.RWMutex
	conns map[*websocket.Conn]chan []byte
}

// NewHub создаёт пустой hub.
func NewHub() *Hub {
	return &Hub{conns: make(map[*websocket.Conn]chan []byte)}
}

// HandleWS обрабатывает HTTP-Upgrade до WebSocket: регистрирует клиента
// и читает входящие сообщения (keep-alive от фронтенда), пока тот не отключится.
func (h *Hub) HandleWS(w http.ResponseWriter, r *http.Request) {
	c, err := websocket.Accept(w, r, &websocket.AcceptOptions{
		InsecureSkipVerify: true, // CORS `*` как в Python-версии
	})
	if err != nil {
		slog.Error("ws accept", "err", err)
		return
	}

	out := make(chan []byte, 64)
	h.add(c, out)
	defer h.remove(c)

	// Писатель: рассылает broadcast-сообщения (завершается при close(out)
	// в remove или отмене контекста запроса).
	go func() {
		for msg := range out {
			if err := c.Write(r.Context(), websocket.MessageText, msg); err != nil {
				return
			}
		}
	}()

	// Читатель: игнорируем входящие (клиент шлёт только keep-alive).
	for {
		if _, _, err := c.Read(r.Context()); err != nil {
			return
		}
	}
}

// Broadcast сериализует payload в JSON и рассылает всем клиентам.
// Безопасно для вызова из любой горутины; не блокируется на медленных клиентах.
func (h *Hub) Broadcast(payload any) {
	data, err := json.Marshal(payload)
	if err != nil {
		slog.Error("ws marshal", "err", err)
		return
	}
	h.mu.RLock()
	defer h.mu.RUnlock()
	if len(h.conns) == 0 {
		return
	}
	for _, out := range h.conns {
		select {
		case out <- data:
		default: // переполнение — дропаем сообщение для медленного клиента
		}
	}
}

func (h *Hub) add(c *websocket.Conn, out chan []byte) {
	h.mu.Lock()
	h.conns[c] = out
	h.mu.Unlock()
	slog.Info("ws client connected", "total", len(h.conns))
}

func (h *Hub) remove(c *websocket.Conn) {
	h.mu.Lock()
	out, ok := h.conns[c]
	if ok {
		delete(h.conns, c)
		close(out)
	}
	h.mu.Unlock()
	if ok {
		_ = c.Close(websocket.StatusNormalClosure, "")
	}
	slog.Info("ws client disconnected", "total", len(h.conns))
}
