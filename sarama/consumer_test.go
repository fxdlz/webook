package sarama

import (
	"context"
	"github.com/IBM/sarama"
	"github.com/stretchr/testify/assert"
	"golang.org/x/sync/errgroup"
	"log"
	"testing"
	"time"
)

func TestConsumer(t *testing.T) {
	cfg := sarama.NewConfig()
	consumer, err := sarama.NewConsumerGroup(addr, "demo", cfg)
	assert.NoError(t, err)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
	defer cancel()
	start := time.Now()
	err = consumer.Consume(ctx, []string{"test_topic"}, &ConsumerHandler{})
	assert.NoError(t, err)
	log.Println(time.Since(start))
}

type ConsumerHandler struct {
}

func (c *ConsumerHandler) Setup(session sarama.ConsumerGroupSession) error {
	log.Println("This is Setup")
	//var offset int64 = 0
	//partitions := session.Claims()["test_topic"]
	//for _, p := range partitions {
	//	session.ResetOffset("test_topic", p, offset, "")
	//}
	return nil
}

func (c *ConsumerHandler) Cleanup(session sarama.ConsumerGroupSession) error {
	log.Println("This is Cleanup")
	return nil
}

func (c *ConsumerHandler) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	msgs := claim.Messages()
	for {
		const batchSize = 10
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		batch := make([]*sarama.ConsumerMessage, 0, batchSize)
		var eg errgroup.Group
		done := false
		for i := 0; i < batchSize && !done; i++ {
			select {
			case <-ctx.Done():
				done = true
			case msg, ok := <-msgs:
				if !ok {
					cancel()
					return nil
				}
				batch = append(batch, msg)
				eg.Go(func() error {
					log.Println(string(msg.Value))
					return nil
				})
			}
		}
		cancel()
		err := eg.Wait()
		if err != nil {
			log.Println(err)
			continue
		}
		for _, msg := range batch {
			session.MarkMessage(msg, "")
		}
	}
}
