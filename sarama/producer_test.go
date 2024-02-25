package sarama

import (
	"github.com/IBM/sarama"
	"github.com/stretchr/testify/assert"
	"testing"
)

var addr = []string{"localhost:9094"}

func TestSyncSarama(t *testing.T) {
	cfg := sarama.NewConfig()
	cfg.Producer.Return.Successes = true
	producer, err := sarama.NewSyncProducer(addr, cfg)
	assert.NoError(t, err)
	p, offset, err := producer.SendMessage(&sarama.ProducerMessage{
		Topic: "test_topic",
		Value: sarama.StringEncoder("这是一条消息"),
		Headers: []sarama.RecordHeader{
			{
				Key:   []byte("header1"),
				Value: []byte("header1_value"),
			},
		},
		Metadata: "metadata"})
	assert.NoError(t, err)
	t.Log(p, offset)
}

func TestAsyncSarama(t *testing.T) {
	cfg := sarama.NewConfig()
	cfg.Producer.Return.Successes = true
	cfg.Producer.Return.Errors = true
	producer, err := sarama.NewAsyncProducer(addr, cfg)
	assert.NoError(t, err)
	producerMsg := producer.Input()
	producerMsg <- &sarama.ProducerMessage{
		Topic: "test_topic",
		Value: sarama.StringEncoder("这是一条消息"),
		Headers: []sarama.RecordHeader{
			{
				Key:   []byte("header1"),
				Value: []byte("header1_value"),
			},
		},
		Metadata: "metadata"}
	select {
	case Msg := <-producer.Successes():
		val, _ := Msg.Value.Encode()
		t.Log("成功了", string(val))
	case err := <-producer.Errors():
		val, _ := err.Msg.Value.Encode()
		t.Log(err.Err, string(val))
	}
}
