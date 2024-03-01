package model

type QProduct struct {
	Name        string `json:"name" binding:"required"`
	Icon        string `json:"icon" binding:"required"`
	Description string `json:"description" binding:"required"`
}

type Product struct {
	ID int `json:"id"`
	QProduct
}

func NewProduct(id int, name, image, description string) *Product {
	return &Product{id, QProduct{name, image, description}}
}

type Products struct {
	Items []Product `json:"products"`
}
