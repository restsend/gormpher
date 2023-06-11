package gormpher

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"net/http"
	"path/filepath"
	"reflect"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/mitchellh/mapstructure"
	"gorm.io/gorm"
)

const (
	DefaultQueryLimit = 50
)

const (
	FilterOpEqual          = "="
	FilterOpNotEqual       = "<>"
	FilterOpIn             = "in"
	FilterOpNotIn          = "not_in"
	FilterOpGreater        = ">"
	FilterOpGreaterOrEqual = ">="
	FilterOpLess           = "<"
	FilterOpLessOrEqual    = "<="
	FilterOpLike           = "like"
)

const (
	OrderOpDesc = "desc"
	OrderOpAsc  = "asc"
)

const (
	GET    = 1 << 1
	CREATE = 1 << 2
	EDIT   = 1 << 3
	DELETE = 1 << 4
	QUERY  = 1 << 5
	BATCH  = 1 << 6
)

type GetDB func(c *gin.Context, isCreate bool) *gorm.DB // designed for group
type PrepareQuery func(db *gorm.DB, c *gin.Context) (*gorm.DB, *QueryForm, error)

type (
	CreateFunc func(ctx *gin.Context, vptr any, vals map[string]any) error
	DeleteFunc func(ctx *gin.Context, vptr any) error
	UpdateFunc func(ctx *gin.Context, vptr any, vals map[string]any) error
	RenderFunc func(ctx *gin.Context, vptr any) error
)

type QueryView struct {
	Name    string
	Method  string
	Prepare PrepareQuery
}

type WebObject struct {
	Model       any
	Group       string
	Name        string
	Editables   []string
	Filterables []string
	Orderables  []string
	Searchables []string
	GetDB       GetDB
	OnCreate    CreateFunc
	OnUpdate    UpdateFunc
	OnDelete    DeleteFunc
	OnRender    RenderFunc

	Views        []QueryView
	AllowMethods int

	PrimaryKeyName     string
	PrimaryKeyJsonName string
	tableName          string

	// Model type
	modelElem reflect.Type
	// Map json tag to struct field name. such as:
	// UUID string `json:"id"` => {"id" : "UUID"}
	jsonToFields map[string]string
	// Map json tag to field kind. such as:
	// UUID string `json:"id"` => {"id": string}
	jsonToKinds map[string]reflect.Kind
}

type Filter struct {
	Name  string `json:"name"`
	Op    string `json:"op"`
	Value any    `json:"value"`
}

type Order struct {
	Name string `json:"name"`
	Op   string `json:"op"`
}

type QueryForm struct {
	Pos          int      `json:"pos"`
	Limit        int      `json:"limit"`
	Keyword      string   `json:"keyword,omitempty"`
	Filters      []Filter `json:"filters,omitempty"`
	Orders       []Order  `json:"orders,omitempty"`
	ViewFields   []string `json:"-"` // for view
	searchFields []string `json:"-"` // for keyword
}

type QueryResult[T any] struct {
	TotalCount int    `json:"total,omitempty"`
	Pos        int    `json:"pos,omitempty"`
	Limit      int    `json:"limit,omitempty"`
	Keyword    string `json:"keyword,omitempty"`
	Items      T      `json:"items"`
}

// GetQuery return the combined filter SQL statement.
// such as "age >= ?", "name IN ?".
func (f *Filter) GetQuery() string {
	var op string
	switch f.Op {
	case FilterOpEqual:
		op = "="
	case FilterOpNotEqual:
		op = "<>"
	case FilterOpIn:
		op = "IN"
	case FilterOpNotIn:
		op = "NOT IN"
	case FilterOpGreater:
		op = ">"
	case FilterOpGreaterOrEqual:
		op = ">="
	case FilterOpLess:
		op = "<"
	case FilterOpLessOrEqual:
		op = "<="
	case FilterOpLike:
		op = "LIKE"
	}

	if op == "" {
		return ""
	}

	return fmt.Sprintf("%s %s ?", f.Name, op)
}

// GetQuery return the combined order SQL statement.
// such as "id DESC".
func (f *Order) GetQuery() string {
	if f.Op == OrderOpDesc {
		return f.Name + " DESC"
	}
	return f.Name + " ASC"
}

func (obj *WebObject) RegisterObject(r gin.IRoutes) error {
	if err := obj.Build(); err != nil {
		return err
	}

	p := filepath.Join(obj.Group, obj.Name)
	allowMethods := obj.AllowMethods
	if allowMethods == 0 {
		allowMethods = GET | CREATE | EDIT | DELETE | QUERY | BATCH
	}

	if allowMethods&GET != 0 {
		r.GET(filepath.Join(p, ":key"), func(c *gin.Context) {
			handleGetObject(c, obj)
		})
	}
	if allowMethods&CREATE != 0 {
		r.PUT(p, func(c *gin.Context) {
			handleCreateObject(c, obj)
		})
	}
	if allowMethods&EDIT != 0 {
		r.PATCH(filepath.Join(p, ":key"), func(c *gin.Context) {
			handleEditObject(c, obj)
		})
	}

	if allowMethods&DELETE != 0 {
		r.DELETE(filepath.Join(p, ":key"), func(c *gin.Context) {
			handleDeleteObject(c, obj)
		})
	}

	if allowMethods&QUERY != 0 {
		r.POST(p, func(c *gin.Context) {
			handleQueryObject(c, obj, DefaultPrepareQuery)
		})
	}

	if allowMethods&BATCH != 0 {
		r.DELETE(p, func(c *gin.Context) {
			handleBatchDelete(c, obj)
		})
	}

	for i := 0; i < len(obj.Views); i++ {
		v := &obj.Views[i]
		if v.Name == "" {
			return errors.New("with invalid view")
		}
		if v.Method == "" {
			v.Method = http.MethodPost
		}
		if v.Prepare == nil {
			v.Prepare = DefaultPrepareQuery
		}
		r.Handle(v.Method, filepath.Join(p, v.Name), func(ctx *gin.Context) {
			handleQueryObject(ctx, obj, v.Prepare)
		})
	}

	return nil
}

func RegisterObject(r gin.IRoutes, obj *WebObject) error {
	return obj.RegisterObject(r)
}

func RegisterObjects(r gin.IRoutes, objs []WebObject) {
	for idx := range objs {
		obj := &objs[idx]
		err := obj.RegisterObject(r)
		if err != nil {
			log.Fatalf("RegisterObject [%s] fail %v\n", obj.Name, err)
		}
	}
}

// Build fill the properties of obj.
func (obj *WebObject) Build() error {
	rt := reflect.TypeOf(obj.Model)
	if rt.Kind() == reflect.Ptr {
		rt = rt.Elem()
	}
	obj.modelElem = rt

	// TODO: optimize
	obj.tableName = obj.modelElem.Name()

	if obj.Name == "" {
		obj.Name = strings.ToLower(obj.tableName)
	}

	obj.jsonToFields = make(map[string]string)
	obj.jsonToKinds = make(map[string]reflect.Kind)
	obj.parseFields(obj.modelElem)

	if obj.PrimaryKeyName == "" {
		return fmt.Errorf("%s not has primaryKey", obj.Name)
	}

	if obj.GetDB == nil {
		return fmt.Errorf("without db")
	}

	return nil
}

// parseFields parse the following properties according to struct tag:
// - jsonToFields, jsonToKinds, primaryKeyName, primaryKeyJsonName
func (obj *WebObject) parseFields(rt reflect.Type) {
	for i := 0; i < rt.NumField(); i++ {
		f := rt.Field(i)

		if f.Anonymous && f.Type.Kind() == reflect.Struct {
			obj.parseFields(f.Type)
		}

		jsonTag := f.Tag.Get("json")
		if jsonTag == "" {
			jsonTag = f.Name
		}

		if jsonTag != "-" {
			obj.jsonToFields[jsonTag] = f.Name
			kind := f.Type.Kind()
			if kind == reflect.Ptr {
				kind = f.Type.Elem().Kind()
			}
			obj.jsonToKinds[jsonTag] = kind
		}

		gormTag := f.Tag.Get("gorm")
		if gormTag == "" || gormTag == "-" {
			continue
		}

		if !strings.Contains(gormTag, "primarykey") &&
			!strings.Contains(gormTag, "primaryKey") {
			continue
		}

		obj.PrimaryKeyName = f.Name
		if jsonTag == "-" || jsonTag == "" {
			obj.PrimaryKeyJsonName = f.Name
		} else {
			obj.PrimaryKeyJsonName = jsonTag
		}
	}
}

func handleGetObject(c *gin.Context, obj *WebObject) {
	key := c.Param("key")
	db := obj.GetDB(c, false)

	// the real name of the primaryKey column
	pkColName := db.NamingStrategy.ColumnName(obj.tableName, obj.PrimaryKeyName)

	val := reflect.New(obj.modelElem).Interface()
	result := db.Where(pkColName, key).Take(&val)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": "not found"})
		} else {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": result.Error.Error()})
		}
		return
	}

	if obj.OnRender != nil {
		if err := obj.OnRender(c, val); err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}

	c.JSON(http.StatusOK, val)
}

func handleCreateObject(c *gin.Context, obj *WebObject) {
	var vals map[string]any
	if err := c.BindJSON(&vals); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	val := reflect.New(obj.modelElem).Interface()

	if err := mapstructure.Decode(vals, val); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if obj.OnCreate != nil {
		if err := obj.OnCreate(c, val, vals); err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
	}

	result := obj.GetDB(c, true).Create(val)
	if result.Error != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": result.Error.Error()})
		return
	}

	c.JSON(http.StatusOK, val)
}

func handleEditObject(c *gin.Context, obj *WebObject) {
	key := c.Param("key")

	var inputVals map[string]any
	if err := c.BindJSON(&inputVals); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	db := obj.GetDB(c, false)

	var vals map[string]any = map[string]any{}

	// can't edit primaryKey
	delete(inputVals, obj.PrimaryKeyJsonName)

	for k, v := range inputVals {
		if v == nil {
			continue
		}
		// Check the kind to be edited.
		kind, ok := obj.jsonToKinds[k]
		if !ok {
			continue
		}

		fname, ok := obj.jsonToFields[k]
		if !ok {
			continue
		}

		if !checkType(kind, reflect.TypeOf(v).Kind()) {
			c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("%s type not match", fname)})
			return
		}

		vals[fname] = v
	}

	if len(obj.Editables) > 0 {
		stripVals := make(map[string]any)
		for _, k := range obj.Editables {
			if v, ok := vals[k]; ok {
				stripVals[k] = v
			}
		}
		vals = stripVals
	} else {
		vals = map[string]any{}
	}

	if len(vals) == 0 {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "not changed"})
		return
	}

	pkColName := db.NamingStrategy.ColumnName(obj.tableName, obj.PrimaryKeyName)

	if obj.OnUpdate != nil {
		val := reflect.New(obj.modelElem).Interface()
		if err := db.First(val, pkColName, key).Error; err != nil {
			c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": "not found"})
			return
		}
		if err := obj.OnUpdate(c, val, inputVals); err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
	}

	model := reflect.New(obj.modelElem).Interface()
	result := db.Model(model).Where(pkColName, key).Updates(vals)
	if result.Error != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": result.Error.Error()})
		return
	}

	c.JSON(http.StatusOK, true)
}

func handleDeleteObject(c *gin.Context, obj *WebObject) {
	key := c.Param("key")
	db := obj.GetDB(c, false)

	pkColName := db.NamingStrategy.ColumnName(obj.tableName, obj.PrimaryKeyName)
	val := reflect.New(obj.modelElem).Interface()

	r := db.First(val, pkColName, key)

	// for gorm delete hook, need to load model first.
	if r.Error != nil {
		if errors.Is(r.Error, gorm.ErrRecordNotFound) {
			c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": "not found"})
		} else {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": r.Error.Error()})
		}
		return
	}

	if obj.OnDelete != nil {
		if err := obj.OnDelete(c, val); err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
	}

	r = db.Delete(val)
	if r.Error != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": r.Error.Error()})
		return
	}

	c.JSON(http.StatusOK, true)
}

func handleBatchDelete(c *gin.Context, obj *WebObject) {
	var form []string
	if err := c.BindJSON(&form); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	db := obj.GetDB(c, false)

	val := reflect.New(obj.modelElem).Interface()
	r := db.Delete(&val, form)
	if r.Error != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": r.Error.Error()})
		return
	}

	c.JSON(http.StatusOK, true)
}

func handleQueryObject(c *gin.Context, obj *WebObject, prepareQuery PrepareQuery) {
	db, form, err := prepareQuery(obj.GetDB(c, false), c)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	namer := db.NamingStrategy

	// Use struct{} makes map like set.
	var filterFields = make(map[string]struct{})
	for _, k := range obj.Filterables {
		filterFields[k] = struct{}{}
	}
	if len(filterFields) > 0 {
		var stripFilters []Filter
		for i := 0; i < len(form.Filters); i++ {
			filter := form.Filters[i]
			// Struct must has this field.
			field, ok := obj.jsonToFields[filter.Name]
			if !ok {
				continue
			}
			if _, ok := filterFields[field]; !ok {
				continue
			}
			filter.Name = namer.ColumnName(obj.tableName, filter.Name)
			stripFilters = append(stripFilters, filter)
		}
		form.Filters = stripFilters
	} else {
		form.Filters = []Filter{}
	}

	var orderFields = make(map[string]struct{})
	for _, k := range obj.Orderables {
		orderFields[k] = struct{}{}
	}
	if len(orderFields) > 0 {
		var stripOrders []Order
		for i := 0; i < len(form.Orders); i++ {
			order := form.Orders[i]
			field, ok := obj.jsonToFields[order.Name]
			if !ok {
				continue
			}
			if _, ok := orderFields[field]; !ok {
				continue
			}
			order.Name = namer.ColumnName(obj.tableName, order.Name)
			stripOrders = append(stripOrders, order)
		}
		form.Orders = stripOrders
	} else {
		form.Orders = []Order{}
	}

	if form.Keyword != "" {
		form.searchFields = []string{}
		for _, v := range obj.Searchables {
			form.searchFields = append(form.searchFields, namer.ColumnName(obj.tableName, v))
		}
	}

	if len(form.ViewFields) > 0 {
		var stripViewFields []string
		for _, v := range form.ViewFields {
			stripViewFields = append(stripViewFields, namer.ColumnName(obj.tableName, v))
		}
		form.ViewFields = stripViewFields
	}

	r, err := QueryObjects(db, obj, form)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if obj.OnRender != nil {
		vals := reflect.ValueOf(r.Items)
		if vals.Kind() == reflect.Slice {
			for i := 0; i < vals.Len(); i++ {
				v := vals.Index(i).Addr().Interface()
				if err := obj.OnRender(c, v); err != nil {
					c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
					return
				}
				vals.Index(i).Set(reflect.ValueOf(v).Elem())
			}
		}
	}

	c.JSON(http.StatusOK, r)
}

// QueryObjects execute query and return data.
func QueryObjects(db *gorm.DB, obj *WebObject, form *QueryForm) (r QueryResult[any], err error) {
	// the real name of the db table
	tblName := db.NamingStrategy.TableName(obj.tableName)

	// TODO:
	for _, v := range form.Filters {
		if q := v.GetQuery(); q != "" {
			db = db.Where(fmt.Sprintf("%s.%s", tblName, q), v.Value)
		}
	}

	for _, v := range form.Orders {
		if q := v.GetQuery(); q != "" {
			db = db.Order(fmt.Sprintf("%s.%s", tblName, q))
		}
	}

	if form.Keyword != "" && len(form.searchFields) > 0 {
		var query []string
		for _, v := range form.searchFields {
			query = append(query, fmt.Sprintf("`%s`.`%s` LIKE @keyword", tblName, v))
		}
		searchKey := strings.Join(query, " OR ")
		db = db.Where(searchKey, sql.Named("keyword", "%"+form.Keyword+"%"))
	}

	if len(form.ViewFields) > 0 {
		db = db.Select(form.ViewFields)
	}

	r.Pos = form.Pos
	r.Limit = form.Limit
	r.Keyword = form.Keyword

	var c int64
	model := reflect.New(obj.modelElem).Interface()
	if err := db.Model(model).Count(&c).Error; err != nil {
		return r, err
	}
	if c <= 0 {
		return r, nil
	}
	r.TotalCount = int(c)

	items := reflect.New(reflect.SliceOf(obj.modelElem))
	result := db.Offset(form.Pos).Limit(form.Limit).Find(items.Interface())
	if result.Error != nil {
		return r, result.Error
	}
	r.Items = items.Elem().Interface()
	r.Pos += int(result.RowsAffected)
	return r, nil
}

// DefaultPrepareQuery return default QueryForm.
func DefaultPrepareQuery(db *gorm.DB, c *gin.Context) (*gorm.DB, *QueryForm, error) {
	var form QueryForm
	if c.Request.ContentLength > 0 {
		if err := c.BindJSON(&form); err != nil {
			return nil, nil, err
		}
	}

	if form.Pos < 0 {
		form.Pos = 0
	}
	if form.Limit <= 0 || form.Limit > 150 {
		form.Limit = DefaultQueryLimit
	}

	return db, &form, nil
}

/*
Check Go type corresponds to JSON type.
- float64, for JSON numbers
- string, for JSON strings
- []any, for JSON arrays
- map[string]any, for JSON objects
- nil, for JSON null
*/
func checkType(goKind, jsonKind reflect.Kind) bool {
	switch goKind {
	case reflect.Struct, reflect.Slice: // time.Time, associated structures
		return true
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
		reflect.Float32, reflect.Float64:
		return jsonKind == reflect.Float64
	default:
		return goKind == jsonKind
	}
}
