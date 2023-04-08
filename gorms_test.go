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
	db, _ := gorm.Open(sqlite.Open("file::memory:"), &gorm.Config{SkipDefaultTransaction: true})
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
		val, err := GetStruct(db, &user{})
		assert.Nil(t, err)
		assert.NotNil(t, val)
	}
	{
		val, err := GetStruct(db, &user{Name: "demo", Enabled: true})
		assert.Nil(t, err)
		assert.NotNil(t, val)
	}
	{
		val, err := GetStruct(db, &user{}, "enabled", true)
		assert.Nil(t, err)
		assert.NotNil(t, val)
	}
}

func TestUpdate(t *testing.T) {
	{
		db := initDB()
		db.Create(&user{ID: 1, Name: "demo", Email: "demo@example.com", Age: 11, Enabled: true})

		err := Update(db, &user{ID: 1, Name: "update"}, "email", "demo@example.com")
		assert.Nil(t, err)

		val, _ := GetStruct(db, &user{Name: "update"})
		assert.Equal(t, "update", val.Name)
	}
	// not found
	{
		db := initDB()
		db.Create(&user{ID: 1, Name: "demo", Email: "demo@example.com", Age: 11, Enabled: true})

		err := Update(db, &user{ID: 1, Name: "update", Email: "update@example.com"}, "name", "xxx")
		assert.Nil(t, err)

		_, err = GetStruct(db, &user{Name: "update"})
		assert.NotNil(t, err)
	}
	{
		db := initDB()
		db.Create(&product{UUID: "aaa", Name: "productAAA"})

		err := Update(db, &product{UUID: "aaa", Name: "productBBB"})
		assert.Nil(t, err)

		val, err := GetStrID[product](db, "uuid", "aaa")
		assert.Nil(t, err)
		assert.Equal(t, "productBBB", val.Name)
	}
}

func TestUpdateMap(t *testing.T) {
	{
		db := initDB()
		db.Create(&user{ID: 1, Name: "demo", Email: "demo@example.com", Age: 11, Enabled: true})

		err := UpdateMap(db, &user{ID: 1}, map[string]any{"name": "update"}, "email", "demo@example.com")
		assert.Nil(t, err)

		val, _ := GetStruct(db, &user{Name: "update"})
		assert.Equal(t, "update", val.Name)
	}
	// not found
	{
		db := initDB()
		db.Create(&user{ID: 1, Name: "demo", Email: "demo@example.com", Age: 11, Enabled: true})

		err := UpdateMap(db, &user{ID: 1}, map[string]any{"name": "update"}, "name", "xxx")
		assert.Nil(t, err)

		_, err = GetStruct(db, &user{Name: "update"})
		assert.NotNil(t, err)
	}
}

func TestGetByMap(t *testing.T) {
	db := initDB()

	db.Create(&user{Name: "demo", Enabled: true})

	val, err := GetMap[user](db, map[string]any{"name": "demo", "enabled": true})
	assert.Nil(t, err)
	assert.NotNil(t, val)
}

func TestGetByIcD(t *testing.T) {
	db := initDB()

	{
		db.Create(&user{Name: "demo", Enabled: true})

		val, err := GetID[user](db, 1)
		assert.Nil(t, err)
		assert.NotEmpty(t, val.ID)

		val, err = GetID[user](db, 1, "name = ? AND enabled = ?", "demo", true)
		assert.Nil(t, err)
		assert.NotEmpty(t, val.ID)
	}
	{
		db.Create(&product{UUID: "aaaa", Name: "demoproduct"})

		// SELECT * FROM `products` WHERE uuid = "aaaa" LIMIT 1
		val, err := GetStrID[product](db, "uuid", "aaaa")
		assert.Nil(t, err)
		assert.NotNil(t, val)

		// SELECT * FROM `products` WHERE `name` = "demoproduct" AND uuid = "aaaa" LIMIT 1
		val, err = GetStrID[product](db, "uuid", "aaaa", "name = ? AND can_buy = ?", "demoproduct", false)
		assert.Nil(t, err)
		assert.NotNil(t, val)
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
		list, count, err := ListKeyword[user](db, map[string]string{"name": "", "email": "example"})
		assert.Nil(t, err)
		assert.Equal(t, 3, count)
		assert.Equal(t, 3, len(list))
	}
	{
		list, count, err := ListKeyword[user](db, map[string]string{"name": "1"})
		assert.Nil(t, err)
		assert.Equal(t, 1, count)
		assert.Equal(t, 1, len(list))
	}
	{
		search := map[string]string{"name": "1"}
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
		search := map[string]string{"name": "2", "email": "example"}
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
		list, count, err := ListPageKeyword[user](db, 1, 2, map[string]string{"name": "", "email": "example"})
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
		list, count, err := ListPageKeyword[user](db, 1, 1, map[string]string{"name": "12"})
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
		list, count, err := ListPageKeyword[user](db, 1, 101, map[string]string{"name": "1"}, "name", "user1")
		assert.Nil(t, err)
		assert.Equal(t, 1, count)
		assert.Equal(t, 1, len(list))

		list, count, err = ListPageKeyword[user](db, 1, 101, map[string]string{"name": "2"}, "name", "user1")
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

func TestListPageKeywordFilterOrder(t *testing.T) {
	db := initDB()

	db.Create(&user{Name: "user1", Email: "user1@example.com", Age: 10})
	db.Create(&user{Name: "user2", Email: "user2@example.com", Age: 20})
	db.Create(&user{Name: "user3", Email: "user3@example.com", Age: 30})

	{
		list, count, err := ListPosKeywordFilterOrder[user](db, 0, 5, nil, nil, "")
		assert.Nil(t, err)
		assert.Equal(t, 3, count)
		assert.Equal(t, 3, len(list))
	}

	// Order
	{
		list, count, err := ListPosKeywordFilterOrder[user](db, 0, 5, nil, nil, "age DESC")
		assert.Nil(t, err)
		assert.Equal(t, 3, count)
		assert.Equal(t, "user3", list[0].Name)

		list, count, err = ListPosKeywordFilterOrder[user](db, 0, 5, nil, nil, "age ASC")
		assert.Nil(t, err)
		assert.Equal(t, 3, count)
		assert.Equal(t, "user1", list[0].Name)
	}

	// Keyword
	{
		keywords := map[string]string{"name": "user1", "email": "user1"}
		list, count, err := ListPosKeywordFilterOrder[user](db, 0, 5, keywords, nil, "")
		assert.Nil(t, err)
		assert.Equal(t, 1, count)
		assert.Equal(t, "user1", list[0].Name)

		keywords = map[string]string{"name": "user", "email": "user"}
		list, count, err = ListPosKeywordFilterOrder[user](db, 0, 5, keywords, nil, "")
		assert.Nil(t, err)
		assert.Equal(t, 3, count)
		assert.Equal(t, "user1", list[0].Name)

		keywords = map[string]string{"age": "0"}
		list, count, err = ListPosKeywordFilterOrder[user](db, 0, 5, keywords, nil, "")
		assert.Nil(t, err)
		assert.Equal(t, 3, count)
		assert.Equal(t, "user1", list[0].Name)
	}

	// Filter
	{
		filters := []Filter{
			{Name: "name", Op: "=", Value: "user2"},
			{Name: "age", Op: "=", Value: 20},
		}
		list, count, err := ListPageKeywordFilterOrder[user](db, 0, 5, nil, filters, "")
		assert.Nil(t, err)
		assert.Equal(t, 1, count)
		assert.Equal(t, "user2", list[0].Name)
	}

	// All
	{
		filters := []Filter{
			{Name: "name", Op: "in", Value: []string{"user2", "user3"}},
			{Name: "age", Op: "in", Value: []int{20, 30}},
		}
		keywords := map[string]string{"name": "user", "email": "example"}

		list, count, err := ListPageKeywordFilterOrder[user](db, 0, 5, keywords, filters, "age DESC")
		assert.Nil(t, err)
		assert.Equal(t, 2, count)
		assert.Equal(t, "user3", list[0].Name)
	}
}

func TestListContext(t *testing.T) {
	db := initDB()

	db.Create(&user{Name: "user1", Email: "user1@example.com", Age: 10})
	db.Create(&user{Name: "user2", Email: "user2@example.com", Age: 20})
	db.Create(&user{Name: "user3", Email: "user3@example.com", Age: 30})

	{
		list, count, err := List[user](db, nil)
		assert.Nil(t, err)
		assert.Equal(t, 3, count)
		assert.Equal(t, "user1", list[0].Name)
	}
	{
		list, count, err := List[user](db.Debug(), &ListContext{
			Pos:      0,
			Limit:    5,
			Keywords: map[string]string{"name": "user", "email": "example"},
			Filters: []Filter{
				{Name: "name", Op: "=", Value: "user2"},
			},
		})
		assert.Nil(t, err)
		assert.Equal(t, 1, count)
		assert.Equal(t, "user2", list[0].Name)
	}
}

func TestFilterScopes(t *testing.T) {
	db := initDB()

	db.Create(&user{Name: "user1", Email: "user1@example.com", Age: 10})
	db.Create(&user{Name: "user2", Email: "user2@example.com", Age: 20})
	db.Create(&user{Name: "user3", Email: "user3@example.com", Age: 30})

	{
		var list []user
		filters := []Filter{
			{Name: "name", Op: "=", Value: "user1"},
			{Name: "age", Op: "<", Value: 15},
		}
		r := db.Scopes(FilterScope(filters)).Find(&list)
		assert.Nil(t, r.Error)
		assert.Equal(t, 1, len(list))
	}
	{
		var list []user
		filters := []Filter{
			{Name: "name", Op: "=", Value: "user3"},
			{Name: "age", Op: ">", Value: 25},
		}
		r := db.Scopes(FilterScope(filters)).Find(&list)
		assert.Nil(t, r.Error)
		assert.Equal(t, 1, len(list))
	}
	{
		var list []user
		filters := []Filter{
			{Name: "age", Op: "in", Value: []int{10, 20}},
		}
		r := db.Scopes(FilterScope(filters)).Find(&list)
		assert.Nil(t, r.Error)
		assert.Equal(t, 2, len(list))
	}
}

func TestKeywordScopes(t *testing.T) {
	db := initDB()

	db.Create(&user{Name: "user1", Email: "user1@example.com", Age: 10})
	db.Create(&user{Name: "user2", Email: "user2@example.com", Age: 20})
	db.Create(&user{Name: "user3", Email: "user3@example.com", Age: 30})

	{
		var list []user
		keywords := map[string]string{"name": "user"}
		r := db.Scopes(KeywordScope(keywords)).Find(&list)
		assert.Nil(t, r.Error)
		assert.Equal(t, 3, len(list))
	}
	{
		var list []user
		keywords := map[string]string{"name": "user1", "email": "user2"}
		r := db.Scopes(KeywordScope(keywords)).Find(&list)
		assert.Nil(t, r.Error)
		assert.Equal(t, 2, len(list))
	}
	{
		var list []user
		keywords := map[string]string{"name": "notexist", "email": "notexist", "age": "0"}
		r := db.Scopes(KeywordScope(keywords)).Find(&list)
		assert.Nil(t, r.Error)
		assert.Equal(t, 3, len(list))
	}
}
