package controllers

import (
	"mime/multipart"
	"net/http"
	"rest/src/model"
	"rest/src/storage"

	"github.com/gin-gonic/gin"
)

func PostProduct(ctx *gin.Context) {
	var qp model.QProduct

	if err := ctx.ShouldBindJSON(&qp); err != nil {
		ctx.IndentedJSON(http.StatusBadRequest, gin.H{"errors": parseErrors(err)})
	} else {
		p := storage.AddProduct(&qp)
		ctx.IndentedJSON(http.StatusCreated, p)
	}
}

func GetProduct(ctx *gin.Context) {
	id := ctx.Param("id")

	if p, status, err := getProductById(id); err != nil {
		ctx.IndentedJSON(status, gin.H{"error": err.Error()})
	} else {
		ctx.IndentedJSON(status, *p)
	}
}

func PutProduct(ctx *gin.Context) {
	id := ctx.Param("id")

	p, status, err := getProductById(id)

	if err != nil {
		ctx.IndentedJSON(status, gin.H{"error": err.Error()})
		return
	}

	var pp struct {
		ID          *int    `json:"id"`
		Name        *string `json:"name"`
		Description *string `json:"description"`
	}

	if err := ctx.Bind(&pp); err != nil {
		ctx.IndentedJSON(http.StatusBadRequest, gin.H{"errors": parseErrors(err)})
	} else if pp.ID != nil {
		if tmp := storage.GetProduct(*pp.ID); tmp != nil {
			ctx.IndentedJSON(status, gin.H{"error": "id is taken by another"})
			return
		}
		p.ID = *pp.ID
	} else if pp.Name != nil {
		p.Name = *pp.Name
	} else if pp.Description != nil {
		p.Description = *pp.Description
	}
	ctx.IndentedJSON(http.StatusOK, p)
}

func DeleteProduct(ctx *gin.Context) {
	id := ctx.Param("id")

	if p, status, err := deleteProductById(id); err != nil {
		ctx.IndentedJSON(status, gin.H{"error": err.Error()})
	} else {
		ctx.IndentedJSON(status, *p)
	}
}

func GetProducts(ctx *gin.Context) {
	ctx.IndentedJSON(http.StatusOK, storage.GetProducts())
}

func UploadImageProduct(ctx *gin.Context) {
	id := ctx.Param("id")

	var p *model.Product
	var f *multipart.FileHeader
	var s int
	var err error

	if p, s, err = getProductById(id); err != nil {
		ctx.IndentedJSON(s, gin.H{"error": err.Error()})
		return
	} else if f, err = ctx.FormFile("icon"); err != nil {
		ctx.IndentedJSON(s, gin.H{"error": err.Error()})
		return
	}

	dst := storage.Images + f.Filename
	err = ctx.SaveUploadedFile(f, dst)

	if err != nil {
		ctx.IndentedJSON(http.StatusOK, gin.H{"error": err.Error()})
	} else {
		p.Icon = dst
		ctx.IndentedJSON(http.StatusOK, *p)
	}
}

func DownloadImageProduct(ctx *gin.Context) {
	id := ctx.Param("id")

	if p, s, err := getProductById(id); err != nil {
		ctx.IndentedJSON(s, gin.H{"error": err.Error()})
		return
	} else {
		ctx.File(p.Icon)
	}
}
