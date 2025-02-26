package models

type Company struct {
	CompanyEmail   string `json:"companyEmail"`
	CompanyName    string `json:"companyName"`
	Id             int    `json:"id"`
	InstrumentType string `json:"instrumentType"`
	RegulatoryBody string `json:"regulatoryBody"`
	SectorName     string `json:"sectorName"`
	SecurityName   string `json:"securityName"`
	Status         string `json:"status"`
	Symbol         string `json:"symbol"`
	Website        string `json:"website"`
}
