package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"log"
	"net"

	"github.com/THETITAN220/GeoStream/backend/consumer"
	pb "github.com/THETITAN220/GeoStream/proto/telemetry/v1"
	"github.com/segmentio/kafka-go"
	"google.golang.org/grpc"
)

type server struct {
	pb.UnimplementedTelemetryServiceServer
	writer *kafka.Writer
}

func initKafkaWriter() *kafka.Writer {
	return &kafka.Writer{
		Addr:     kafka.TCP("geostream-kafka:9092"),
		Topic:    "truck-telemetry",
		Balancer: &kafka.Hash{}, // Ensures same TruckID always goes to same partition
	}

}

func initDB() *sql.DB {

	connStr := "host=postgres user=geo password=geostream dbname=geostream sslmode=disable"
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatalf(" FAILED TO CONNECT TO THE DATABASE: %v", err)
	}
	if err := db.Ping(); err != nil {
		log.Fatalf(" DB PING FAILED: %v", err)
	}
	log.Println(" CONNECTED TO THE DB")
	return db
}

func startBackgroundWorkers(db *sql.DB, dataChan chan *pb.SendDataRequest) {
	c := consumer.NewTelemetryConsumer([]string{"geostream-kafka:9092"}, "truck-telemetry", "geostream-group")
	go c.Start(context.Background(), dataChan)
	go SaveToDB(db, dataChan)
	log.Println(" BACKGROUND WORKERS INITIALIZD")
}

func startGRPCServer(writer *kafka.Writer) {
	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	s := grpc.NewServer()
	pb.RegisterTelemetryServiceServer(s, &server{writer: writer})

	log.Println(" Ingestor running on :50051...")
	if err := s.Serve(lis); err != nil {
		log.Fatalf("FAILED TO SERVE: %v", err)
	}
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

	log.Println("****Starting GeoStream Server****")

	writer := initKafkaWriter()
	defer writer.Close()

	db := initDB()
	defer db.Close()

	dataChan := make(chan *pb.SendDataRequest, 100)

	startBackgroundWorkers(db, dataChan)

	startGRPCServer(writer)
}
