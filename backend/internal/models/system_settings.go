package models

// SystemSettings represents configurable system-wide settings
// that are stored in a JSON file (e.g. map center and zoom).
type SystemSettings struct {
	CityName            string  `json:"city_name"`
	MapCenterLat        float64 `json:"map_center_lat"`
	MapCenterLng        float64 `json:"map_center_lng"`
	MapZoom             int     `json:"map_zoom"`
	ConfidenceThreshold float64 `json:"confidence_threshold"`
}
