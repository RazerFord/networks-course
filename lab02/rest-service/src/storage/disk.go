package storage

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"rest/src/model"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

const (
	Images   = "./icons/"
	database = "./data/data.json"
	prefix   = ""
	ident    = "    "
)

var (
	storage []model.Product
	lastId  = 0
)

func exitIfErr(err error) {
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	initStorage()
	initLastId()
}

func initStorage() {
	bytes, err := readFile(database)
	exitIfErr(err)

	var products model.Products

	err = json.Unmarshal(bytes, &products)
	exitIfErr(err)

	storage = products.Items[:]
}

func initLastId() {
	for _, v := range storage {
		lastId = max(v.ID, lastId)
	}
}

func readFile(filename string) ([]byte, error) {
	file, err := os.Open(database)
	defer func() error {
		return file.Close()
	}()

	if err != nil {
		return nil, err
	}

	return io.ReadAll(file)
}

func AddProduct(qp *model.PostProduct) *model.Product {
	lastId++
	p := model.Product{ID: lastId, PostProduct: *qp}
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
		storage[i], storage[len(storage)-1] = storage[len(storage)-1], storage[i]
		tmp := &storage[len(storage)-1]
		storage = storage[:len(storage)-1]
		return tmp
	}
	return nil
}

func GetProducts() []model.Product {
	return storage
}

func SaveIcon(ctx *gin.Context) (*string, error) {
	f, err := ctx.FormFile("icon")

	if err != nil {
		return nil, err
	}

	dst := Images + strconv.Itoa(time.Now().Nanosecond()) + f.Filename
	return &dst, ctx.SaveUploadedFile(f, dst)
}

func getIdByIndex(id int) int {
	for i := range storage {
		if storage[i].ID == id {
			return i
		}
	}
	return -1
}
