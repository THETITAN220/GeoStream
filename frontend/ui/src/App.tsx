import React, { useState } from 'react';
import { MapContainer, TileLayer, Marker, Popup } from 'react-leaflet';
import useWebSocket from 'react-use-websocket';
import 'leaflet/dist/leaflet.css';
import L from 'leaflet';
import icon from 'leaflet/dist/images/marker-icon.png';
import iconShadow from 'leaflet/dist/images/marker-shadow.png';

let DefaultIcon = L.icon({
    iconUrl: icon,
    shadowUrl: iconShadow,
    iconSize: [25, 41],
    iconAnchor: [12, 41]
});
L.Marker.prototype.options.icon = DefaultIcon;

interface TruckTelemetry {
    truck_id: string;
    latitude: number;
    longitude: number;
    speed: number;
    engine_temp: number;
    timestamp: string;
}

const App: React.FC = () => {
    const [fleet, setFleet] = useState<Record<string, TruckTelemetry>>({});

    const BENGALURU_CENTER: [number, number] = [12.9716, 77.5946];
    const WS_URL = 'ws://localhost:8080/ws';

    const { lastJsonMessage } = useWebSocket<TruckTelemetry>(WS_URL, {
        onOpen: () => console.log('✅ Connected to GeoStream WebSocket'),
        shouldReconnect: () => true,
        reconnectInterval: 3000,
        onMessage: (event) => {
            try {
                const data: TruckTelemetry = JSON.parse(event.data);
                setFleet(prevFleet => ({
                    ...prevFleet,
                    [data.truck_id]: data
                }));
            } catch (err) {
                console.error("Failed to parse WebSocket message:", err);
            }
        }
    });

    return (
        <div style={{ position: 'relative', height: '100vh', width: '100vw' }}>
            <div style={{
                position: 'absolute', top: 20, right: 20, zIndex: 1000,
                background: 'white', padding: '1.5rem', borderRadius: '12px',
                boxShadow: '0 8px 16px rgba(0,0,0,0.15)', fontFamily: 'Inter, system-ui, sans-serif',
                minWidth: '220px'
            }}>
                <h2 style={{ margin: '0 0 0.5rem 0', color: '#1a1a1a' }}>🚚 GeoStream Fleet</h2>
                <div style={{ fontSize: '1.1rem', color: '#2563eb', marginBottom: '1rem' }}>
                    Active Trucks: <strong>{Object.keys(fleet).length}</strong>
                </div>

                <div style={{ maxHeight: '300px', overflowY: 'auto', fontSize: '0.85rem' }}>
                    {Object.values(fleet).map(truck => (
                        <div key={truck.truck_id} style={{ marginBottom: '8px', paddingBottom: '8px', borderBottom: '1px solid #eee' }}>
                            <strong>{truck.truck_id}</strong>: {truck.speed.toFixed(1)} km/h
                        </div>
                    ))}
                </div>
            </div>

            <MapContainer
                center={BENGALURU_CENTER}
                zoom={13}
                zoomControl={true}
                style={{ height: '100%', width: '100%' }}
            >
                <TileLayer
                    url="https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png"
                    attribution='&copy; <a href="https://www.openstreetmap.org/copyright">OpenStreetMap</a>'
                />

                {Object.values(fleet).map((truck) => (
                    <Marker
                        key={truck.truck_id}
                        position={[truck.latitude, truck.longitude]}
                    >
                        <Popup>
                            <div style={{ textAlign: 'center' }}>
                                <strong>{truck.truck_id}</strong><br />
                                Speed: {truck.speed.toFixed(1)} km/h<br />
                                Temp: {truck.engine_temp.toFixed(1)}°C
                            </div>
                        </Popup>
                    </Marker>
                ))}
            </MapContainer>
        </div>
    );
};

export default App;
