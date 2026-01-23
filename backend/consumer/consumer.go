package consumer

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	pb "github.com/THETITAN220/GeoStream/proto/telemetry/v1"
	"github.com/segmentio/kafka-go"
)

type TelemetryConsumer struct {
	reader *kafka.Reader
}

func NewTelemetryConsumer(brokers []string, topic, groupID string) *TelemetryConsumer {
	return &TelemetryConsumer{
		reader: kafka.NewReader(kafka.ReaderConfig{
			Brokers:           brokers,
			Topic:             topic,
			GroupID:           groupID,
			SessionTimeout:    10 * time.Second,
			HeartbeatInterval: 3 * time.Second,
			MinBytes:          1,
			MaxBytes:          10e6,
			MaxWait:           100 * time.Millisecond,
			StartOffset:       kafka.FirstOffset,
		}),
	}
}

func (c *TelemetryConsumer) Start(ctx context.Context, dataChan chan<- *pb.SendDataRequest) {
	defer c.reader.Close()

	fmt.Println("Consumer started: listening for telemetry...")

	for {
		msg, err := c.reader.FetchMessage(ctx)
		if err != nil {
			log.Printf("Could not fetch messages: %v", err)
			time.Sleep(200 * time.Millisecond)
		}
		fmt.Printf("🔍 DEBUG: Kafka message received (key: %s)\n", string(msg.Key))
		var telemetry pb.SendDataRequest
		if err := json.Unmarshal(msg.Value, &telemetry); err != nil {
			log.Printf(" Failed to unmarshal telemetry: %v", err)
			continue
		}
		fmt.Printf("🔍 DEBUG: Sending to channel: %s\n", telemetry.TruckId)
		dataChan <- &telemetry
		fmt.Println("🔍 DEBUG: Sent to channel successfully")

		if err := c.reader.CommitMessages(ctx, msg); err != nil {
			log.Printf("Failed to commit messages: %v", err)
		}
	}

}
