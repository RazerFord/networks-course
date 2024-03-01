package model

type QProduct struct {
	Name        string `json:"name" binding:"required"`
	Description string `json:"description" binding:"required"`
}

type Product struct {
	ID   int    `json:"id"`
	Icon string `json:"icon"`
	QProduct
}

func NewProduct(id int, name, image, description string) *Product {
	return &Product{id, image, QProduct{name, description}}
}

type Products struct {
	Items []Product `json:"products"`
}
