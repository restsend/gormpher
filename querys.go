package gormpher

import "gorm.io/gorm"

// type QueryFunc[T any] func(*Query[T])

type Query[T any] struct {
	db       *gorm.DB
	pos      int
	limit    int
	keywords map[string]string
	filters  []Filter
	orders   string
	wheres   []any
}

func (q *Query[T]) List(wheres ...any) ([]T, int, error) {
	return ListPosKeywordFilterOrder[T](q.db, q.pos, q.limit, q.keywords, q.filters, q.orders, q.wheres...)
}

func NewQuery[T any](db *gorm.DB) *Query[T] {
	q := &Query[T]{
		db:    db,
		pos:   0,
		limit: 50,
	}
	return q
}

func (q *Query[T]) PosLimit(pos, limit int) *Query[T] {
	q.pos = pos
	q.limit = limit
	return q
}

func (q *Query[T]) Keyword(key, value string) *Query[T] {
	if q.keywords == nil {
		q.keywords = make(map[string]string)
	}
	q.keywords[key] = value
	return q
}

func (q *Query[T]) Filter(name, op string, value any) *Query[T] {
	q.filters = append(q.filters, Filter{
		Name:  name,
		Op:    op,
		Value: value,
	})
	return q
}

func (q *Query[T]) Order(order string) *Query[T] {
	q.orders += order
	return q
}
