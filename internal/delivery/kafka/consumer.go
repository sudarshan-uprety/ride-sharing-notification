package kafka

import (
	"context"
	"time"

	"github.com/segmentio/kafka-go"
)

type Consumer struct {
	reader  *kafka.Reader
	handler Handler
}

type Config struct {
	Brokers []string
	Topic   string
	GroupID string
}

func NewConsumer(broker []string, topic string, groupID string, handler Handler) *Consumer {
	return &Consumer{
		reader: kafka.NewReader(kafka.ReaderConfig{
			Brokers:        broker,
			Topic:          topic,
			GroupID:        groupID,
			MinBytes:       10,
			MaxBytes:       10e6,
			CommitInterval: time.Second,
			StartOffset:    kafka.FirstOffset,
		}),
		handler: handler,
	}
}

func (c *Consumer) Start(ctx context.Context) {

	for {
		select {
		case <-ctx.Done():
			return
		default:
			m, err := c.reader.ReadMessage(ctx)
			if err != nil {
				if ctx.Err() != nil {
					return
				}
				time.Sleep(1 * time.Second)
				continue
			}
			go func(msg kafka.Message) {
				if err := c.handler.Handle(msg); err != nil {
					return
				}
			}(m)
		}
	}
}

func (c *Consumer) Close() error {
	return c.reader.Close()
}
