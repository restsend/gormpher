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

// GetByStruct is a generic function to get a record by struct
func GetByStruct[T any](db *gorm.DB, val *T, where ...any) (*T, error) {
	if len(where) > 0 {
		db = db.Where(where[0], where[1:]...)
	}

	result := db.Where(val).Take(val)
	if result.Error != nil {
		return nil, result.Error
	}
	return val, nil
}

// GetByMap is a generic function to get a record by map
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

// GetByID is a generic function to get a record by id
// primary key type in table is int
func GetByID[T any, E ~int](db *gorm.DB, id E, where ...any) (*T, error) {
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

// GetByStrID is a generic function to get a record by primary key
// primary key type in table can be string
func GetByStrID[T any, E ~string](db *gorm.DB, pk string, id E, where ...any) (*T, error) {
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

// List is a generic function to list records
func List[T any](db *gorm.DB, where ...any) ([]T, int, error) {
	return ListPos[T](db, 0, -1, where...)
}

// ListPos is a generic function to list records with pos and limit
func ListPos[T any](db *gorm.DB, pos, limit int, where ...any) ([]T, int, error) {
	return ListPosKeyword[T](db, pos, limit, nil, where...)
}

// ListPage is a generic function to list records with pagination
func ListPage[T any](db *gorm.DB, page int, pageSize int, where ...any) ([]T, int, error) {
	return ListPos[T](db, (page-1)*pageSize, pageSize, where...)
}

// ListPos is a generic function to list records with order
func ListOrder[T any](db *gorm.DB, order string, where ...any) ([]T, int, error) {
	return ListPosOrder[T](db, 0, -1, order, where...)
}

// ListKeyword is a generic function to list records with keyword
func ListKeyword[T any](db *gorm.DB, keys map[string]any, where ...any) ([]T, int, error) {
	return ListPosKeyword[T](db, 0, -1, keys, where...)
}

// ListPageKeyword is a generic function to list records with pagination and keyword
func ListPageKeyword[T any](db *gorm.DB, page, pageSize int, keys map[string]any, where ...any) ([]T, int, error) {
	return ListPosKeywordOrder[T](db, (page-1)*pageSize, pageSize, keys, "", where...)
}

// ListKeywordOrder is a generic function to list records with keyword and order
func ListPosKeyword[T any](db *gorm.DB, pos, limit int, keys map[string]any, where ...any) ([]T, int, error) {
	return ListPosKeywordOrder[T](db, pos, limit, keys, "", where...)
}

// ListPosOrder is a generic function to list records with pos and limit and order
func ListPosOrder[T any](db *gorm.DB, pos, limit int, order string, where ...any) ([]T, int, error) {
	return ListPosKeywordOrder[T](db, pos, limit, nil, order, where...)
}

// ListPageOrder is a generic function to list records with pagination and orde
func ListPageOrder[T any](db *gorm.DB, page, pageSize int, order string, where ...any) ([]T, int, error) {
	return ListPosOrder[T](db, (page-1)*pageSize, pageSize, order, where...)
}

// ListPageKeywordOrder is a generic function to list records with pagination, keyword and orde
func ListPageKeywordOrder[T any](db *gorm.DB, page, pageSize int, keys map[string]any, order string, where ...any) ([]T, int, error) {
	return ListPosKeywordFilterOrder[T](db, (page-1)*pageSize, pageSize, keys, nil, order, where...)
}

// ListKeywordOrder is a generic function to list records with pagination, keyword and order
func ListPosKeywordOrder[T any](db *gorm.DB, pos, limit int, keys map[string]any, order string, where ...any) ([]T, int, error) {
	return ListPosKeywordFilterOrder[T](db, pos, limit, keys, nil, order, where...)
}

func ListPageKeywordFilterOrder[T any](db *gorm.DB, page, pageSize int, keys map[string]any, filters map[string]any, order string, where ...any) ([]T, int, error) {
	return ListPosKeywordFilterOrder[T](db, (page-1)*pageSize, pageSize, keys, filters, order, where...)
}

func ListPosKeywordFilterOrder[T any](db *gorm.DB, pos, limit int, keys map[string]any, filters map[string]any, order string, where ...any) ([]T, int, error) {
	var items []T = make([]T, 0)
	var count int64

	db = db.Model(new(T))
	db = db.Scopes(KeywordScopes(keys))
	db = db.Scopes(FilterScopes(filters))

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

func KeywordScopes(keys map[string]any) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		var where string

		for k, v := range keys {
			if v == nil {
				continue
			}

			var val string
			switch v := v.(type) {
			case string:
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

			where += k + " LIKE '%" + val + "%'"
		}

		if where == "" {
			return db
		}

		return db.Where(where)
	}
}

func FilterScopes(filters map[string]any) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		var where string

		for k, v := range filters {
			if v == nil {
				continue
			}

			var val string
			switch v := v.(type) {
			case string:
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

// func Paginate(page, pageSize int) func(db *gorm.DB) *gorm.DB {
// 	return func(db *gorm.DB) *gorm.DB {
// 		if page <= 0 {
// 			page = 1
// 		}

// 		switch {
// 		case pageSize > 100:
// 			pageSize = 100
// 		case pageSize <= 0:
// 			pageSize = 10
// 		}

// 		offset := (page - 1) * pageSize
// 		return db.Offset(offset).Limit(pageSize)
// 	}
// }
