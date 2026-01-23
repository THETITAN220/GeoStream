import random
import time
from datetime import datetime
from concurrent.futures import ThreadPoolExecutor

import grpc
import telemetry_pb2
import telemetry_pb2_grpc

# --- CONFIGURATION ---
BENGALURU_CENTER = (12.9716, 77.5946)
NUM_TRUCKS = 5  
BACKEND_ADDR = "localhost:50051"

def simulate_truck(truck_id):
    """
    Independent logic for a single truck. 
    It creates its own gRPC channel to simulate a real hardware device.
    """
    channel = grpc.insecure_channel(BACKEND_ADDR)
    stub = telemetry_pb2_grpc.TelemetryServiceStub(channel)

    lat = BENGALURU_CENTER[0] + random.uniform(-0.02, 0.02)
    lon = BENGALURU_CENTER[1] + random.uniform(-0.02, 0.02)

    print(f"🚛 {truck_id} coming online in Bengaluru...")

    try:
        while True:
            # Simulate movement
            lat += random.uniform(-0.0003, 0.0003)
            lon += random.uniform(-0.0003, 0.0003)

            request = telemetry_pb2.SendDataRequest(
                truck_id=truck_id,
                latitude=lat,
                longitude=lon,
                speed=random.uniform(20.0, 60.0),
                engine_temp=random.uniform(70.0, 100.0), 
                timestamp=datetime.now().isoformat(),
            )

            try:
                response = stub.SendData(request)
                if not response.success:
                    print(f"⚠️ {truck_id} REJECTED: {response.message}")
            except grpc.RpcError as e:
                print(f"❌ {truck_id} gRPC Error: {e.code()}")

            time.sleep(random.uniform(0.8, 1.2))

    except Exception as e:
        print(f"🛑 {truck_id} crashed: {e}")
    finally:
        channel.close()

def run_fleet_simulation():
    print(f"🚀 Starting GeoStream Fleet Simulation ({NUM_TRUCKS} trucks)...")
    
    with ThreadPoolExecutor(max_workers=NUM_TRUCKS) as executor:
        truck_ids = [f"TRUCK-{str(i+1).zfill(3)}" for i in range(NUM_TRUCKS)]
        executor.map(simulate_truck, truck_ids)

if __name__ == "__main__":
    try:
        run_fleet_simulation()
    except KeyboardInterrupt:
        print("\n🛑 Simulation stopped by user.")
