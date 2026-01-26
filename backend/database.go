package main

import (
	"database/sql"
	"log"

	pb "github.com/THETITAN220/GeoStream/proto/telemetry/v1"
	_ "github.com/lib/pq"
)

func SaveToDB(db *sql.DB, dataChan <-chan *pb.SendDataRequest) {
	log.Println(" Database worker is waiting for data...")
	query := `INSERT INTO truck_locations (truck_id, latitude, longitude, speed, engine_temp, sent_at) VALUES ($1,$2,$3,$4,$5,$6)`

	for data := range dataChan {
		log.Printf(" DEBUG: Worker pulled from channel: %s", data.TruckId)
		_, err := db.Exec(query,
			data.TruckId,
			data.Latitude,
			data.Longitude,
			data.Speed,
			data.EngineTemp,
			data.Timestamp)
		if err != nil {
			log.Printf(" X DB INSERT ERROR: %v", err)
			continue
		}
		log.Printf(" SAVED TO DB: %s", data.TruckId)
	}

}

func GetLatestLocation(db *sql.DB, truckID string) (*pb.SendDataRequest, error) {
	var t pb.SendDataRequest
	query := ` SELECT truck_id, latitude, longitude, speed , engine_temp, sent_at
			FROM truck_locations
			WHERE truck_id = $1
			ORDER BY sent_at DESC LIMIT 1`
	err := db.QueryRow(query, truckID).Scan(&t.TruckId, &t.Latitude, &t.Longitude, &t.Speed, &t.EngineTemp, &t.Timestamp)
	if err != nil {
		return nil, err
	}
	return &t, nil
}
