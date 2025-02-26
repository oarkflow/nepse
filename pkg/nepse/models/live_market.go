package models

type LiveFloorsheet struct {
	Floorsheets struct {
		Content       []Transaction `json:"content,omitempty"`
		Empty         bool          `json:"empty,omitempty"`
		First         bool          `json:"first,omitempty"`
		Last          bool          `json:"last,omitempty"`
		Number        int           `json:"number,omitempty"`
		NumberOfItems int           `json:"numberOfElements,omitempty"`
		TotalPages    int           `json:"totalPages,omitempty"`
	} `json:"floorsheets,omitempty"`
	Content     []map[string]any `json:"content"`
	TotalAmount float64          `json:"totalAmount"`
	TotalQty    float64          `json:"totalQty"`
	TotalTrades int              `json:"totalTrades"`
}
