package models

type Report struct {
	TotalRevenue      int        `json:"total_revenue"`
	TotalTransactions int        `json:"total_transactions"`
	TopProduct        TopProduct `json:"top_product"`
}

type TopProduct struct {
	ProductName  string `json:"product_name"`
	QuantitySold int    `json:"quantity_sold"`
}
