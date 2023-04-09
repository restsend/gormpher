package gormpher

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestQuery(t *testing.T) {
	db := initDB()

	db.Create(&user{Name: "user1", Email: "user1@example.com", Age: 10})
	db.Create(&user{Name: "user2", Email: "user2@example.com", Age: 20})
	db.Create(&user{Name: "user3", Email: "user3@example.com", Age: 30})

	list, count, err := NewQuery[user](db).
		PosLimit(0, 5).
		Keyword("name", "user").
		Keyword("email", "user").
		Filter("age", ">=", 20).
		Filter("age", "<=", 40).
		Order("age DESC").
		List()

	assert.Nil(t, err)
	assert.Equal(t, 2, count)
	assert.Equal(t, 2, len(list))
	assert.Equal(t, "user3", list[0].Name)
}
