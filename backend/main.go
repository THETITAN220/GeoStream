package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"log"
	"net"
	"net/http"

	"github.com/THETITAN220/GeoStream/backend/consumer"
	pb "github.com/THETITAN220/GeoStream/proto/telemetry/v1"
	"github.com/gorilla/websocket"
	"github.com/segmentio/kafka-go"
	"google.golang.org/grpc"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

var clients = make(map[*websocket.Conn]bool)
var broadcast = make(chan *pb.SendDataRequest)

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

	go startBroadcastWorker()

	go func() {
		for data := range dataChan {
			broadcast <- data
			go SaveToDB(db, data)
		}
	}()


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

func enableCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*") // Allow any frontend
		w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func startRESTServer(db *sql.DB) {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /truck/{id}", func(w http.ResponseWriter, r *http.Request) {
		truckID := r.PathValue("id")

		location, err := GetLatestLocation(db, truckID)
		if err != nil {
			http.Error(w, "Truck Not Found", http.StatusNotFound)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(location)
	})
	
	mux.HandleFunc("/ws", handleWebSockets)

	log.Println(" NATIVE REST API RUNNING ON :8080...")
	log.Fatal(http.ListenAndServe(":8080", enableCORS(mux)))
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

func handleWebSockets(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket upgrade error: %v", err)
	}
	defer conn.Close()

	clients[conn] = true
	log.Println("WebSocket connection is successful")

	for {
		if _, _, err := conn.ReadMessage(); err != nil {
			delete(clients, conn)
			log.Println("Client disconnected")
			break
		}
	}

}

func startBroadcastWorker() {
	for {
		data := <-broadcast
		for client := range clients {
			err := client.WriteJSON(data)
			if err != nil {
				log.Printf("WebSocket write error: %v", err)
			}
			client.Close()
			delete(clients, client)
		}
	}
}

func main() {

	log.Println("****Starting GeoStream Server****")

	writer := initKafkaWriter()
	defer writer.Close()

	db := initDB()
	defer db.Close()

	dataChan := make(chan *pb.SendDataRequest, 100)

	startBackgroundWorkers(db, dataChan)

	go startRESTServer(db)

	startGRPCServer(writer)
}
