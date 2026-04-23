package main

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/gorilla/websocket"
)

type IncomingMessage struct {
	Type      string          `json:"type"`
	RequestID string          `json:"requestId,omitempty"`
	Payload   json.RawMessage `json:"payload,omitempty"`
}

type OutgoingMessage struct {
	Type      string      `json:"type"`
	MachineID string      `json:"machineId"`
	RequestID string      `json:"requestId,omitempty"`
	Timestamp int64       `json:"timestamp"`
	Payload   interface{} `json:"payload,omitempty"`
	Success   bool        `json:"success"`
	Error     string      `json:"error,omitempty"`
}

type MouseMovePayload struct {
	X int `json:"x"`
	Y int `json:"y"`
}

type MouseClickPayload struct {
	Button string `json:"button"`
}

type KeyPayload struct {
	Key  string `json:"key"`
	Text string `json:"text,omitempty"`
}

type Agent struct {
	cfg     *Config
	logger  *log.Logger
	conn    *websocket.Conn
	stopped bool
}

func NewAgent(cfg *Config, logger *log.Logger) *Agent {
	return &Agent{cfg: cfg, logger: logger}
}

func (a *Agent) Stop() {
	a.stopped = true
	if a.conn != nil {
		a.conn.Close()
	}
}

func (a *Agent) Connect() error {
	headers := map[string][]string{
		"X-Machine-ID": {a.cfg.MachineID},
		"X-Auth-Token": {a.cfg.AuthToken},
	}

	dialer := websocket.Dialer{HandshakeTimeout: 10 * time.Second}
	conn, _, err := dialer.Dial(a.cfg.ServerURL, headers)
	if err != nil {
		return fmt.Errorf("dial failed: %w", err)
	}
	a.conn = conn
	defer conn.Close()

	a.logger.Println("Connected to server")

	a.send(OutgoingMessage{
		Type:      "AGENT_REGISTER",
		MachineID: a.cfg.MachineID,
		Timestamp: time.Now().UnixMilli(),
		Success:   true,
		Payload: map[string]string{
			"machineId": a.cfg.MachineID,
			"version":   "1.0.0",
			"os":        "windows",
		},
	})

	go a.heartbeat()

	for {
		_, data, err := conn.ReadMessage()
		if err != nil {
			if a.stopped {
				return nil
			}
			return fmt.Errorf("read error: %w", err)
		}
		a.handleMessage(data)
	}
}

func (a *Agent) heartbeat() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()
	for range ticker.C {
		if a.stopped || a.conn == nil {
			return
		}
		a.send(OutgoingMessage{
			Type:      "HEARTBEAT",
			MachineID: a.cfg.MachineID,
			Timestamp: time.Now().UnixMilli(),
			Success:   true,
		})
	}
}

func (a *Agent) send(msg OutgoingMessage) {
	if a.conn == nil {
		return
	}
	data, err := json.Marshal(msg)
	if err != nil {
		a.logger.Printf("Failed to marshal message: %v", err)
		return
	}
	if err := a.conn.WriteMessage(websocket.TextMessage, data); err != nil {
		a.logger.Printf("Failed to send message: %v", err)
	}
}

func (a *Agent) sendError(requestID, msgType, errMsg string) {
	a.send(OutgoingMessage{
		Type:      msgType + "_RESPONSE",
		MachineID: a.cfg.MachineID,
		RequestID: requestID,
		Timestamp: time.Now().UnixMilli(),
		Success:   false,
		Error:     errMsg,
	})
}

func (a *Agent) handleMessage(data []byte) {
	var msg IncomingMessage
	if err := json.Unmarshal(data, &msg); err != nil {
		a.logger.Printf("Failed to parse message: %v", err)
		return
	}

	a.logger.Printf("Received: %s (requestId: %s)", msg.Type, msg.RequestID)

	switch msg.Type {
	case "SCREENSHOT_REQUEST":
		go a.handleScreenshot(msg.RequestID)

	case "MOUSE_MOVE":
		var p MouseMovePayload
		if err := json.Unmarshal(msg.Payload, &p); err != nil {
			a.sendError(msg.RequestID, "MOUSE_MOVE", "invalid payload")
			return
		}
		go a.handleMouseMove(msg.RequestID, p.X, p.Y)

	case "MOUSE_CLICK":
		var p MouseClickPayload
		if err := json.Unmarshal(msg.Payload, &p); err != nil {
			a.sendError(msg.RequestID, "MOUSE_CLICK", "invalid payload")
			return
		}
		go a.handleMouseClick(msg.RequestID, p.Button)

	case "KEY_PRESS":
		var p KeyPayload
		if err := json.Unmarshal(msg.Payload, &p); err != nil {
			a.sendError(msg.RequestID, "KEY_PRESS", "invalid payload")
			return
		}
		go a.handleKeyPress(msg.RequestID, p.Key)

	case "KEY_TYPE":
		var p KeyPayload
		if err := json.Unmarshal(msg.Payload, &p); err != nil {
			a.sendError(msg.RequestID, "KEY_TYPE", "invalid payload")
			return
		}
		go a.handleKeyType(msg.RequestID, p.Text)

	default:
		a.logger.Printf("Unknown message type: %s", msg.Type)
	}
}

func (a *Agent) handleScreenshot(requestID string) {
	img, err := CaptureScreen()
	if err != nil {
		a.logger.Printf("Screenshot failed: %v", err)
		a.sendError(requestID, "SCREENSHOT", err.Error())
		return
	}
	a.send(OutgoingMessage{
		Type:      "SCREENSHOT_RESPONSE",
		MachineID: a.cfg.MachineID,
		RequestID: requestID,
		Timestamp: time.Now().UnixMilli(),
		Success:   true,
		Payload:   map[string]string{"image": img, "format": "jpeg"},
	})
}

func (a *Agent) handleMouseMove(requestID string, x, y int) {
	if err := MouseMove(x, y); err != nil {
		a.sendError(requestID, "MOUSE_MOVE", err.Error())
		return
	}
	a.send(OutgoingMessage{Type: "MOUSE_MOVE_RESPONSE", MachineID: a.cfg.MachineID, RequestID: requestID, Timestamp: time.Now().UnixMilli(), Success: true})
}

func (a *Agent) handleMouseClick(requestID string, button string) {
	if err := MouseClick(button); err != nil {
		a.sendError(requestID, "MOUSE_CLICK", err.Error())
		return
	}
	a.send(OutgoingMessage{Type: "MOUSE_CLICK_RESPONSE", MachineID: a.cfg.MachineID, RequestID: requestID, Timestamp: time.Now().UnixMilli(), Success: true})
}

func (a *Agent) handleKeyPress(requestID string, key string) {
	if err := KeyPress(key); err != nil {
		a.sendError(requestID, "KEY_PRESS", err.Error())
		return
	}
	a.send(OutgoingMessage{Type: "KEY_PRESS_RESPONSE", MachineID: a.cfg.MachineID, RequestID: requestID, Timestamp: time.Now().UnixMilli(), Success: true})
}

func (a *Agent) handleKeyType(requestID string, text string) {
	if err := KeyType(text); err != nil {
		a.sendError(requestID, "KEY_TYPE", err.Error())
		return
	}
	a.send(OutgoingMessage{Type: "KEY_TYPE_RESPONSE", MachineID: a.cfg.MachineID, RequestID: requestID, Timestamp: time.Now().UnixMilli(), Success: true})
}
