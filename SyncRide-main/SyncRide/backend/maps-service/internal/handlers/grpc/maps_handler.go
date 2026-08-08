package grpc

import (
	"context"
	"math"
	"math/rand"

	pb "github.com/syncride/backend/shared/proto/maps"
)

type MapsHandler struct {
	pb.UnimplementedMapsServiceServer
}

func NewMapsHandler() *MapsHandler {
	return &MapsHandler{}
}

// Haversine formula to calculate the great-circle distance between two points on a sphere
func calculateHaversineDistance(lat1, lon1, lat2, lon2 float64) float64 {
	const R = 6371000 // Earth radius in meters

	phi1 := lat1 * math.Pi / 180
	phi2 := lat2 * math.Pi / 180
	deltaPhi := (lat2 - lat1) * math.Pi / 180
	deltaLambda := (lon2 - lon1) * math.Pi / 180

	a := math.Sin(deltaPhi/2)*math.Sin(deltaPhi/2) +
		math.Cos(phi1)*math.Cos(phi2)*
			math.Sin(deltaLambda/2)*math.Sin(deltaLambda/2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))

	return R * c
}

func (h *MapsHandler) GetRoute(ctx context.Context, req *pb.GetRouteRequest) (*pb.RouteResponse, error) {
	// Calculate straight-line distance
	straightDistance := calculateHaversineDistance(req.OriginLat, req.OriginLng, req.DestLat, req.DestLng)

	// Mock a realistic road distance (roughly 1.3x to 1.5x the straight line due to road networks)
	roadDistance := straightDistance * 1.4

	// Assume an average city speed of 30 km/h (approx 8.3 meters/second)
	durationSeconds := int64(roadDistance / 8.3)

	// In a real application, you would return a Google Maps Encoded Polyline here.
	// For example, routing from Malviya Nagar to MI Road would return a complex string
	// that the Android osmdroid library decodes into map points.
	mockPolyline := "mock_encoded_polyline_string"

	return &pb.RouteResponse{
		DistanceMeters:  roadDistance,
		DurationSeconds: durationSeconds,
		Polyline:        mockPolyline,
	}, nil
}

func (h *MapsHandler) GetNearbyDrivers(ctx context.Context, req *pb.GetNearbyDriversRequest) (*pb.NearbyDriversResponse, error) {
	// In production, this would query PostGIS or a Redis Geospatial index (GEOSEARCH)
	// to find actual drivers marked as ONLINE in the driver-service.

	var drivers []*pb.DriverLocation

	// Mocking 3 nearby drivers hovering around the requested coordinates
	for i := 0; i < 3; i++ {
		latOffset := (rand.Float64() - 0.5) * 0.01
		lngOffset := (rand.Float64() - 0.5) * 0.01

		driverLat := req.PassengerLat + latOffset
		driverLng := req.PassengerLng + lngOffset
		dist := calculateHaversineDistance(req.PassengerLat, req.PassengerLng, driverLat, driverLng)

		if dist <= req.RadiusMeters {
			drivers = append(drivers, &pb.DriverLocation{
				DriverId:       "mock-driver-uuid-00" + string(rune(i)),
				Lat:            driverLat,
				Lng:            driverLng,
				DistanceMeters: math.Round(dist),
			})
		}
	}

	return &pb.NearbyDriversResponse{
		Drivers: drivers,
	}, nil
}
