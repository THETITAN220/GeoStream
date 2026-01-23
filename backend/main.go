package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"log"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/THETITAN220/GeoStream/backend/consumer"
	pb "github.com/THETITAN220/GeoStream/proto/telemetry/v1"
	"github.com/gorilla/websocket"
	"github.com/segmentio/kafka-go"
	"google.golang.org/grpc"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

var (
	clients   = make(map[*websocket.Conn]bool)
	clientsMu sync.Mutex
)

var (
	dataChan = make(chan *pb.SendDataRequest, 100)
	dbChan   = make(chan *pb.SendDataRequest, 100)
	wsChan   = make(chan *pb.SendDataRequest, 100)
)

type server struct {
	pb.UnimplementedTelemetryServiceServer
	writer *kafka.Writer
}

func initKafkaWriter() *kafka.Writer {
	return &kafka.Writer{
		Addr:     kafka.TCP("geostream-kafka:9092"),
		Topic:    "truck-telemetry",
		Balancer: &kafka.Hash{},
	}
}

func initDB() *sql.DB {
	connStr := "host=postgres user=geo password=geostream dbname=geostream sslmode=disable"

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatalf("❌ FAILED TO CONNECT TO DB: %v", err)
	}

	if err := db.Ping(); err != nil {
		log.Fatalf("❌ DB PING FAILED: %v", err)
	}

	log.Println("✅ CONNECTED TO DATABASE")
	return db
}

func waitForKafka(brokers []string, topic string) {
	log.Println("⏳ Waiting for Kafka to be ready...")
	
	for i := range 30 {
		conn, err := kafka.DialLeader(context.Background(), "tcp", brokers[0], topic, 0)
		if err == nil {
			conn.Close()
			log.Println("✅ Kafka is ready!")
			return
		}
		
		if i%5 == 0 {
			log.Printf("⏳ Kafka not ready yet... attempt %d/30 (error: %v)", i+1, err)
		}
		time.Sleep(1 * time.Second)
	}
	log.Fatal("❌ Kafka not available after 30 seconds")
}

func startFanoutWorker(
	in <-chan *pb.SendDataRequest,
	dbChan chan<- *pb.SendDataRequest,
	wsChan chan<- *pb.SendDataRequest,
) {
	log.Println("🔀 Fanout worker started")
	for data := range in {
		dbChan <- data
		select {
		case wsChan <- data:
		default:
			log.Println("⚠️ Websocket buffer full")
		}
	}
}

func startBackgroundWorkers(db *sql.DB) {
	brokers := []string{"geostream-kafka:9092"}
	topic := "truck-telemetry"
	
	waitForKafka(brokers, topic)
	
	c := consumer.NewTelemetryConsumer(brokers, topic, "geostream-group")
	
	// Start consumer in background
	go c.Start(context.Background(), dataChan)
	
	log.Println("⏳ Waiting for consumer to join group...")
	time.Sleep(5 * time.Second)
	
	go startFanoutWorker(dataChan, dbChan, wsChan)
	go SaveToDB(db, dbChan)
	go startBroadcastWorker()

	log.Println("✅ BACKGROUND WORKERS INITIALIZED")
}

func (s *server) SendData(ctx context.Context, req *pb.SendDataRequest) (*pb.SendDataResponse, error) {
	payload, _ := json.Marshal(req)

	err := s.writer.WriteMessages(ctx, kafka.Message{
		Key:   []byte(req.TruckId),
		Value: payload,
	})
	if err != nil {
		log.Printf("❌ Kafka write error: %v", err)
		return &pb.SendDataResponse{Success: false, Message: "Kafka Error"}, err
	}

	log.Printf("📥 Ingested: Truck=%s Lat=%.4f Lng=%.4f", req.TruckId, req.Latitude, req.Longitude)
	return &pb.SendDataResponse{Success: true, Message: "Stored in Kafka"}, nil
}

func startGRPCServer(writer *kafka.Writer) {
	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("❌ FAILED TO LISTEN: %v", err)
	}

	s := grpc.NewServer()
	pb.RegisterTelemetryServiceServer(s, &server{writer: writer})

	log.Println("🚀 gRPC INGESTOR RUNNING ON :50051")
	if err := s.Serve(lis); err != nil {
		log.Fatalf("❌ GRPC SERVE ERROR: %v", err)
	}
}

func enableCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func handleWebSockets(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("❌ WebSocket upgrade error: %v", err)
		return
	}

	clientsMu.Lock()
	clients[conn] = true
	clientsMu.Unlock()

	log.Println("🔌 WebSocket client connected")

	defer func() {
		clientsMu.Lock()
		delete(clients, conn)
		clientsMu.Unlock()
		conn.Close()
		log.Println("🔌 WebSocket client disconnected")
	}()

	for {
		if _, _, err := conn.ReadMessage(); err != nil {
			break
		}
	}
}

func startBroadcastWorker() {
	log.Println("📡 Broadcast worker started")
	for data := range wsChan {
		clientsMu.Lock()
		for client := range clients {
			if err := client.WriteJSON(data); err != nil {
				log.Printf("⚠️ Failed to send to WebSocket client: %v", err)
				client.Close()
				delete(clients, client)
			}
		}
		clientsMu.Unlock()
	}
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

	log.Println("🌐 REST API RUNNING ON :8080")
	log.Fatal(http.ListenAndServe(":8080", enableCORS(mux)))
}

func main() {
	log.Println("🚚 **** Starting GeoStream Server ****")

	writer := initKafkaWriter()
	defer writer.Close()

	db := initDB()
	defer db.Close()

	startBackgroundWorkers(db)

	go startRESTServer(db)

	startGRPCServer(writer)
}
