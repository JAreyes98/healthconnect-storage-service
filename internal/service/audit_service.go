package service

import (
	"encoding/json"
	"log"
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
	if url == "" {
		log.Println("[WARN] RABBITMQ_URL no configurada. Servicio de auditoría deshabilitado.")
		return &AuditService{}, nil
	}

	conn, err := amqp.Dial(url)
	if err != nil {
		return nil, err
	}

	ch, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, err
	}

	err = ch.ExchangeDeclare(
		"audit_exchange", // name
		"topic",          // type
		true,             // durable
		false,            // auto-deleted
		false,            // internal
		false,            // no-wait
		nil,              // arguments
	)
	if err != nil {
		ch.Close()
		conn.Close()
		return nil, err
	}

	return &AuditService{conn: conn, channel: ch}, nil
}

func (s *AuditService) LogEvent(action, details, severity string) {
	if s == nil || s.channel == nil {
		return
	}

	msg := AuditMessage{
		Timestamp: time.Now().Format(time.RFC3339),
		Service:   "StorageService-Go",
		Action:    action,
		Details:   details,
		Severity:  severity,
	}

	body, _ := json.Marshal(msg)

	_ = s.channel.Publish(
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

func (s *AuditService) Close() {
	if s == nil {
		return
	}
	if s.channel != nil {
		s.channel.Close()
	}
	if s.conn != nil {
		s.conn.Close()
	}
}
