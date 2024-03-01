package controllers

import (
	"net/http"
	"rest/src/model"
	"rest/src/storage"

	"github.com/gin-gonic/gin"
)

func PostProduct(ctx *gin.Context) {
	var pp model.PostProduct

	if err := ctx.ShouldBindJSON(&pp); err != nil {
		ctx.IndentedJSON(http.StatusBadRequest, gin.H{"errors": parseErrors(err)})
	} else {
		p := storage.AddProduct(&pp)
		ctx.IndentedJSON(http.StatusCreated, p)
	}
}

func GetProduct(ctx *gin.Context) {
	if p, status, err := getProductById(ctx.Param("id")); err != nil {
		ctx.IndentedJSON(status, gin.H{"error": err.Error()})
	} else {
		ctx.IndentedJSON(status, *p)
	}
}

func PutProduct(ctx *gin.Context) {
	p, s, err := getProductById(ctx.Param("id"))

	if err != nil {
		ctx.IndentedJSON(s, gin.H{"error": err.Error()})
		return
	}
	var pu model.ProductUpdate

	if err := ctx.ShouldBindJSON(&pu); err != nil {
		ctx.IndentedJSON(http.StatusBadRequest, gin.H{"errors": parseErrors(err)})
	} else if existsProductByPointerId(pu.ID) {
		ctx.IndentedJSON(http.StatusBadRequest, gin.H{"error": "id is taken by another"})
	}
	p.Update(&pu)
	ctx.IndentedJSON(http.StatusOK, p)
}

func DeleteProduct(ctx *gin.Context) {
	if p, status, err := deleteProductById(ctx.Param("id")); err != nil {
		ctx.IndentedJSON(status, gin.H{"error": err.Error()})
	} else {
		ctx.IndentedJSON(status, *p)
	}
}

func GetProducts(ctx *gin.Context) {
	ctx.IndentedJSON(http.StatusOK, storage.GetProducts())
}

func UploadImageProduct(ctx *gin.Context) {
	p, s, err := getProductById(ctx.Param("id"))

	if err != nil {
		ctx.IndentedJSON(s, gin.H{"error": err.Error()})
		return
	}

	dst, err := storage.SaveIcon(ctx)

	if err != nil {
		ctx.IndentedJSON(http.StatusOK, gin.H{"error": err.Error()})
	} else {
		p.Icon = *dst
		ctx.IndentedJSON(http.StatusOK, *p)
	}
}

func DownloadImageProduct(ctx *gin.Context) {
	if p, s, err := getProductById(ctx.Param("id")); err != nil {
		ctx.IndentedJSON(s, gin.H{"error": err.Error()})
	} else if p.Icon != "" {
		ctx.File(p.Icon)
	} else {
		ctx.IndentedJSON(http.StatusNotFound, gin.H{"error": "file not found"})
	}
}
