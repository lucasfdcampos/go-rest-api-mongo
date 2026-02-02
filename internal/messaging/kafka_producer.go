package messaging

import (
	"context"

	"github.com/lucas/go-rest-api-mongo/internal/config"
	"github.com/segmentio/kafka-go"
)

type KafkaProducer struct {
	writer *kafka.Writer
}

func NewKafkaProducer(cfg *config.Config) *KafkaProducer {
	return &KafkaProducer{
		writer: &kafka.Writer{
			Addr:         kafka.TCP(cfg.Kafka.Brokers...),
			Topic:        cfg.Kafka.TopicUserRegistration,
			Balancer:     &kafka.LeastBytes{},
			RequiredAcks: kafka.RequireOne,
			Async:        false,
		},
	}
}

func (kp *KafkaProducer) PublishUserRegistration(ctx context.Context, data []byte) error {
	return kp.writer.WriteMessages(ctx, kafka.Message{
		Value: data,
	})
}

func (kp *KafkaProducer) Close() error {
	return kp.writer.Close()
}
