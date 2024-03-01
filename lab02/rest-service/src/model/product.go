package model

type PostProduct struct {
	Name        string `json:"name" binding:"required"`
	Description string `json:"description" binding:"required"`
}

type Product struct {
	ID   int    `json:"id"`
	Icon string `json:"icon"`
	PostProduct
}

type ProductUpdate struct {
	ID          *int    `json:"id"`
	Icon        *string `json:"icon"`
	Name        *string `json:"name"`
	Description *string `json:"description"`
}

func NewProduct(id int, name, image, description string) *Product {
	return &Product{id, image, PostProduct{name, description}}
}

func (p *Product) Update(pu *ProductUpdate) {
	updateIfNotNull(&p.ID, pu.ID)
	updateIfNotNull(&p.Icon, pu.Icon)
	updateIfNotNull(&p.Name, pu.Name)
	updateIfNotNull(&p.Description, pu.Description)
}

func updateIfNotNull[T string | int](value *T, update *T) {
	if update != nil && value != nil {
		*value = *update
	}
}

type Products struct {
	Items []Product `json:"products"`
}
