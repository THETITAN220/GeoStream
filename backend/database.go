package main

import (
	"database/sql"
	"log"

	pb "github.com/THETITAN220/GeoStream/proto/telemetry/v1"
	_ "github.com/lib/pq"
)

func SaveToDB(db *sql.DB, dataChan <-chan *pb.SendDataRequest) {

	query := `INSERT INTO truck_locations (truck_id, latitude, longitude, speed, engine_temp, sent_at) VALUES ($1,$2,$3,$4,$5,$6)`

	for data := range dataChan {
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

