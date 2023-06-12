package gormpher

import (
	"fmt"
	"reflect"

	"gorm.io/gorm"
	"gorm.io/gorm/schema"
	"gorm.io/gorm/utils"
)

// gorm utils
func getPkColumnName(rt reflect.Type) string {
	var columnName string
	for i := 0; i < rt.NumField(); i++ {
		field := rt.Field(i)
		tagSetting := schema.ParseTagSetting(field.Tag.Get("gorm"), ";")
		isPrimaryKey := utils.CheckTruth(tagSetting["PRIMARYKEY"], tagSetting["PRIMARY_KEY"])
		if isPrimaryKey {
			name, ok := tagSetting["COLUMN"]
			if !ok {
				namingStrategy := schema.NamingStrategy{}
				name = namingStrategy.ColumnName("", field.Name)
			}
			columnName = name
			break
		}
	}
	return columnName
}

func getColumnName(rt reflect.Type, name string) string {
	field, ok := rt.FieldByName(name)
	if !ok {
		return ""
	}

	tagSetting := schema.ParseTagSetting(field.Tag.Get("gorm"), ";")
	val, ok := tagSetting["COLUMN"]
	if !ok {
		namingStrategy := schema.NamingStrategy{}
		val = namingStrategy.ColumnName("", field.Name)
	}
	return val
}

// gorm functions

func UpdateFields[T any](db *gorm.DB, model *T, vals map[string]any) error {
	return db.Model(model).Updates(vals).Error
}

func New[T any](db *gorm.DB, val *T) (*T, error) {
	result := db.Create(val)
	if result.Error != nil {
		return nil, result.Error
	}
	return val, nil
}

func Count[T any](db *gorm.DB, where ...any) (int, error) {
	var count int64
	if len(where) > 0 {
		db = db.Where(where[0], where[1:]...)
	}
	result := db.Model(new(T)).Count(&count)
	if result.Error != nil {
		return 0, result.Error
	}
	return int(count), nil
}

// Delete
func Delete[T any](db *gorm.DB, val *T, where ...any) error {
	if len(where) > 0 {
		db = db.Where(where[0], where[1:]...)
	}
	return db.Where(val).Delete(val).Error
}

func DeleteByID[T any, E ~int | ~string](db *gorm.DB, id E, where ...any) error {
	if len(where) > 0 {
		db = db.Where(where[0], where[1:]...)
	}
	return db.Where(GetPkColumnName[T](), id).Delete(new(T)).Error
}

func DeleteByMap[T any](db *gorm.DB, m map[string]any, where ...any) error {
	if len(where) > 0 {
		db = db.Where(where[0], where[1:]...)
	}
	return db.Where(m).Delete(new(T)).Error
}

// Get
func Get[T any](db *gorm.DB, val *T, where ...any) (*T, error) {
	if len(where) > 0 {
		db = db.Where(where[0], where[1:]...)
	}

	result := db.Where(val).Take(val)
	if result.Error != nil {
		return nil, result.Error
	}
	return val, nil
}

func GetByMap[T any](db *gorm.DB, m map[string]any, where ...any) (*T, error) {
	var val T

	if len(where) > 0 {
		db = db.Where(where[0], where[1:]...)
	}

	result := db.Model(&val).Where(m).Take(&val)
	if result.Error != nil {
		return nil, result.Error
	}
	return &val, nil
}

func GetByID[T any, E ~int | ~string](db *gorm.DB, id E, where ...any) (*T, error) {
	var val T

	if len(where) > 0 {
		db = db.Where(where[0], where[1:]...)
	}

	result := db.Take(&val, GetPkColumnName[T](), id)
	if result.Error != nil {
		return nil, result.Error
	}
	return &val, nil
}

// Update
func Update[T any](db *gorm.DB, val *T, where ...any) error {
	if len(where) > 0 {
		db = db.Where(where[0], where[1:]...)
	}
	return db.Model(val).Updates(val).Error
}

func UpdateByID[T any, E ~string | ~int](db *gorm.DB, id E, val *T, where ...any) error {
	if len(where) > 0 {
		db = db.Where(where[0], where[1:]...)
	}

	return db.Model(new(T)).Where(GetPkColumnName[T](), id).Updates(val).Error
}

func UpdateSelectByID[T any, E ~string | ~int](db *gorm.DB, id E, selects []string, val *T, where ...any) error {
	pk := GetPkColumnName[T]()

	for _, s := range selects {
		if s == pk {
			return fmt.Errorf("can not update primary key")
		}
	}

	if len(where) > 0 {
		db = db.Where(where[0], where[1:]...)
	}

	return db.Model(new(T)).Where(GetPkColumnName[T](), id).Select(selects).Updates(val).Error
}

func UpdateMapByID[T any, E ~string | ~int](db *gorm.DB, id E, m map[string]any, where ...any) error {
	if len(where) > 0 {
		db = db.Where(where[0], where[1:]...)
	}
	return db.Model(new(T)).Where(GetPkColumnName[T](), id).Updates(m).Error
}

// Query List
func ListPos[T any](db *gorm.DB, pos, limit int, where ...any) ([]T, int, error) {
	return ListPosKeyword[T](db, pos, limit, nil, where...)
}

func ListOrder[T any](db *gorm.DB, order string, where ...any) ([]T, int, error) {
	return ListPosOrder[T](db, 0, -1, order, where...)
}

func ListKeyword[T any](db *gorm.DB, keys map[string]string, where ...any) ([]T, int, error) {
	return ListPosKeyword[T](db, 0, -1, keys, where...)
}

func ListFilter[T any](db *gorm.DB, filter []Filter, where ...any) ([]T, int, error) {
	return ListPosFilter[T](db, 0, -1, filter, where...)
}

func ListPosKeyword[T any](db *gorm.DB, pos, limit int, keys map[string]string, where ...any) ([]T, int, error) {
	return ListPosKeywordOrder[T](db, pos, limit, keys, "", where...)
}

func ListPosOrder[T any](db *gorm.DB, pos, limit int, order string, where ...any) ([]T, int, error) {
	return ListPosKeywordOrder[T](db, pos, limit, nil, order, where...)
}

func ListPosFilter[T any](db *gorm.DB, pos, limit int, filter []Filter, where ...any) ([]T, int, error) {
	return ListPosKeywordFilterOrder[T](db, pos, limit, nil, filter, "", where...)
}

func ListPosKeywordOrder[T any](db *gorm.DB, pos, limit int, keys map[string]string, order string, where ...any) ([]T, int, error) {
	return ListPosKeywordFilterOrder[T](db, pos, limit, keys, nil, order, where...)
}

func ListPosKeywordFilter[T any](db *gorm.DB, pos, limit int, keys map[string]string, filter []Filter, where ...any) ([]T, int, error) {
	return ListPosKeywordFilterOrder[T](db, pos, limit, keys, filter, "", where...)
}

func ListPosKeywordFilterOrder[T any](db *gorm.DB, pos, limit int, keys map[string]string, filters []Filter, order string, where ...any) ([]T, int, error) {
	return ListPosKeywordFilterOrderModel[T, T](db, pos, limit, keys, filters, order, where...)
}

// List Model
func ListPosKeywordFilterOrderModel[T, R any](db *gorm.DB, pos, limit int, keys map[string]string, filters []Filter, order string, where ...any) ([]R, int, error) {
	var items []R = make([]R, 0)
	var count int64

	db = db.Model(new(T))
	db = db.Scopes(KeywordScope(keys))
	db = db.Scopes(FilterScope(filters))

	if len(where) > 0 {
		db = db.Where(where[0], where[1:]...)
	}

	if err := db.Count(&count).Error; err != nil {
		return items, 0, err
	}

	db = db.Offset(pos).Limit(limit)

	if order != "" {
		db = db.Order(order)
	}

	if err := db.Scan(&items).Error; err != nil {
		return items, 0, err
	}

	return items, int(count), nil
}

// List Context
type ListContext struct {
	Pos      int
	Limit    int
	Keywords map[string]string
	Filters  []Filter
	Order    string
	Where    []any
}

func List[T any](db *gorm.DB, ctx *ListContext) ([]T, int, error) {
	if ctx == nil {
		return ListPosKeywordFilterOrder[T](db, 0, 50, nil, nil, "", nil)
	}

	pos, limit := ctx.Pos, ctx.Limit
	if pos < 0 {
		pos = 0
	}
	switch {
	case limit <= 0:
		limit = 50
	case limit > 200:
		limit = 200
	}
	return ListPosKeywordFilterOrder[T](db, pos, limit, ctx.Keywords, ctx.Filters, ctx.Order, ctx.Where...)
}

func ListModel[T, R any](db *gorm.DB, ctx *ListContext) ([]R, int, error) {
	if ctx == nil {
		return ListPosKeywordFilterOrderModel[T, R](db, 0, 50, nil, nil, "", nil)
	}

	pos, limit := ctx.Pos, ctx.Limit
	if pos < 0 {
		pos = 0
	}
	switch {
	case limit <= 0:
		limit = 50
	case limit > 200:
		limit = 200
	}
	return ListPosKeywordFilterOrderModel[T, R](db, pos, limit, ctx.Keywords, ctx.Filters, ctx.Order, ctx.Where...)
}

// Pagination functions

func ListPage[T any](db *gorm.DB, page int, pageSize int, where ...any) ([]T, int, error) {
	return ListPos[T](db, (page-1)*pageSize, pageSize, where...)
}

func ListPageKeyword[T any](db *gorm.DB, page, pageSize int, keys map[string]string, where ...any) ([]T, int, error) {
	return ListPosKeywordOrder[T](db, (page-1)*pageSize, pageSize, keys, "", where...)
}

func ListPageOrder[T any](db *gorm.DB, page, pageSize int, order string, where ...any) ([]T, int, error) {
	return ListPosOrder[T](db, (page-1)*pageSize, pageSize, order, where...)
}

func ListPageKeywordOrder[T any](db *gorm.DB, page, pageSize int, keys map[string]string, order string, where ...any) ([]T, int, error) {
	return ListPosKeywordFilterOrder[T](db, (page-1)*pageSize, pageSize, keys, nil, order, where...)
}

func ListPageKeywordFilterOrder[T any](db *gorm.DB, page, pageSize int, keys map[string]string, filters []Filter, order string, where ...any) ([]T, int, error) {
	return ListPosKeywordFilterOrder[T](db, (page-1)*pageSize, pageSize, keys, filters, order, where...)
}

func ListPageKeywordFilterOrderModel[T, R any](db *gorm.DB, page, pageSize int, keys map[string]string, filters []Filter, order string, where ...any) ([]R, int, error) {
	return ListPosKeywordFilterOrderModel[T, R](db, (page-1)*pageSize, pageSize, keys, filters, order, where...)
}

// {"name": "mockname", "nick": "mocknick" }
// => name LIKE '%mockname%' OR nick LIKE '%mocknick%'
func KeywordScope(keys map[string]string) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		var where string

		for k, v := range keys {
			if v == "" {
				continue
			}
			if where != "" {
				where += " OR "
			}
			where += fmt.Sprintf("`%s` LIKE '%%%s%%'", k, v)
		}

		if where == "" {
			return db
		}

		return db.Where(where)
	}
}

// [{"name": "name", op: "=", "value": "mockname" }, {"name": "age", "op": "<", "value": 20 }]
// => name = 'mockname' AND age < 20
func FilterScope(filters []Filter) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		for _, f := range filters {
			q := f.GetQuery()
			if q != "" {
				db = db.Where(q, f.Value)
			}
		}
		return db
	}
}

func GetPkColumnName[T any]() string {
	rt := reflect.TypeOf(new(T)).Elem()

	var columnName string
	for i := 0; i < rt.NumField(); i++ {
		field := rt.Field(i)
		tagSetting := schema.ParseTagSetting(field.Tag.Get("gorm"), ";")
		isPrimaryKey := utils.CheckTruth(tagSetting["PRIMARYKEY"], tagSetting["PRIMARY_KEY"])
		if isPrimaryKey {
			name, ok := tagSetting["COLUMN"]
			if !ok {
				namingStrategy := schema.NamingStrategy{}
				name = namingStrategy.ColumnName("", field.Name)
			}
			columnName = name
			break
		}
	}
	if columnName == "" {
		return "id"
	}
	return columnName
}

// {"name": "mockname", "age": 10 }
// => name = 'mockname' AND age = 10
func FilterEqualScope(filters map[string]any) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		var where string

		for k, v := range filters {
			if v == nil {
				continue
			}

			var val string
			switch v := v.(type) {
			case string:
				if v == "" {
					continue
				}
				val = v
			case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
				val = fmt.Sprintf("%d", v)
			case bool:
				val = fmt.Sprintf("%t", v)
			case float32, float64:
				val = fmt.Sprintf("%f", v)
			default:
				continue
			}

			if where != "" {
				where += " AND "
			}

			where = fmt.Sprintf("%s %s = '%s'", where, k, val)
		}

		if where == "" {
			return db
		}

		return db.Where(where)
	}
}

func PageScope(page, pageSize int) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if page <= 0 {
			page = 1
		}

		switch {
		case pageSize > 100:
			pageSize = 100
		case pageSize <= 0:
			pageSize = 10
		}

		offset := (page - 1) * pageSize
		return db.Offset(offset).Limit(pageSize)
	}
}
