package model

type QProduct struct {
	Name        string `json:"name" binding:"required"`
	Description string `json:"description" binding:"required"`
}

type Product struct {
	ID int `json:"id"`
	QProduct
}

func NewProduct(id int, name, description string) *Product {
	return &Product{id, QProduct{name, description}}
}

type Products struct {
	Items []Product `json:"products"`
}
