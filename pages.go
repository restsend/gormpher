package gormpher

import "gorm.io/gorm"

// ListPage is a generic function to list records with pagination
func ListPage[T any](db *gorm.DB, page int, pageSize int, where ...any) ([]T, int, error) {
	return ListPos[T](db, (page-1)*pageSize, pageSize, where...)
}

// ListPageKeyword is a generic function to list records with pagination and keyword
func ListPageKeyword[T any](db *gorm.DB, page, pageSize int, keys map[string]string, where ...any) ([]T, int, error) {
	return ListPosKeywordOrder[T](db, (page-1)*pageSize, pageSize, keys, "", where...)
}

// ListPageOrder is a generic function to list records with pagination and orde
func ListPageOrder[T any](db *gorm.DB, page, pageSize int, order string, where ...any) ([]T, int, error) {
	return ListPosOrder[T](db, (page-1)*pageSize, pageSize, order, where...)
}

// ListPageKeywordOrder is a generic function to list records with pagination, keyword and orde
func ListPageKeywordOrder[T any](db *gorm.DB, page, pageSize int, keys map[string]string, order string, where ...any) ([]T, int, error) {
	return ListPosKeywordFilterOrder[T](db, (page-1)*pageSize, pageSize, keys, nil, order, where...)
}

// ListPageKeywordFilterOrder is a generic function to list records with pagination, keyword, filter and order
func ListPageKeywordFilterOrder[T any](db *gorm.DB, page, pageSize int, keys map[string]string, filters []Filter, order string, where ...any) ([]T, int, error) {
	return ListPosKeywordFilterOrder[T](db, (page-1)*pageSize, pageSize, keys, filters, order, where...)
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
