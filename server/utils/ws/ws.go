// Package ws 提供 WebSocket 服务端能力, 当前用于跳板机交互终端。
package ws

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

// 默认 upgrader: 关闭压缩以外的限制, 允许跨源(鉴权由调用方控制)
var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	// 跳板机 WS 由 PrivateGroup 中间件链完成鉴权, 这里放行所有 Origin
	CheckOrigin: func(r *http.Request) bool { return true },
}

// Conn 封装单个 WebSocket 连接, 提供并发安全的读写与关闭通知。
type Conn struct {
	ws     *websocket.Conn
	mu     sync.Mutex
	closed bool
	done   chan struct{}
	once   sync.Once
}

// Upgrade 把 gin.Context 升级为 WebSocket 连接。
func Upgrade(c *gin.Context) (*Conn, error) {
	wsConn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		return nil, err
	}
	// 终端交互可能长时间空闲, 关闭默认读超时
	wsConn.SetReadDeadline(time.Time{})
	return &Conn{ws: wsConn, done: make(chan struct{})}, nil
}

// ReadText 读取一条文本消息(阻塞)。
func (c *Conn) ReadText() (string, error) {
	_, msg, err := c.ws.ReadMessage()
	return string(msg), err
}

// WriteText 发送一条文本消息(并发安全)。
func (c *Conn) WriteText(msg string) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.closed {
		return nil
	}
	return c.ws.WriteMessage(websocket.TextMessage, []byte(msg))
}

// Close 关闭连接并解除 Done 阻塞(幂等)。
func (c *Conn) Close() error {
	c.once.Do(func() {
		c.mu.Lock()
		c.closed = true
		c.mu.Unlock()
		_ = c.ws.Close()
		close(c.done)
	})
	return nil
}

// Done 返回连接关闭时解除阻塞的 channel。
func (c *Conn) Done() <-chan struct{} {
	return c.done
}
