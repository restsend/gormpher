package gormpher

import (
	"errors"
	"net/http"
	"reflect"

	"github.com/gin-gonic/gin"
	"github.com/mitchellh/mapstructure"
	"gorm.io/gorm"
)

func HandleGet(c *gin.Context, db *gorm.DB, rt reflect.Type, onRender RenderFunc) {
	key := c.Param("key")

	val := reflect.New(rt).Interface()
	result := db.Where(getPkColumnName(rt), key).Take(&val)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": "not found"})
		} else {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": result.Error.Error()})
		}
		return
	}

	if onRender != nil {
		if err := onRender(c, val); err != nil {
			c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
	}

	c.JSON(http.StatusOK, val)
}

func HandleCreate(c *gin.Context, db *gorm.DB, rt reflect.Type, onCreate CreateFunc) {
	var vals map[string]any
	if err := c.BindJSON(&vals); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	val := reflect.New(rt).Interface()

	if err := mapstructure.Decode(vals, val); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if onCreate != nil {
		if err := onCreate(c, val, vals); err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}

	result := db.Create(val)
	if result.Error != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": result.Error.Error()})
		return
	}

	c.JSON(http.StatusOK, val)
}

func HandleDelete(c *gin.Context, db *gorm.DB, rt reflect.Type, onDelete DeleteFunc) {
	key := c.Param("key")
	val := reflect.New(rt).Interface()

	result := db.First(val, getPkColumnName(rt), key)

	// for gorm delete hook, need to load model first.
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			// c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": "not found"})
			c.JSON(http.StatusOK, true)
		} else {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": result.Error.Error()})
		}
		return
	}

	if onDelete != nil {
		if err := onDelete(c, val); err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}

	result = db.Delete(val)
	if result.Error != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": result.Error.Error()})
		return
	}

	c.JSON(http.StatusOK, true)
}

func HandleEdit() {}

func HandleQuery() {}
