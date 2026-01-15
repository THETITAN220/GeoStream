import random
import time
from datetime import datetime

import grpc
import telemetry_pb2
import telemetry_pb2_grpc


def run_simulation():
    """
    Connects to the Go gRPC backend and simulates a truck moving
    through city, sending telemetry every second.
    """
    channel = grpc.insecure_channel("localhost:50051")

    stub = telemetry_pb2_grpc.TelemetryServiceStub(channel)

    truck_id = "TRUCK-001"
    lat, lon = 40.7128, -74.0060  # Starting Coordinates (NYC)

    print(f"🚀 Simulation started for {truck_id}...")
    print("Press Ctrl+C to stop.")

    try:
        while True:
            lat += random.uniform(-0.0005, 0.0005)
            lon += random.uniform(-0.0005, 0.0005)

            request = telemetry_pb2.SendDataRequest(
                truck_id=truck_id,
                latitude=lat,
                longitude=lon,
                speed=random.uniform(30.0, 65.0),
                engine_temp=random.uniform(180.0, 210.0),
                timestamp=datetime.now().isoformat(),
            )

            try:
                # 5. Send the data to the Go Backend
                response = stub.SendData(request)

                if response.success:
                    print(
                        f"✅ Sent: {truck_id} | Pos: ({lat:.4f}, {lon:.4f}) | Msg: {response.message}"
                    )
                else:
                    print(f"⚠️ Backend rejected data: {response.message}")

            except grpc.RpcError as e:
                print(f"❌ gRPC Error: {e.code()} - {e.details()}")

            # 6. Wait 1 second before next update
            time.sleep(1)

    except KeyboardInterrupt:
        print("\n🛑 Simulation stopped by user.")
    finally:
        channel.close()


if __name__ == "__main__":
    run_simulation()
