package gormpher

import (
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"reflect"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/mitchellh/mapstructure"
	"gorm.io/gorm"
)

type onRenderFunc[T any] func(ctx *gin.Context, v *T) error
type onDeleteFunc[T any] func(ctx *gin.Context, v *T) error
type onCreateFunc[T any] func(ctx *gin.Context, v *T, vals map[string]any) error
type onUpdateFunc[T any] func(ctx *gin.Context, v *T, vals map[string]any) error

type Key interface {
	string | uint | int
}

type QueryOption struct {
	Pagination  bool
	Filterables []string
	Editables   []string
	Orderables  []string
	Searchables []string
}

func HandleGet[T any](c *gin.Context, db *gorm.DB, onRender onRenderFunc[T]) {
	key := c.Param("key")

	val, err := ExecuteGet[T](db, key)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.AbortWithError(http.StatusNotFound, err)
		} else {
			c.AbortWithError(http.StatusInternalServerError, err)
		}
		return
	}

	if onRender != nil {
		if err := onRender(c, val); err != nil {
			c.AbortWithError(http.StatusNotFound, err)
			return
		}
	}

	c.JSON(http.StatusOK, val)
}

func ExecuteGet[T any, V Key](db *gorm.DB, key V) (*T, error) {
	var val T
	result := db.Where(GetPkColumnName[T](), key).First(&val)
	if result.Error != nil {
		return nil, result.Error
	}
	return &val, nil
}

func HandleDelete[T any](c *gin.Context, db *gorm.DB, onDelete onDeleteFunc[T]) {
	key := c.Param("key")

	pkName := GetPkColumnName[T]()
	val := new(T)

	// form gorm delete hook, need to load model first
	result := db.Where(pkName, key).First(val)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusOK, true)
		} else {
			c.AbortWithError(http.StatusInternalServerError, result.Error)
		}
		return
	}

	if onDelete != nil {
		if err := onDelete(c, val); err != nil {
			c.AbortWithError(http.StatusInternalServerError, err)
			return
		}
	}

	result = db.Delete(val)
	if result.Error != nil {
		c.AbortWithError(http.StatusInternalServerError, result.Error)
		return
	}

	c.JSON(http.StatusOK, true)
}

func HandleCreate[T any](c *gin.Context, db *gorm.DB, onCreate onCreateFunc[T]) {
	var vals map[string]any
	if err := c.BindJSON(&vals); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	val := new(T)

	if err := mapstructure.Decode(vals, val); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if onCreate != nil {
		if err := onCreate(c, val, vals); err != nil {
			c.AbortWithError(http.StatusInternalServerError, err)
			return
		}
	}

	result := db.Create(val)
	if result.Error != nil {
		c.AbortWithError(http.StatusInternalServerError, result.Error)
		return
	}

	c.JSON(http.StatusOK, val)
}

func HandleEdit[T any](c *gin.Context, db *gorm.DB, editables []string, onUpdate onUpdateFunc[T]) {
	key := c.Param("key")

	var formVals map[string]any
	if err := c.BindJSON(&formVals); err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	rt := reflect.TypeOf(new(T)).Elem()

	// cannot edit primarykey
	delete(formVals, getPkJsonName(rt))

	var vals map[string]any = map[string]any{}
	for k, v := range formVals {
		field, ok := getFieldByJsonTag(rt, k)
		if !ok {
			continue
		}

		// check type
		kind := field.Type.Kind()
		if v == nil && kind != reflect.Ptr {
			continue
		}
		if v != nil && !checkType(kind, reflect.TypeOf(v).Kind()) {
			c.AbortWithError(http.StatusBadRequest, fmt.Errorf("%s type not match", field.Name))
			return
		}

		vals[field.Name] = v
	}

	if len(editables) > 0 {
		stripVals := make(map[string]any)
		for _, k := range editables {
			if v, ok := vals[k]; ok {
				// columnName, _ := getColumnNameByField(rt, k)
				// stripVals[columnName] = v
				stripVals[k] = v
			}
		}
		vals = stripVals
	} else {
		vals = map[string]any{}
	}

	if len(vals) == 0 {
		c.AbortWithError(http.StatusBadRequest, errors.New("not changed"))
		return
	}

	pkColumnName := GetPkColumnName[T]()

	if onUpdate != nil {
		val := new(T)
		if err := db.First(val, pkColumnName, key).Error; err != nil {
			c.AbortWithError(http.StatusNotFound, err)
			return
		}
		if err := onUpdate(c, val, formVals); err != nil {
			c.AbortWithError(http.StatusBadRequest, err)
			return
		}
	}

	model, err := ExecuteEdit[T](db, key, vals)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusOK, model)
}

func ExecuteEdit[T any, V Key](db *gorm.DB, key V, vals map[string]any) (*T, error) {
	var model T
	result := db.Model(&model).Where(GetPkColumnName[T](), key).Updates(vals)
	if result.Error != nil {
		return nil, result.Error
	}
	return &model, nil
}

// TODO: add onRender hook
// QueryForm: json format key
func HandleQuery[T any](c *gin.Context, db *gorm.DB, ctx *QueryOption) {
	var form QueryForm
	if err := c.BindJSON(&form); err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	if ctx == nil {
		ctx = &QueryOption{}
	}

	if ctx.Pagination {
		if form.Pos < 1 {
			form.Pos = 1
		}
	} else {
		if form.Pos < 0 {
			form.Pos = 0
		}
	}
	if form.Limit <= 0 || form.Limit > MaxQueryLimit {
		form.Limit = DefaultQueryLimit
	}

	var filterFields = make(map[string]struct{})
	for _, k := range ctx.Filterables {
		filterFields[k] = struct{}{}
	}

	rt := reflect.TypeOf(new(T)).Elem()
	if len(filterFields) > 0 {
		var stripFilters []Filter
		for i := 0; i < len(form.Filters); i++ {
			filter := form.Filters[i]
			field, ok := getFieldByJsonTag(rt, filter.Name)
			if !ok {
				continue
			}
			if _, ok := filterFields[field.Name]; !ok {
				continue
			}
			filter.Name, _ = getColumnNameByField(rt, field.Name)
			stripFilters = append(stripFilters, filter)
		}
		form.Filters = stripFilters
	} else {
		form.Filters = []Filter{}
	}

	var orderFields = make(map[string]struct{})
	for _, k := range ctx.Orderables {
		orderFields[k] = struct{}{}
	}
	if len(orderFields) > 0 {
		var stripOrders []Order
		for i := 0; i < len(form.Orders); i++ {
			order := form.Orders[i]
			field, ok := getFieldByJsonTag(rt, order.Name)
			if !ok {
				continue
			}
			if _, ok := orderFields[field.Name]; !ok {
				continue
			}
			order.Name, _ = getColumnNameByField(rt, field.Name)
			stripOrders = append(stripOrders, order)
		}
		form.Orders = stripOrders
	} else {
		form.Orders = []Order{}
	}

	if form.Keyword != "" {
		form.searchFields = ctx.Searchables
		for _, v := range form.searchFields {
			sf, _ := getColumnNameByField(rt, v)
			form.searchFields = append(form.searchFields, sf)
		}
	}

	qr := QueryResult[[]T]{
		Pos:     form.Pos,
		Limit:   form.Limit,
		Keyword: form.Keyword,
	}
	list, count, err := ExecuteQuery[T](db, form, ctx.Pagination)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	qr.Items = list
	qr.TotalCount = count

	c.JSON(http.StatusOK, qr)
}

// QueryForm: database column format key
func ExecuteQuery[T any](db *gorm.DB, form QueryForm, pagination bool) (items []T, count int, err error) {
	tableName := GetTableName[T](db)

	for _, v := range form.Filters {
		if q := v.GetQuery(); q != "" {
			db = db.Where(fmt.Sprintf("%s.%s", tableName, q), v.Value)
		}
	}

	for _, v := range form.Orders {
		if q := v.GetQuery(); q != "" {
			db = db.Order(fmt.Sprintf("%s.%s", tableName, q))
		}
	}

	if form.Keyword != "" && len(form.searchFields) > 0 {
		var query []string
		for _, v := range form.searchFields {
			query = append(query, fmt.Sprintf("`%s`.`%s` LIKE @keyword", tableName, v))
		}
		searchKey := strings.Join(query, " OR ")
		db = db.Where(searchKey, sql.Named("keyword", "%"+form.Keyword+"%"))
	}

	items = make([]T, 0, form.Limit)

	var c int64
	result := db.Model(new(T)).Count(&c)
	if result.Error != nil {
		return items, 0, result.Error
	}

	if c == 0 {
		return items, 0, nil
	}

	var offset int
	if pagination {
		offset = (form.Pos - 1) * form.Limit
	} else {
		offset = form.Pos
	}

	result = db.Offset(offset).Limit(form.Limit).Find(&items)
	return items, int(c), result.Error
}
