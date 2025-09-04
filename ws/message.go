package ws

import "time"

type Message struct {
	Event     string    `json:"event"`
	Content   any       `json:"content"`
	Timestamp time.Time `json:"timestamp"`
}

func NewMessage(event string, content any) Message {
	return Message{
		Event:     event,
		Content:   content,
		Timestamp: time.Now(),
	}
}
