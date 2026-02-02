package service

import (
	"encoding/json"
	"time"

	"github.com/streadway/amqp"
)

type AuditMessage struct {
	Timestamp string `json:"timestamp"`
	Service   string `json:"service"`
	Action    string `json:"action"`
	Details   string `json:"details"`
	Severity  string `json:"severity"`
}

type AuditService struct {
	conn    *amqp.Connection
	channel *amqp.Channel
}

func NewAuditService(url string) (*AuditService, error) {
	conn, err := amqp.Dial(url)
	if err != nil {
		return nil, err
	}

	ch, err := conn.Channel()
	if err != nil {
		return nil, err
	}

	// Importante: En Go también declaramos el exchange por si no existe
	err = ch.ExchangeDeclare(
		"audit_exchange", // name
		"topic",          // type
		true,             // durable
		false,            // auto-deleted
		false,            // internal
		false,            // no-wait
		nil,              // arguments
	)

	return &AuditService{conn: conn, channel: ch}, err
}

func (s *AuditService) LogEvent(action, details, severity string) {
	msg := AuditMessage{
		Timestamp: time.Now().Format(time.RFC3339),
		Service:   "StorageService-Go",
		Action:    action,
		Details:   details,
		Severity:  severity,
	}

	body, _ := json.Marshal(msg)

	s.channel.Publish(
		"audit_exchange",    // exchange
		"audit.routing.key", // routing key
		false,
		false,
		amqp.Publishing{
			ContentType: "application/json",
			Body:        body,
		},
	)
}

// internal/service/audit_service.go

func (s *AuditService) Close() {
	if s.channel != nil {
		s.channel.Close()
	}
	if s.conn != nil {
		s.conn.Close()
	}
}
