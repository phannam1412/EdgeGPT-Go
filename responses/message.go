package responses

import (
	"encoding/json"
	"github.com/gorilla/websocket"
	"sync"
	"time"
)

const DelimiterByte = uint8(30)

const (
	TypeUpdate float64 = 1
	TypeFinish float64 = 2
)

// MessageWrapper wrapper for message worker
type MessageWrapper struct {
	Final    bool
	Question string
	Answer   GptResponse
	Chan     chan []byte
	mu       *sync.Mutex
	conn     *websocket.Conn
}

// Message struct for bing answer
type Message struct {
	Text          string `json:"text"`
	Author        string `json:"author"`
	AdaptiveCards []struct {
		Type    string `json:"type"`
		Version string `json:"version"`
		Body    []struct {
			Type string `json:"type"`
			Text string `json:"text"`
			Wrap bool   `json:"wrap"`
		} `json:"body"`
	} `json:"adaptiveCards,omitempty"`
	SuggestedResponses []*Suggestion `json:"suggestedResponses,omitempty"`
}

// Suggestion for this question
type Suggestion struct {
	Text        string    `json:"text"`
	Author      string    `json:"author"`
	CreatedAt   time.Time `json:"createdAt"`
	Timestamp   time.Time `json:"timestamp"`
	MessageId   string    `json:"messageId"`
	MessageType string    `json:"messageType"`
	Offense     string    `json:"offense"`
	Feedback    struct {
		Tag       interface{} `json:"tag"`
		UpdatedOn interface{} `json:"updatedOn"`
		Type      string      `json:"type"`
	} `json:"feedback"`
	ContentOrigin string      `json:"contentOrigin"`
	Privacy       interface{} `json:"privacy"`
}

// NewMessageWrapper create new wrapper
func NewMessageWrapper(question string, mutex *sync.Mutex, conn *websocket.Conn) *MessageWrapper {
	return &MessageWrapper{
		Question: question,
		Chan:     make(chan []byte, 1),
		mu:       mutex,
		conn:     conn,
	}
}

// Worker write and parse websocket messages
func (m *MessageWrapper) Worker() error {
	defer m.mu.Unlock()

	var response map[string]any
	var updateResponse Update
	var finalResponse Final
	var undefinedResponse Undefined

	for {
		var messages []byte

		/* EXAMPLE ERROR MESSAGE:
		{
		  "type": 2,
		  "invocationId": "\u0000",
		  "item": {
		    "firstNewMessageIndex": null,
		    "defaultChatName": null,
		    "conversationId": "...",
		    "requestId": "...",
		    "telemetry": {
		      "startTime": "2023-09-08T05:09:42.4656485Z"
		    },
		    "result": {
		      "value": "UnauthorizedRequest",
		      "message": "Conversation signature verification failed. The signature was not the one that was expected.",
		      "error": "UnauthorizedRequest",
		      "renewCert": true,
		      "serviceVersion": "20230906.174"
		    }
		  }
		}{
		  "type": 3,
		  "invocationId": "\u0000"
		}
		*/
		_, original, err := m.conn.ReadMessage()
		if err != nil {
			return err
		}

		// read to delimiter
		for _, b := range original {
			if b == DelimiterByte {
				break
			}

			messages = append(messages, b)
		}

		if err := json.Unmarshal(messages, &response); err != nil {
			return err
		}

		switch response["type"] {
		case TypeUpdate:
			if err := json.Unmarshal(messages, &updateResponse); err != nil {
				return err
			}

			m.Answer = &updateResponse
			break
		case TypeFinish:
			if err := json.Unmarshal(messages, &finalResponse); err != nil {
				return err
			}
			m.Answer = &finalResponse
			m.Final = true
			m.Chan <- messages
			close(m.Chan)
			return nil

		default:
			if err := json.Unmarshal(messages, &undefinedResponse); err != nil {
				return err
			}

			m.Answer = &undefinedResponse
		}

		m.Chan <- messages
	}

}
