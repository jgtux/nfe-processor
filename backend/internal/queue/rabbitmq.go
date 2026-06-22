package queue

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/nfe-processor/backend/internal/config"
	"github.com/nfe-processor/backend/internal/domain"
)

// RabbitMQ wraps a persistent AMQP connection and channel.
type RabbitMQ struct {
	conn      *amqp.Connection
	ch        *amqp.Channel
	queueName string
}

// New connects to RabbitMQ with retry logic and declares a durable queue.
func New(cfg *config.RabbitMQConfig) (*RabbitMQ, error) {
	var conn *amqp.Connection
	var err error

	// Retry up to 10 times (RabbitMQ may take a few seconds to start)
	for i := 0; i < 10; i++ {
		conn, err = amqp.Dial(cfg.URL)
		if err == nil {
			break
		}
		log.Printf("[queue] waiting for RabbitMQ (%d/10): %v", i+1, err)
		time.Sleep(3 * time.Second)
	}
	if err != nil {
		return nil, fmt.Errorf("connect rabbitmq: %w", err)
	}

	ch, err := conn.Channel()
	if err != nil {
		return nil, fmt.Errorf("open channel: %w", err)
	}

	// Declare a durable queue so messages survive broker restarts
	_, err = ch.QueueDeclare(
		cfg.QueueName,
		true,  // durable
		false, // auto-delete
		false, // exclusive
		false, // no-wait
		nil,
	)
	if err != nil {
		return nil, fmt.Errorf("declare queue: %w", err)
	}

	// Fair dispatch: send at most 1 unacked message per worker
	if err := ch.Qos(1, 0, false); err != nil {
		return nil, fmt.Errorf("set qos: %w", err)
	}

	log.Printf("[queue] connected to RabbitMQ, queue=%s", cfg.QueueName)
	return &RabbitMQ{conn: conn, ch: ch, queueName: cfg.QueueName}, nil
}

// Publish serialises a QueueMessage and sends it to the queue with persistence.
func (r *RabbitMQ) Publish(msg domain.QueueMessage) error {
	body, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("marshal message: %w", err)
	}

	err = r.ch.Publish(
		"",           // default exchange
		r.queueName,
		false,        // mandatory
		false,        // immediate
		amqp.Publishing{
			ContentType:  "application/json",
			DeliveryMode: amqp.Persistent, // survive broker restart
			Body:         body,
		},
	)
	if err != nil {
		return fmt.Errorf("publish: %w", err)
	}

	log.Printf("[queue] published message id=%s", msg.UploadID)
	return nil
}

// Consume registers a handler function that is called for each incoming message.
// Messages are acknowledged only after the handler returns without error.
func (r *RabbitMQ) Consume(handler func(domain.QueueMessage) error) error {
	msgs, err := r.ch.Consume(
		r.queueName,
		"",    // consumer tag (auto)
		false, // auto-ack — we ack manually
		false, // exclusive
		false, // no-local
		false, // no-wait
		nil,
	)
	if err != nil {
		return fmt.Errorf("start consuming: %w", err)
	}

	log.Printf("[queue] consumer started, waiting for messages on %s", r.queueName)

	go func() {
		for d := range msgs {
			var msg domain.QueueMessage
			if err := json.Unmarshal(d.Body, &msg); err != nil {
				log.Printf("[queue] invalid message body: %v — nacking", err)
				_ = d.Nack(false, false) // discard malformed message
				continue
			}

			if err := handler(msg); err != nil {
				log.Printf("[queue] handler error for id=%s: %v — requeuing", msg.UploadID, err)
				_ = d.Nack(false, true) // requeue on transient errors
				continue
			}

			_ = d.Ack(false)
		}
	}()

	return nil
}

// Close gracefully shuts down the channel and connection.
func (r *RabbitMQ) Close() {
	if r.ch != nil {
		_ = r.ch.Close()
	}
	if r.conn != nil {
		_ = r.conn.Close()
	}
}
