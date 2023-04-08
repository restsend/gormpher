package gormpher

import (
	"fmt"

	"gorm.io/gorm"
)

// New is a generic function to create a record
func New[T any](db *gorm.DB, val *T) (*T, error) {
	result := db.Create(val)
	if result.Error != nil {
		return nil, result.Error
	}
	return val, nil
}

// Delete is a generic function to delete a record by struct
func Delete[T any](db *gorm.DB, val *T, where ...any) error {
	if len(where) > 0 {
		db = db.Where(where[0], where[1:]...)
	}
	return db.Where(val).Delete(val).Error
}

// Count is a generic function to count records
func Count[T any](db *gorm.DB, where ...any) (int, error) {
	var val T
	var count int64
	if len(where) > 0 {
		db = db.Where(where[0], where[1:]...)
	}
	result := db.Model(&val).Count(&count)
	if result.Error != nil {
		return 0, result.Error
	}
	return int(count), nil
}

// Update is a generic function to update a record by struct (cannot update zero value)
func Update[T any](db *gorm.DB, val *T, where ...any) error {
	if len(where) > 0 {
		db = db.Where(where[0], where[1:]...)
	}
	return db.Model(val).Updates(val).Error
}

// UpdateMap is a generic function to update a record by map (can update zero value)
func UpdateMap[T any](db *gorm.DB, val *T, m map[string]any, where ...any) error {
	if len(where) > 0 {
		db = db.Where(where[0], where[1:]...)
	}
	return db.Model(val).Updates(m).Error
}

// Get is a generic function to get a record by query
func Get[T any](db *gorm.DB, query any, args ...any) (*T, error) {
	var val T
	if query != nil && len(args) > 0 {
		db = db.Where(query, args...)
	}

	result := db.Model(&val).Take(&val)
	if result.Error != nil {
		return nil, result.Error
	}
	return &val, nil
}

// GetStruct is a generic function to get a record by struct
func GetStruct[T any](db *gorm.DB, val *T, where ...any) (*T, error) {
	if len(where) > 0 {
		db = db.Where(where[0], where[1:]...)
	}

	result := db.Where(val).Take(val)
	if result.Error != nil {
		return nil, result.Error
	}
	return val, nil
}

// GetMap is a generic function to get a record by map
func GetMap[T any](db *gorm.DB, m map[string]any, where ...any) (*T, error) {
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

// GetID is a generic function to get a record by id
// primary key type in table is int
func GetID[T any, E ~int](db *gorm.DB, id E, where ...any) (*T, error) {
	var val T

	if len(where) > 0 {
		db = db.Where(where[0], where[1:]...)
	}

	result := db.Take(&val, id)
	if result.Error != nil {
		return nil, result.Error
	}
	return &val, nil
}

// GetStrID is a generic function to get a record by primary key
// primary key type in table can be string
func GetStrID[T any, E ~string](db *gorm.DB, pk string, id E, where ...any) (*T, error) {
	var val T

	if len(where) > 0 {
		db = db.Where(where[0], where[1:]...)
	}

	result := db.Take(&val, pk, id)
	if result.Error != nil {
		return nil, result.Error
	}

	return &val, nil
}

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

type ListContext struct {
	Pos      int
	Limit    int
	Keywords map[string]string
	Filters  []Filter
	Order    string
	Where    []any
}

func ListPosKeywordFilterOrder[T any](db *gorm.DB, pos, limit int, keys map[string]string, filters []Filter, order string, where ...any) ([]T, int, error) {
	var items []T = make([]T, 0)
	var count int64

	db = db.Model(new(T))
	db = db.Scopes(KeywordScope(keys))
	db = db.Scopes(FilterScope(filters))

	if len(where) > 0 {
		db = db.Where(where[0], where[1:]...)
	}

	result := db.Count(&count)
	if result.Error != nil {
		return items, 0, result.Error
	}

	db = db.Offset(pos).Limit(limit)

	if order != "" {
		db = db.Order(order)
	}

	result = db.Find(&items)
	if result.Error != nil {
		return items, 0, result.Error
	}

	return items, int(count), nil
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
			where += k + " LIKE '%" + v + "%'"
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
