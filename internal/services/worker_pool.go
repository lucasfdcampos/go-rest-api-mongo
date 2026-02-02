package services

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/lucas/go-rest-api-mongo/internal/dto"
	"github.com/lucas/go-rest-api-mongo/internal/messaging"
)

type WorkerPool struct {
	userService   *UserService
	kafkaProducer *messaging.KafkaProducer
	jobQueue      chan *dto.RegisterRequest
	workerCount   int
	batchSize     int
	batchTimeout  time.Duration
}

func NewWorkerPool(
	userService *UserService,
	kafkaProducer *messaging.KafkaProducer,
	workerCount, batchSize int,
	batchTimeout time.Duration) *WorkerPool {

	return &WorkerPool{
		userService:   userService,
		kafkaProducer: kafkaProducer,
		jobQueue:      make(chan *dto.RegisterRequest, 100), // Buffer size can be adjusted
		workerCount:   workerCount,
		batchSize:     batchSize,
		batchTimeout:  batchTimeout,
	}
}

func (wp *WorkerPool) Start(ctx context.Context) {
	for i := 0; i < wp.workerCount; i++ {
		go wp.worker(ctx)
	}
}

func (wp *WorkerPool) Submit(req *dto.RegisterRequest) error {
	select {
	case wp.jobQueue <- req:
		return nil
	default:
		return fmt.Errorf("worker pool queue is full")
	}
}

func (wp *WorkerPool) worker(ctx context.Context) {
	batch := make([]*dto.RegisterRequest, 0, wp.batchSize)
	ticker := time.NewTicker(wp.batchTimeout)
	defer ticker.Stop()

	for {
		select {
		case req := <-wp.jobQueue:
			batch = append(batch, req)
			if len(batch) >= wp.batchSize {
				wp.processBatch(ctx, batch)
				batch = batch[:0] // Reset batch
			}

		case <-ticker.C:
			if len(batch) > 0 {
				wp.processBatch(ctx, batch)
				batch = batch[:0]
			}

		case <-ctx.Done():
			// Processa batch restante antes de sair
			if len(batch) > 0 {
				wp.processBatch(context.Background(), batch)
			}
			return
		}
	}
}

func (wp *WorkerPool) processBatch(ctx context.Context, batch []*dto.RegisterRequest) {
	log.Printf("Processing batch of %d registrations", len(batch))

	for _, req := range batch {
		// Registra o usuário no banco
		user, err := wp.userService.Register(ctx, req)
		if err != nil {
			log.Printf("Error registering user %s: %v", req.Email, err)
			continue
		}

		// Publica evento no Kafka
		event := map[string]interface{}{
			"event_type": "user_registered",
			"user_id":    user.ID.Hex(),
			"email":      user.Email,
			"name":       user.Name,
			"timestamp":  time.Now().Unix(),
		}

		eventData, err := json.Marshal(event)
		if err != nil {
			log.Printf("Error marshaling event for user %s: %v", user.Email, err)
			continue
		}

		if err := wp.kafkaProducer.PublishUserRegistration(ctx, eventData); err != nil {
			log.Printf("Error publishing event for user %s: %v", user.Email, err)
		} else {
			log.Printf("Successfully registered and published event for user: %s", user.Email)
		}
	}
}
