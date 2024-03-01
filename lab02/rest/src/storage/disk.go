package storage

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"rest/src/model"
)

const (
	Images    = "./icons/"
	database = "./data/data.json"
	prefix   = ""
	ident    = "    "
)

var storage []model.Product

func exitIfErr(err error) {
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	file, err := os.Open(database)
	defer func() {
		exitIfErr(file.Close())
	}()
	exitIfErr(err)

	bytes, err := io.ReadAll(file)
	exitIfErr(err)

	var products model.Products

	err = json.Unmarshal(bytes, &products)
	exitIfErr(err)

	storage = products.Items[:]
}

func save() {
	file, err := os.OpenFile(database, os.O_TRUNC|os.O_WRONLY, os.ModeExclusive)
	defer func() {
		exitIfErr(file.Close())
	}()
	exitIfErr(err)

	p := model.Products{Items: storage}
	bytes, err := json.MarshalIndent(p, prefix, ident)
	exitIfErr(err)

	_, err = file.Write(bytes)
	exitIfErr(err)
}

func AddProduct(qp *model.QProduct) *model.Product {
	p := model.Product{ID: len(storage) + 1, QProduct: *qp}
	storage = append(storage, p)
	return &p
}

func GetProduct(id int) *model.Product {
	if i := getIdByIndex(id); i != -1 {
		return &storage[i]
	}
	return nil
}

func DeleteProduct(id int) *model.Product {
	if i := getIdByIndex(id); i != -1 {
		tmp := &storage[i]
		storage[i], storage[len(storage)-1] = storage[len(storage)-1], storage[i]
		storage = storage[:len(storage)-1]
		return tmp
	}
	return nil
}

func GetProducts() []model.Product {
	return storage
}

func getIdByIndex(id int) int {
	for i := range storage {
		if storage[i].ID == id {
			return i
		}
	}
	return -1
}
