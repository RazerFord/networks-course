package controllers

import (
	"errors"
	"net/http"
	"rest/src/model"
	"rest/src/storage"
	"strconv"

	"github.com/go-playground/validator/v10"
)

func parseErrors(err error) []string {
	var ve validator.ValidationErrors
	if errors.As(err, &ve) {
		out := make([]string, len(ve))
		for k, v := range ve {
			out[k] = v.Error()
		}
		return out[:]
	}
	return []string{err.Error()}
}

func getProductById(id string) (*model.Product, int, error) {
	if idn, err := strconv.Atoi(id); err != nil {
		return nil, http.StatusBadRequest, errors.New("wrong id format")
	} else if p := storage.GetProduct(idn); p != nil {
		return p, http.StatusOK, nil
	} else {
		return nil, http.StatusNotFound, errors.New("product not found")
	}
}

func deleteProductById(id string) (*model.Product, int, error) {
	if idn, err := strconv.Atoi(id); err != nil {
		return nil, http.StatusBadRequest, errors.New("wrong id format")
	} else if p := storage.DeleteProduct(idn); p != nil {
		return p, http.StatusOK, nil
	} else {
		return nil, http.StatusNotFound, errors.New("product not found")
	}
}
