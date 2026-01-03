package main

import (
	"context"
	"encoding/json"
	"log"
	"net"

	pb "github.com/THETITAN220/GeoStream/proto/telemetry/v1"
	"github.com/segmentio/kafka-go"
	"google.golang.org/grpc"
)

type server struct {
	pb.UnimplementedTelemetryServiceServer
	writer *kafka.Writer
}

func (s *server) SendData(ctx context.Context, req *pb.SendDataRequest) (*pb.SendDataResponse, error) {
    payload, _ := json.Marshal(req)

    err := s.writer.WriteMessages(ctx, kafka.Message{
        Key:   []byte(req.TruckId),
        Value: payload,
    })

    if err != nil {
        return &pb.SendDataResponse{Success: false, Message: "Kafka Error"}, err
    }

    log.Printf(" Ingested: Truck %s | Lat: %.4f", req.TruckId, req.Latitude)
    return &pb.SendDataResponse{Success: true, Message: "Stored in Kafka"}, nil
}

func main() {
	
	log.Println("****Starting server****")
	writer := &kafka.Writer{
		Addr:     kafka.TCP("geostream-kafka:29092"), 
		Topic:    "truck-telemetry",
		Balancer: &kafka.Hash{}, // Ensures same TruckID always goes to same partition
	}
	defer writer.Close()

	// Start gRPC Server
	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	s := grpc.NewServer()
	pb.RegisterTelemetryServiceServer(s, &server{writer: writer})

	log.Println("🚀 Ingestor running on :50051...")
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
