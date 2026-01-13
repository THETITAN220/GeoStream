package consumer

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	pb "github.com/THETITAN220/GeoStream/proto/telemetry/v1"
	"github.com/segmentio/kafka-go"
)

type TelemetryConsumer struct {
	reader *kafka.Reader
}

func NewTelemetryConsumer(brokers []string, topic, groupID string) *TelemetryConsumer {
	return &TelemetryConsumer{
		reader: kafka.NewReader(kafka.ReaderConfig{
			Brokers:  brokers,
			Topic:    topic,
			GroupID:  groupID,
			MinBytes: 10e3,
			MaxBytes: 10e6,
		}),
	}
}

func (c *TelemetryConsumer) Start(ctx context.Context, dataChan chan<- *pb.SendDataRequest) {
	defer c.reader.Close()

	fmt.Println("Consumer started: listening for telemetry...")

	for {
		msg, err := c.reader.ReadMessage(ctx)
		if err != nil {
			log.Printf("Could not read the message: %v", err)
		}

		var telemetry pb.SendDataRequest
		if err := json.Unmarshal(msg.Value, &telemetry); err != nil {
			log.Printf(" Failed to unmarshal telemetry: %v", err)
			continue
		}

		dataChan <- &telemetry

	}

}
