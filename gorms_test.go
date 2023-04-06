package gormpher

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type user struct {
	ID        uint `gorm:"primarykey"`
	CreatedAt time.Time
	UpdatedAt time.Time
	Name      string
	Email     string
	Age       int
	Enabled   bool
}

type product struct {
	UUID      string `gorm:"primarykey"`
	CreatedAt time.Time
	UpdatedAt time.Time
	Name      string
	CanBuy    bool
}

func initDB() *gorm.DB {
	db, _ := gorm.Open(sqlite.Open("file::memory:"), &gorm.Config{})
	db.AutoMigrate(user{}, product{})
	return db
}

func TestNew(t *testing.T) {
	db := initDB()

	val, err := New(db, &user{
		Name:    "mockname",
		Age:     11,
		Enabled: true,
	})
	assert.Nil(t, err)
	assert.NotEmpty(t, val.ID)

	p, err := New(db, &product{
		UUID:   "aaaa",
		Name:   "demoproduct",
		CanBuy: true,
	})
	assert.Nil(t, err)
	assert.NotNil(t, p)
}

func TestGet(t *testing.T) {
	db := initDB()

	db.Create(&user{Name: "demo", Enabled: true})

	{
		val, err := Get[user](db, "name = ? AND enabled = ?", "demo", true)
		assert.Nil(t, err)
		assert.NotNil(t, val.ID)
	}
}

func TestGetStruct(t *testing.T) {
	db := initDB()

	db.Create(&user{Name: "demo", Enabled: true})
	{
		val, err := GetByStruct(db, &user{})
		assert.Nil(t, err)
		assert.NotNil(t, val)
	}
	{
		val, err := GetByStruct(db, &user{Name: "demo", Enabled: true})
		assert.Nil(t, err)
		assert.NotNil(t, val)
	}
	{
		val, err := GetByStruct(db, &user{}, "enabled", true)
		assert.Nil(t, err)
		assert.NotNil(t, val)
	}
}

func TestGetByMap(t *testing.T) {
	db := initDB()

	db.Create(&user{Name: "demo", Enabled: true})

	val, err := GetByMap[user](db, map[string]any{"name": "demo", "enabled": true})
	assert.Nil(t, err)
	assert.NotNil(t, val)
}

func TestGetByIcD(t *testing.T) {
	db := initDB()

	{
		db.Create(&user{Name: "demo", Enabled: true})

		val, err := GetByID[user](db, 1)
		assert.Nil(t, err)
		assert.NotEmpty(t, val.ID)

		val, err = GetByID[user](db, 1, "name = ? AND enabled = ?", "demo", true)
		assert.Nil(t, err)
		assert.NotEmpty(t, val.ID)
	}
	{
		db.Create(&product{UUID: "aaaa", Name: "demoproduct"})

		// SELECT * FROM `products` WHERE uuid = "aaaa" LIMIT 1
		val, err := GetByStrID[product](db, "uuid", "aaaa")
		assert.Nil(t, err)
		assert.NotNil(t, val)

		// SELECT * FROM `products` WHERE `name` = "demoproduct" AND uuid = "aaaa" LIMIT 1
		val, err = GetByStrID[product](db, "uuid", "aaaa", "name = ? AND can_buy = ?", "demoproduct", false)
		assert.Nil(t, err)
		assert.NotNil(t, val)
	}
}

func TestGetList(t *testing.T) {
	db := initDB()

	db.Create(&user{Name: "user1", Enabled: true})
	db.Create(&user{Name: "user2", Enabled: true})
	db.Create(&user{Name: "user3", Enabled: true})

	{
		list, count, err := List[user](db)
		assert.Nil(t, err)
		assert.Equal(t, 3, count)
		assert.Equal(t, 3, len(list))
		assert.Equal(t, "user1", list[0].Name)
	}
	{
		list, count, err := List[user](db, "name", "user1")
		assert.Nil(t, err)
		assert.Equal(t, 1, count)
		assert.Equal(t, 1, len(list))
		assert.Equal(t, "user1", list[0].Name)
	}
}

func TestListPage(t *testing.T) {
	db := initDB()

	db.Create(&user{Name: "user1", Enabled: true})
	db.Create(&user{Name: "user2", Enabled: true})
	db.Create(&user{Name: "user3", Enabled: true})

	{
		list, count, err := ListPos[user](db, 0, 2)
		assert.Nil(t, err)
		assert.Equal(t, 3, count)
		assert.Equal(t, 2, len(list))
		assert.Equal(t, "user1", list[0].Name)
	}
	{
		list, count, err := ListPage[user](db, 1, 2)
		assert.Nil(t, err)
		assert.Equal(t, 3, count)
		assert.Equal(t, 2, len(list))
		assert.Equal(t, "user1", list[0].Name)
	}
	{
		list, count, err := ListPage[user](db, 1, 2, "name", "user1")
		assert.Nil(t, err)
		assert.Equal(t, 1, count)
		assert.Equal(t, 1, len(list))
		assert.Equal(t, "user1", list[0].Name)
	}
}

func TestListKeyword(t *testing.T) {
	db := initDB()

	db.Create(&user{Name: "user1", Email: "user1@example.com", Enabled: true})
	db.Create(&user{Name: "user2", Email: "user2@example.com", Enabled: true})
	db.Create(&user{Name: "user3", Email: "user3@example.com", Enabled: true})

	{
		list, count, err := ListKeyword[user](db, map[string]any{"name": nil, "email": "example"})
		assert.Nil(t, err)
		assert.Equal(t, 3, count)
		assert.Equal(t, 3, len(list))
	}
	{
		list, count, err := ListKeyword[user](db, map[string]any{"name": "1"})
		assert.Nil(t, err)
		assert.Equal(t, 1, count)
		assert.Equal(t, 1, len(list))
	}
	{
		search := map[string]any{"name": "1"}
		list, count, err := ListKeyword[user](db, search, "name", "user1")
		assert.Nil(t, err)
		assert.Equal(t, 1, count)
		assert.Equal(t, 1, len(list))

		list, count, err = ListKeyword[user](db, search, "name", "user2")
		assert.Nil(t, err)
		assert.Equal(t, 0, count)
		assert.Equal(t, 0, len(list))
	}
	{
		search := map[string]any{"name": "2", "email": "example"}
		list, count, err := ListKeyword[user](db, search, "name", "user2")
		assert.Nil(t, err)
		assert.Equal(t, 1, count)
		assert.Equal(t, 1, len(list))
	}
}

func TestListPageKeyword(t *testing.T) {
	db := initDB()

	db.Create(&user{Name: "user1", Email: "user1@example.com", Enabled: true})
	db.Create(&user{Name: "user2", Email: "user2@example.com", Enabled: true})
	db.Create(&user{Name: "user3", Email: "user3@example.com", Enabled: true})

	{
		list, count, err := ListPageKeyword[user](db, 1, 2, map[string]any{"name": nil, "email": "example"})
		assert.Nil(t, err)
		assert.Equal(t, 3, count)
		assert.Equal(t, 2, len(list))
	}
	{
		list, count, err := ListPageKeyword[user](db, 1, 2, nil)
		assert.Nil(t, err)
		assert.Equal(t, 3, count)
		assert.Equal(t, 2, len(list))
	}
	{
		list, count, err := ListPageKeyword[user](db, 1, 1, map[string]any{"name": "12"})
		assert.Nil(t, err)
		assert.Equal(t, 0, count)
		assert.Equal(t, 0, len(list))
	}
	{
		list, count, err := ListPageKeyword[user](db, 0, 101, nil)
		assert.Nil(t, err)
		assert.Equal(t, 3, count)
		assert.Equal(t, 3, len(list))
	}
	{
		list, count, err := ListPageKeyword[user](db, 1, 101, map[string]any{"name": "1"}, "name", "user1")
		assert.Nil(t, err)
		assert.Equal(t, 1, count)
		assert.Equal(t, 1, len(list))

		list, count, err = ListPageKeyword[user](db, 1, 101, map[string]any{"name": "2"}, "name", "user1")
		assert.Nil(t, err)
		assert.Equal(t, 0, count)
		assert.Equal(t, 0, len(list))
	}
}

func TestListPageKeywordOrder(t *testing.T) {
	db := initDB()

	db.Create(&user{Name: "user1", Email: "user1@example.com", Age: 10})
	db.Create(&user{Name: "user2", Email: "user2@example.com", Age: 20})
	db.Create(&user{Name: "user3", Email: "user3@example.com", Age: 30})

	{
		list, count, err := ListPageKeywordOrder[user](db, 1, 1, nil, "")
		assert.Nil(t, err)
		assert.Equal(t, 3, count)
		assert.Equal(t, 1, len(list))
		assert.Equal(t, "user1", list[0].Name)

		list, count, err = ListPageKeywordOrder[user](db, 1, 3, nil, "age DESC")
		assert.Nil(t, err)
		assert.Equal(t, 3, count)
		assert.Equal(t, 3, len(list))
		assert.Equal(t, "user3", list[0].Name)

		list, count, err = ListPageKeywordOrder[user](db, 1, 3, nil, "age ASC")
		assert.Nil(t, err)
		assert.Equal(t, 3, count)
		assert.Equal(t, 3, len(list))
		assert.Equal(t, "user1", list[0].Name)
	}
}

func TestCount(t *testing.T) {
	db := initDB()

	db.Create(&user{Name: "user1", Email: "user1@example.com", Age: 10})
	count, err := Count[user](db, nil)
	assert.Nil(t, err)
	assert.Equal(t, 1, count)

	db.Create(&user{Name: "user2", Email: "user2@example.com", Age: 20})
	count, err = Count[user](db, nil)
	assert.Nil(t, err)
	assert.Equal(t, 2, count)

	db.Create(&user{Name: "user3", Email: "user3@example.com", Age: 30})
	count, err = Count[user](db, nil)
	assert.Nil(t, err)
	assert.Equal(t, 3, count)
}

func TestDelete(t *testing.T) {
	db := initDB()

	db.Create(&user{Name: "user1", Email: "user1@example.com", Age: 10})
	db.Create(&user{Name: "user2", Email: "user2@example.com", Age: 20})
	db.Create(&user{Name: "user3", Email: "user3@example.com", Age: 30})

	err := Delete(db, &user{Name: "user1"})
	assert.Nil(t, err)
	count, _ := Count[user](db, nil)
	assert.Equal(t, 2, count)

	err = Delete(db, &user{ID: 2})
	assert.Nil(t, err)
	count, _ = Count[user](db, nil)
	assert.Equal(t, 1, count)

	// with age = -1, unable to delete
	err = Delete(db, &user{ID: 3, Name: "user3", Age: -1})
	assert.Nil(t, err)
	count, _ = Count[user](db, nil)
	assert.Equal(t, 1, count)

	err = Delete(db, &user{}, "name", "user3")
	assert.Nil(t, err)
	count, _ = Count[user](db, nil)
	assert.Equal(t, 0, count)
}
