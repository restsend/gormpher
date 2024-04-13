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
	"time"

	"github.com/gin-gonic/gin"
	"github.com/mitchellh/mapstructure"
	"gorm.io/gorm"
)

const (
	DefaultQueryLimit = 102400 // 100k
	MaxQueryLimit     = 150
)

// Request method
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
	BeforeCreateFunc      func(ctx *gin.Context, vptr any, vals map[string]any) error
	BeforeDeleteFunc      func(ctx *gin.Context, vptr any) error
	BeforeUpdateFunc      func(ctx *gin.Context, vptr any, vals map[string]any) error
	BeforeRenderFunc      func(ctx *gin.Context, vptr any) (any, error)
	BeforeQueryRenderFunc func(db *gorm.DB, ctx *gin.Context, r *QueryResult) (any, error)
)

type QueryView struct {
	Path    string `json:"path"`
	Method  string `json:"method"`
	Desc    string `json:"desc"`
	Prepare PrepareQuery
}

type WebObjectPrimaryField struct {
	IsPrimary bool
	Name      string
	Kind      reflect.Kind
	JSONName  string
}

type WebObject struct {
	Model any
	Group string
	Name  string
	Desc  string

	GetDB        GetDB
	PrepareQuery PrepareQuery

	// config
	AllowMethods int

	// for query
	EditFields   []string
	FilterFields []string
	OrderFields  []string
	SearchFields []string
	Views        []QueryView

	// hooks
	BeforeCreate      BeforeCreateFunc
	BeforeUpdate      BeforeUpdateFunc
	BeforeDelete      BeforeDeleteFunc
	BeforeRender      BeforeRenderFunc
	BeforeQUeryRender BeforeQueryRenderFunc

	jsonPKName string
	gormPKName string
	preloads   []string // for gorm preload

	primaryKeys []WebObjectPrimaryField
	uniqueKeys  []WebObjectPrimaryField
	tableName   string

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
	isTimeType bool   `json:"-"`
	Name       string `json:"name"`
	Op         string `json:"op"`
	Value      any    `json:"value"`
}

type Order struct {
	Name string `json:"name"`
	Op   string `json:"op"`
}

type QueryForm struct {
	Pagination bool `json:"pagination"`

	Pos          int      `json:"pos"`
	Limit        int      `json:"limit"`
	Keyword      string   `json:"keyword,omitempty"`
	Filters      []Filter `json:"filters,omitempty"`
	Orders       []Order  `json:"orders,omitempty"`
	ViewFields   []string `json:"-"` // for view
	searchFields []string `json:"-"` // for keyword
}

type QueryResult struct {
	Total   int    `json:"total,omitempty"`
	Pos     int    `json:"pos,omitempty"`
	Limit   int    `json:"limit,omitempty"`
	Keyword string `json:"keyword,omitempty"`
	Items   []any  `json:"items"`
}

// GetQuery return the combined filter SQL statement.
// such as "age >= ?", "name IN ?".
func (f *Filter) GetQuery() string {
	var op string
	switch f.Op {
	case "in", "IN":
		op = "IN"
	case "not_in", "NOT_IN":
		op = "NOT IN"
	case "like", "LIKE":
		op = "LIKE"
	case "=", "equal", "EQUAL":
		op = "="
	case "<>", "not_equal", "NOT_EQUAL", "!=":
		op = "<>"
	case ">", "greater", "GREATER":
		op = ">"
	case "greater_or_equal", "GREATER_OR_EQUAL", ">=":
		op = ">="
	case "<", "less", "LESS":
		op = "<"
	case "less_or_equal", "LESS_OR_EQUAL", "<=":
		op = "<="
	case "between", "BETWEEN": // for time range
		op = "BETWEEN"
		return fmt.Sprintf("`%s` BETWEEN ? AND ?", f.Name)
	}

	if op == "" {
		return ""
	}

	return fmt.Sprintf("`%s` %s ?", f.Name, op)
}

// GetQuery return the combined order SQL statement.
// such as "id DESC".
func (o *Order) GetQuery() string {
	switch o.Op {
	case "desc", "DESC":
		return o.Name + " DESC"
	default:
		return o.Name + " ASC"
	}
}

func (obj *WebObject) RegisterObject(r gin.IRoutes) error {
	if err := obj.Build(); err != nil {
		return err
	}

	// p := filepath.Join(obj.Group, obj.Name)
	p := obj.Name
	allowMethods := obj.AllowMethods
	if allowMethods == 0 {
		allowMethods = GET | CREATE | EDIT | DELETE | QUERY | BATCH
	}

	primaryKeyPath := obj.BuildPrimaryPath(p)
	if allowMethods&GET != 0 {
		r.GET(primaryKeyPath, func(c *gin.Context) {
			handleGetObject(c, obj)
		})
	}
	if allowMethods&CREATE != 0 {
		r.PUT(p, func(c *gin.Context) {
			handleCreateObject(c, obj)
		})
	}
	if allowMethods&EDIT != 0 {
		r.PATCH(primaryKeyPath, func(c *gin.Context) {
			handleUpdateObject(c, obj)
		})
	}
	if allowMethods&DELETE != 0 {
		r.DELETE(primaryKeyPath, func(c *gin.Context) {
			handleDeleteObject(c, obj)
		})
	}

	if allowMethods&QUERY != 0 {
		r.POST(p, func(c *gin.Context) {
			handleQueryObject(c, obj, obj.PrepareQuery)
		})
	}

	// if allowMethods&BATCH != 0 {
	// 	r.DELETE(p, func(c *gin.Context) {
	// 		handleBatchDelete(c, obj)
	// 	})
	// }

	for i := 0; i < len(obj.Views); i++ {
		v := &obj.Views[i]
		if v.Path == "" {
			return errors.New("with invalid view")
		}
		if v.Method == "" {
			v.Method = http.MethodPost
		}
		if v.Prepare == nil {
			v.Prepare = DefaultPrepareQuery
		}
		r.Handle(v.Method, filepath.Join(p, v.Path), func(ctx *gin.Context) {
			handleQueryObject(ctx, obj, v.Prepare)
		})
	}

	return nil
}

func (obj *WebObject) BuildPrimaryPath(prefix string) string {
	var primaryKeyPath []string
	for _, v := range obj.uniqueKeys {
		primaryKeyPath = append(primaryKeyPath, ":"+v.JSONName)
	}
	return filepath.Join(prefix, filepath.Join(primaryKeyPath...))
}

func (obj *WebObject) getPrimaryValues(c *gin.Context) ([]string, error) {
	var result []string
	for _, field := range obj.uniqueKeys {
		v := c.Param(field.JSONName)
		if v == "" {
			return nil, fmt.Errorf("invalid primary: %s", field.JSONName)
		}
		result = append(result, v)
	}
	return result, nil
}

func (obj *WebObject) buildPrimaryCondition(db *gorm.DB, keys []string) *gorm.DB {
	var tx *gorm.DB
	for i := 0; i < len(obj.uniqueKeys); i++ {
		colName := obj.uniqueKeys[i].Name
		col := db.NamingStrategy.ColumnName(obj.tableName, colName)
		tx = db.Where(col, keys[i])
	}
	return tx
}

func RegisterObject(r *gin.RouterGroup, obj *WebObject) error {
	return obj.RegisterObject(r)
}

func RegisterObjects(r *gin.RouterGroup, objs []WebObject) {
	for idx := range objs {
		obj := &objs[idx]
		if err := obj.RegisterObject(r); err != nil {
			log.Fatalf("RegisterObject [%s] fail %v\n", obj.Name, err)
		}
	}
}

// Build fill the properties of obj.
func (obj *WebObject) Build() error {
	// if obj.GetDB == nil {
	// 	return fmt.Errorf("without db")
	// }

	rt := reflect.TypeOf(obj.Model)
	if rt.Kind() == reflect.Ptr {
		rt = rt.Elem()
	}

	obj.modelElem = rt
	obj.tableName = obj.modelElem.Name()

	if obj.Name == "" {
		obj.Name = strings.ToLower(obj.tableName)
	}

	obj.jsonToFields = make(map[string]string)
	obj.jsonToKinds = make(map[string]reflect.Kind)
	obj.parseFields(obj.modelElem)

	if obj.primaryKeys != nil {
		obj.uniqueKeys = obj.primaryKeys
	}

	if len(obj.uniqueKeys) == 0 && len(obj.primaryKeys) == 0 {
		return fmt.Errorf("%s not has primaryKey", obj.Name)
	}

	// obj.gormPKName = getPkColumnName(obj.modelElem)
	// if obj.gormPKName == "" {
	// 	return fmt.Errorf("%s not has primary key", obj.Name)
	// }

	return nil
}

// parseFields parse the following properties according to struct tag:
// - jsonToFields, jsonToKinds, primaryKeyName, primaryKeyJsonName
func (obj *WebObject) parseFields(rt reflect.Type) {
	for i := 0; i < rt.NumField(); i++ {
		f := rt.Field(i)

		if f.Anonymous && f.Type.Kind() == reflect.Struct {
			obj.parseFields(f.Type)
			continue
		}

		jsonTag := f.Tag.Get("json")
		if strings.Contains(jsonTag, ",") {
			jsonTag = strings.TrimSpace(strings.Split(jsonTag, ",")[0])
		}

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

		gormTag := strings.ToLower(f.Tag.Get("gorm"))
		if gormTag == "" || gormTag == "-" {
			continue
		}

		pkField := WebObjectPrimaryField{
			Name:      f.Name,
			JSONName:  strings.Split(jsonTag, ",")[0],
			Kind:      f.Type.Kind(),
			IsPrimary: strings.Contains(gormTag, "primaryKey"),
		}

		if pkField.JSONName == "" {
			pkField.JSONName = pkField.Name
		}

		if pkField.IsPrimary {
			obj.primaryKeys = append(obj.primaryKeys, pkField)
		} else if strings.Contains(gormTag, "unique") {
			obj.uniqueKeys = append(obj.uniqueKeys, pkField)
		}

		// TODO: how to decide whether to preload?
		// 	if strings.Contains(gormTag, "foreignKey") ||
		// 		strings.Contains(gormTag, "references") ||
		// 		strings.Contains(gormTag, "many2many") {

		// 		exist := false
		// 		for _, preload := range obj.preloads {
		// 			if preload == f.Name {
		// 				exist = true
		// 			}
		// 		}

		// 		if !exist {
		// 			obj.preloads = append(obj.preloads, f.Name)
		// 		}
		// 	}

		// 	if strings.Contains(strings.ToLower(gormTag), "primarykey") {
		// 		if jsonTag == "-" || jsonTag == "" {
		// 			obj.jsonPKName = f.Name
		// 		} else {
		// 			obj.jsonPKName = jsonTag
		// 		}
		// 	}
		// }
	}
}

func getDbConnection(c *gin.Context, objFn GetDB, isCreate bool) (tx *gorm.DB) {
	if objFn != nil {
		tx = objFn(c, isCreate)
	} else {
		panic("without db")
	}
	return tx.Session(&gorm.Session{})
}

func handleGetObject(c *gin.Context, obj *WebObject) {
	keys, err := obj.getPrimaryValues(c)
	if err != nil {
		handleError(c, http.StatusBadRequest, err)
		return
	}

	db := getDbConnection(c, obj.GetDB, false)

	val := reflect.New(obj.modelElem).Interface() // ptr

	// preload
	if len(obj.preloads) > 0 {
		for _, preload := range obj.preloads {
			db = db.Preload(preload)
		}
	}

	result := obj.buildPrimaryCondition(db, keys).Take(val)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			handleError(c, http.StatusNotFound, "not found")
		} else {
			handleError(c, http.StatusInternalServerError, result.Error)
		}
		return
	}

	if obj.BeforeRender != nil {
		if err := obj.BeforeRender(c, val); err != nil {
			handleError(c, http.StatusInternalServerError, err)
			return
		}
	}

	c.JSON(http.StatusOK, val)
}

func handleCreateObject(c *gin.Context, obj *WebObject) {
	var vals map[string]any
	if err := c.BindJSON(&vals); err != nil {
		handleError(c, http.StatusBadRequest, err)
		return
	}

	val := reflect.New(obj.modelElem).Interface()

	// fix mapstructure decode time.Time
	// try parse time from different layout
	config := mapstructure.DecoderConfig{
		DecodeHook: func(f reflect.Type, t reflect.Type, data any) (any, error) {
			if f.Kind() != reflect.String || t != reflect.TypeOf(time.Time{}) {
				return data, nil
			}
			layouts := []string{time.RFC3339, "2006-01-02T15:04", time.DateTime, time.DateOnly}
			for _, layout := range layouts {
				if val, err := time.Parse(layout, data.(string)); err == nil {
					return val, nil
				}
			}
			return data, nil
		},
		Result: &val,
	}
	decoder, _ := mapstructure.NewDecoder(&config)
	if err := decoder.Decode(vals); err != nil {
		handleError(c, http.StatusBadRequest, err)
		return
	}

	if obj.BeforeCreate != nil {
		if err := obj.BeforeCreate(c, val, vals); err != nil {
			handleError(c, http.StatusBadRequest, err)
			return
		}
	}

	result := obj.GetDB(c, true).Create(val)
	if result.Error != nil {
		handleError(c, http.StatusInternalServerError, result.Error)
		return
	}

	c.JSON(http.StatusOK, val)
}

func handleUpdateObject(c *gin.Context, obj *WebObject) {
	key := c.Param("key")

	var inputVals map[string]any
	if err := c.BindJSON(&inputVals); err != nil {
		handleError(c, http.StatusBadRequest, err)
		return
	}

	db := obj.GetDB(c, false)

	var vals map[string]any = map[string]any{}
	// can't edit primaryKey
	delete(inputVals, obj.jsonPKName)

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
			handleError(c, http.StatusBadRequest, fname+" type not match")
			return
		}

		vals[fname] = v
	}

	if len(obj.EditFields) > 0 {
		stripVals := make(map[string]any)
		for _, k := range obj.EditFields {
			if v, ok := vals[k]; ok {
				stripVals[k] = v
			}
		}
		vals = stripVals
	} else {
		vals = map[string]any{}
	}

	if len(vals) == 0 {
		handleError(c, http.StatusBadRequest, "not changed")
		return
	}

	if obj.BeforeUpdate != nil {
		val := reflect.New(obj.modelElem).Interface()
		if err := db.First(val, obj.gormPKName, key).Error; err != nil {
			handleError(c, http.StatusNotFound, "not found")
			return
		}
		if err := obj.BeforeUpdate(c, val, inputVals); err != nil {
			handleError(c, http.StatusBadRequest, err)
			return
		}
	}

	model := reflect.New(obj.modelElem).Interface()
	result := db.Model(model).Where(obj.gormPKName, key).Updates(vals)
	if result.Error != nil {
		handleError(c, http.StatusInternalServerError, result.Error)
		return
	}

	c.JSON(http.StatusOK, true)
}

func handleDeleteObject(c *gin.Context, obj *WebObject) {
	key := c.Param("key")
	db := obj.GetDB(c, false)

	val := reflect.New(obj.modelElem).Interface()

	// for gorm delete hook, need to load model first.
	result := db.First(val, obj.gormPKName, key)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			handleError(c, http.StatusNotFound, "not found")
		} else {
			handleError(c, http.StatusInternalServerError, result.Error)
		}
		return
	}

	if obj.BeforeDelete != nil {
		if err := obj.BeforeDelete(c, val); err != nil {
			handleError(c, http.StatusBadRequest, err)
			return
		}
	}

	result = db.Delete(val)
	if result.Error != nil {
		handleError(c, http.StatusInternalServerError, result.Error)
		return
	}

	c.JSON(http.StatusOK, true)
}

func handleBatchDelete(c *gin.Context, obj *WebObject) {
	var form []string
	if err := c.BindJSON(&form); err != nil {
		handleError(c, http.StatusBadRequest, err)
		return
	}

	db := obj.GetDB(c, false)

	val := reflect.New(obj.modelElem).Interface()
	r := db.Delete(&val, form)
	if r.Error != nil {
		handleError(c, http.StatusInternalServerError, r.Error)
		return
	}

	c.JSON(http.StatusOK, true)
}

func handleQueryObject(c *gin.Context, obj *WebObject, prepareQuery PrepareQuery) {
	db, form, err := prepareQuery(obj.GetDB(c, false), c)
	if err != nil {
		handleError(c, http.StatusBadRequest, err)
		return
	}

	// Use struct{} makes map like set.
	var filterFields = make(map[string]struct{})
	for _, k := range obj.FilterFields {
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
			filter.Name = getColumnName(obj.modelElem, field)
			stripFilters = append(stripFilters, filter)
		}
		form.Filters = stripFilters
	} else {
		form.Filters = []Filter{}
	}

	var orderFields = make(map[string]struct{})
	for _, k := range obj.OrderFields {
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
			order.Name = getColumnName(obj.modelElem, field)
			stripOrders = append(stripOrders, order)
		}
		form.Orders = stripOrders
	} else {
		form.Orders = []Order{}
	}

	if form.Keyword != "" {
		form.searchFields = []string{}
		for _, v := range obj.SearchFields {
			form.searchFields = append(form.searchFields, getColumnName(obj.modelElem, v))
		}
	}

	if len(form.ViewFields) > 0 {
		var stripViewFields []string
		for _, v := range form.ViewFields {
			stripViewFields = append(stripViewFields, getColumnName(obj.modelElem, v))
		}
		form.ViewFields = stripViewFields
	}

	r, err := QueryObjects(db, obj, form)
	if err != nil {
		handleError(c, http.StatusInternalServerError, err)
		return
	}

	if obj.BeforeRender != nil {
		vals := reflect.ValueOf(r.Items)
		if vals.Kind() == reflect.Slice {
			for i := 0; i < vals.Len(); i++ {
				v := vals.Index(i).Addr().Interface()
				if err := obj.BeforeRender(c, v); err != nil {
					handleError(c, http.StatusInternalServerError, err)
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
	// the real name of the db tableName
	tableName := db.NamingStrategy.TableName(obj.modelElem.Name())

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
			query = append(query, fmt.Sprintf("%s.%s LIKE @keyword", tableName, v))
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

	var count int64
	model := reflect.New(obj.modelElem).Interface()
	if err := db.Model(model).Count(&count).Error; err != nil {
		return r, err
	}
	if count <= 0 {
		return r, nil
	}
	r.Total = int(count)

	items := reflect.New(reflect.SliceOf(obj.modelElem))

	var offset int
	if form.Pagination {
		offset = (form.Pos - 1) * form.Limit
	} else {
		offset = form.Pos
	}

	if len(obj.preloads) > 0 {
		for _, v := range obj.preloads {
			db = db.Preload(v)
		}
	}

	result := db.Offset(offset).Limit(form.Limit).Find(items.Interface())
	if result.Error != nil {
		return r, result.Error
	}
	r.Items = items.Elem().Interface()
	// r.Pos += int(result.RowsAffected)
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

	if form.Pagination {
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
	case reflect.Ptr, reflect.Struct, reflect.Slice: // time.Time, associated structures
		return true
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
		reflect.Float32, reflect.Float64:
		return jsonKind == reflect.Float64
	default:
		return goKind == jsonKind
	}
}

func handleError(c *gin.Context, code int, err any) {
	switch err := err.(type) {
	case error:
		c.AbortWithStatusJSON(code, gin.H{"error": err.Error()})
		c.Error(err)
	case string:
		c.AbortWithStatusJSON(code, gin.H{"error": err})
		c.Error(errors.New(err))
	default:
		c.AbortWithStatusJSON(code, gin.H{"error": fmt.Sprintf("unknown error: %v", err)})
	}
}

// type ErrorWithCode interface {
// 	StatusCode() int
// }

// func AbortWithJSONError(c *gin.Context, code int, err error) {
// 	var errWithFileNum error = err
// 	if log.Flags()&(log.Lshortfile|log.Llongfile) != 0 {
// 		var ok bool
// 		_, file, line, ok := runtime.Caller(1)
// 		if !ok {
// 			file = "???"
// 			line = 0
// 		}
// 		pos := strings.LastIndex(file, "/")
// 		if log.Flags()&log.Lshortfile != 0 && pos >= 0 {
// 			file = file[pos+1:]
// 		}
// 		errWithFileNum = fmt.Errorf("%s:%d: %w", file, line, err)
// 	}
// 	c.Error(errWithFileNum)

// 	if e, ok := err.(ErrorWithCode); ok {
// 		code = e.StatusCode()
// 	}

// 	if c.IsAborted() {
// 		c.JSON(code, gin.H{"error": err.Error()})
// 	} else {
// 		c.AbortWithStatusJSON(code, gin.H{"error": err.Error()})
// 	}
// }
