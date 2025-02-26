package models

type Sector struct {
	BusinessDate     string  `json:"businessDate"`
	SectorName       string  `json:"sectorName"`
	TotalTransaction float64 `json:"totalTransaction"`
	TurnOverValues   float64 `json:"turnOverValues"`
	TurnOverVolume   int     `json:"turnOverVolume"`
}
