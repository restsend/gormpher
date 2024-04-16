package gormpher

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestObjectCRUD(t *testing.T) {
	type User struct {
		ID   uint   `json:"uid" gorm:"primarykey"`
		Name string `gorm:"size:100"`
		Age  int
		Body string `json:"-" gorm:"-"`
	}

	db, _ := gorm.Open(sqlite.Open("file::memory:"), nil)
	db.AutoMigrate(User{})

	err := db.Create(&User{ID: 1, Name: "user", Age: 10}).Error
	assert.Nil(t, err)

	r := gin.Default()
	r.Use(WithGormDB(db))
	webobject := WebObject{
		Model:        User{},
		EditFields:   []string{"Name"},
		FilterFields: []string{"Name"},
		SearchFields: []string{"Name"},
	}
	err = webobject.RegisterObject(&r.RouterGroup)
	assert.Nil(t, err)

	// Create
	{
		b, _ := json.Marshal(User{Name: "add"})
		req := httptest.NewRequest(http.MethodPut, "/user", bytes.NewReader(b))
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Result().StatusCode)
		assert.Contains(t, w.Body.String(), `"uid":2`)
	}
	// Single Query
	{
		req := httptest.NewRequest(http.MethodGet, "/user/1", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Result().StatusCode)
		assert.Contains(t, w.Body.String(), `"uid":1`)
		assert.Contains(t, w.Body.String(), `"Name":"user"`)
	}
	// Update
	{
		b, _ := json.Marshal(User{Name: "update", Age: 11})
		req := httptest.NewRequest(http.MethodPatch, "/user/1", bytes.NewReader(b))
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Result().StatusCode)
		assert.Equal(t, "true", w.Body.String())
	}
	// Query
	{
		data := map[string]any{
			"pos":     0,
			"limit":   5,
			"keyword": "",
			"filters": []map[string]any{
				{
					"name":  "Name",
					"op":    "=",
					"value": "update",
				},
			},
		}
		b, _ := json.Marshal(data)
		req := httptest.NewRequest(http.MethodPost, "/user", bytes.NewReader(b))
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Result().StatusCode)

		var res QueryResult
		err := json.Unmarshal(w.Body.Bytes(), &res)
		assert.Nil(t, err)
		assert.Equal(t, 1, res.TotalCount)
		assert.Equal(t, "update", res.Items[0].(map[string]any)["Name"])
		fmt.Println(res.Items[0].(map[string]any))
	}
	// Delete
	{
		req := httptest.NewRequest(http.MethodDelete, "/user/1", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Result().StatusCode)
	}
	// Query After Delete
	{
		req := httptest.NewRequest(http.MethodGet, "/user/1", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		assert.Equal(t, http.StatusNotFound, w.Result().StatusCode)

		b, _ := json.Marshal(map[string]any{"pos": 0, "limit": 5})
		req = httptest.NewRequest(http.MethodPost, "/user", bytes.NewReader(b))
		w = httptest.NewRecorder()
		r.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Result().StatusCode)
		log.Println(w.Body.String())

		var res QueryResult
		err := json.Unmarshal(w.Body.Bytes(), &res)
		assert.Nil(t, err)
		assert.Equal(t, 1, res.TotalCount)
	}
}

func TestObjectQuery(t *testing.T) {
	type Super struct {
		Fly bool
	}
	type User struct {
		ID       uint      `json:"uid" gorm:"primarykey"`
		Name     string    `json:"name" gorm:"size:100"`
		Body     string    `json:"-" gorm:"-"`
		Birthday time.Time `json:"birthday"`
		Enabled  bool      `json:"enabled"`
		Age      int
		Super
	}

	db, _ := gorm.Open(sqlite.Open("file::memory:"), nil)
	db.AutoMigrate(User{})

	r := gin.Default()
	r.Use(WithGormDB(db))
	webobject := WebObject{
		Model:        User{},
		FilterFields: []string{"Name", "Age", "Birthday", "Enabled"},
		SearchFields: []string{"Name"},
		BeforeRender: func(db *gorm.DB, c *gin.Context, obj any) (any, error) {
			return obj, nil
		},
	}
	err := webobject.RegisterObject(&r.RouterGroup)
	assert.Nil(t, err)

	// Mock
	{
		db.Create(&User{ID: 1, Name: "alice", Age: 10, Enabled: true, Birthday: time.Now()})
		db.Create(&User{ID: 2, Name: "bob", Age: 10, Enabled: true, Birthday: time.Now()})
		db.Create(&User{ID: 3, Name: "foo", Age: 13})
		db.Create(&User{ID: 4, Name: "bar", Age: 13})
	}
	// Query
	{
		type Param struct {
			Keyword string
			Filters []map[string]any
		}
		type Except struct {
			Num int
		}
		tests := []struct {
			name   string
			params Param
			expect Except
		}{
			{"base_case_1",
				Param{Keyword: "", Filters: nil},
				Except{4},
			},
			{"base_case_2",
				Param{Keyword: "bob", Filters: nil},
				Except{1},
			},
			{"base_case_3",
				Param{Keyword: "", Filters: []map[string]any{
					{"name": "name", "op": "=", "value": "alice"},
				}},
				Except{1},
			},
			{
				"base_case_4",
				Param{Keyword: "", Filters: []map[string]any{
					{"name": "Age", "op": ">=", "value": "10"},
					{"name": "Age", "op": "<=", "value": "12"}},
				},
				Except{2},
			},
			{
				"base_case_5: multiple filters",
				Param{Keyword: "", Filters: []map[string]any{
					{"name": "Age", "op": ">", "value": "11"},
					{"name": "Age", "op": "<", "value": "15"}},
				},
				Except{2},
			},
			{
				"base_case_6:",
				Param{Keyword: "", Filters: []map[string]any{
					{"name": "Age", "op": ">", "value": "11"},
					{"name": "Age", "op": "<", "value": "15"}},
				},
				Except{2},
			},
			{
				"base_case_7:",
				Param{Keyword: "", Filters: []map[string]any{
					{"name": "name", "op": "in", "value": []any{"alice", "bob", "foo"}}},
				},
				Except{3},
			},
			{
				"base_case_8:",
				Param{Keyword: "", Filters: []map[string]any{
					{"name": "name", "op": "in", "value": []any{"alice", "bob"}},
					{"name": "Age", "op": "<>", "value": "10"}},
				},
				Except{0},
			},
			{
				"base_case_9:",
				Param{Keyword: "", Filters: []map[string]any{
					{"name": "birthday", "op": ">=", "value": "2023-01-01"}},
				},
				Except{2},
			},
			{
				"bool_case_1",
				Param{Filters: []map[string]any{
					{"name": "enabled", "op": "=", "value": false}},
				},
				Except{2},
			},
			{
				"bool_case_2",
				Param{Filters: []map[string]any{
					{"name": "enabled", "op": "=", "value": true}},
				},
				Except{2},
			},
			{
				"bool_case_3",
				Param{Filters: []map[string]any{
					{"name": "enabled", "op": "=", "value": "xxxx"}},
				},
				Except{0},
			},
			{
				"like_case_1",
				Param{Filters: []map[string]any{
					{"name": "name", "op": "like", "value": "%a%"},
				}},
				Except{2},
			},
			{
				"like_case_2",
				Param{Filters: []map[string]any{
					{"name": "name", "op": "like", "value": "%o%"},
					{"name": "name", "op": "like", "value": "%b%"},
				}},
				Except{1},
			},
			{
				"like_case_3",
				Param{Filters: []map[string]any{
					{"name": "Age", "op": "like", "value": "%3%"},
				}},
				Except{2},
			},
			{
				"not_int_case_1",
				Param{Filters: []map[string]any{
					{"name": "name", "op": "not_in", "value": []any{"alice", "bob", "foo"}},
				}},
				Except{1},
			},
			{
				"not_int_case_2",
				Param{Filters: []map[string]any{
					{"name": "Age", "op": "not_in", "value": []any{10, 13}},
				}},
				Except{0},
			},
			{
				"bad_case_1: for op not exist",
				Param{Filters: []map[string]any{
					{"name": "name", "op": "notexist", "value": "xxxx"},
				}},
				Except{4},
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				data := map[string]any{
					"pos":     0,
					"limit":   5,
					"keyword": tt.params.Keyword,
					"filters": tt.params.Filters,
				}

				b, _ := json.Marshal(data)
				req := httptest.NewRequest(http.MethodPost, "/user", bytes.NewReader(b))
				w := httptest.NewRecorder()
				r.ServeHTTP(w, req)
				assert.Equal(t, http.StatusOK, w.Result().StatusCode)

				var res QueryResult
				err := json.Unmarshal(w.Body.Bytes(), &res)
				assert.Nil(t, err)
				assert.Equal(t, tt.expect.Num, res.TotalCount)
			})
		}

	}
}

func TestObjectOrder(t *testing.T) {
	type User struct {
		UUID      string    `json:"uid" gorm:"primarykey"`
		CreatedAt time.Time `json:"createdAt"`
		Name      string    `json:"name" gorm:"size:100"`
		Age       int
	}

	db, _ := gorm.Open(sqlite.Open("file::memory:"), nil)
	db.AutoMigrate(User{})

	r := gin.Default()
	r.Use(WithGormDB(db))
	webobject := WebObject{
		Model:       User{},
		OrderFields: []string{"UUID", "Name", "Age", "CreatedAt"},
	}
	err := webobject.RegisterObject(&r.RouterGroup)
	assert.Nil(t, err)

	// Mock data
	{
		db.Create(&User{UUID: "aaa", Name: "alice", Age: 9, CreatedAt: time.Now()})
		db.Create(&User{UUID: "bbb", Name: "bob", Age: 10, CreatedAt: time.Now().Add(time.Second * 5)})
		db.Create(&User{UUID: "ccc", Name: "foo", Age: 13, CreatedAt: time.Now().Add(time.Second * 10)})
		db.Create(&User{UUID: "ddd", Name: "zoom", Age: 15, CreatedAt: time.Now().Add(time.Second * 15)})
	}
	// Query
	{
		type Param struct {
			Keyword string
			Orders  []map[string]any
		}
		type Except struct {
			ID string
		}
		tests := []struct {
			name   string
			params Param
			expect Except
		}{
			{"base_case_1:name_desc",
				Param{Orders: []map[string]any{
					{"name": "name", "op": "desc"},
				}},
				Except{"ddd"},
			},
			{"base_case_2:name_asc",
				Param{Orders: []map[string]any{
					{"name": "name", "op": "asc"},
				}},
				Except{"aaa"},
			},
			{"base_case_3:nil",
				Param{Orders: nil},
				Except{"aaa"},
			},
			{"base_case_4:age_asc",
				Param{Orders: []map[string]any{
					{"name": "Age", "op": "asc"},
				}},
				Except{"aaa"},
			},
			{"base_case_5:age_desc",
				Param{Orders: []map[string]any{
					{"name": "Age", "op": "desc"},
				}},
				Except{"ddd"},
			},
			{"base_case_6:createdAt_asc",
				Param{Orders: []map[string]any{
					{"name": "createdAt", "op": "asc"},
				}},
				Except{"aaa"},
			},
			{"base_case_5:createdAt_desc",
				Param{Orders: []map[string]any{
					{"name": "createdAt", "op": "desc"},
				}},
				Except{"ddd"},
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				data := map[string]any{
					"pos":    0,
					"limit":  5,
					"orders": tt.params.Orders,
				}

				b, _ := json.Marshal(data)
				req := httptest.NewRequest(http.MethodPost, "/user", bytes.NewReader(b))
				w := httptest.NewRecorder()
				r.ServeHTTP(w, req)
				assert.Equal(t, http.StatusOK, w.Result().StatusCode)

				var res QueryResult
				err := json.Unmarshal(w.Body.Bytes(), &res)
				assert.Nil(t, err)
				assert.Equal(t, tt.expect.ID, res.Items[0].(map[string]any)["uid"])
			})
		}

	}
}

// TODO:
func TestObjectEdit(t *testing.T) {
	type User struct {
		UUID       string     `json:"uid" gorm:"primarykey"`
		Name       string     `json:"name" gorm:"size:100"`
		Age        int        `json:"age"`
		Enabled    bool       `json:"enabled"`
		Birthday   time.Time  `json:"birthday"`
		CannotEdit string     `json:"cannotEdit"`
		PtrTime    *time.Time `json:"ptrTime"`
	}

	// Query
	{
		type Param struct {
			ID   uint
			Data map[string]any
		}
		type Except struct {
			Code int
		}
		tests := []struct {
			name   string
			params Param
			expect Except
		}{
			{"base_case_1",
				Param{1, map[string]any{
					"name": "hhhhh",
					"age":  12,
				}},
				Except{http.StatusOK},
			},
			{"base_case_2",
				Param{1, map[string]any{
					"name": true,
					"age":  "12",
				}},
				Except{http.StatusBadRequest},
			},
			{"base_case_3",
				Param{1, map[string]any{
					"name": 11,
				}},
				Except{http.StatusBadRequest},
			},
			{"base_case_4",
				Param{1, map[string]any{
					"enabled": true,
				}},
				Except{http.StatusOK},
			},
			{"bad_case_1",
				Param{1, map[string]any{}},
				Except{http.StatusBadRequest},
			},
			{"bad_case_2",
				Param{1, map[string]any{
					"xxxxxx": "xxxxxx",
				}},
				Except{http.StatusBadRequest},
			},
			{"bad_case_3",
				Param{1, map[string]any{
					"cannotEdit": "xxxxxx",
				}},
				Except{http.StatusBadRequest},
			},
			// TODO:
			{"bad_case_4",
				Param{1, map[string]any{
					"name": nil,
				}},
				Except{http.StatusBadRequest},
			},
			{"time_case_1",
				Param{
					1, map[string]any{
						"birthday": "2023-03-13T10:27:11.9802049+08:00",
					}},
				Except{http.StatusOK},
			},
			{"time_case_2",
				Param{
					1, map[string]any{
						"birthday": nil,
					}},
				Except{http.StatusBadRequest},
			},
			{"ptr_case_1",
				Param{
					1, map[string]any{
						"ptrTime": "2023-03-16T15:03:04.21432577Z",
					}},
				Except{http.StatusOK},
			},
			{"ptr_case_2",
				Param{
					1, map[string]any{
						"ptrTime": nil,
					}},
				Except{http.StatusBadRequest},
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {

				db, _ := gorm.Open(sqlite.Open("file::memory:"), nil)
				db.AutoMigrate(User{})

				r := gin.Default()
				r.Use(WithGormDB(db))
				webobject := WebObject{
					Model:      User{},
					EditFields: []string{"Name", "Age", "Enabled", "Birthday", "PtrTime"},
				}
				err := webobject.RegisterObject(&r.RouterGroup)
				assert.Nil(t, err)

				// Mock data
				{
					db.Create(&User{UUID: "aaa", Name: "alice", Age: 9})
				}

				b, _ := json.Marshal(tt.params.Data)
				req := httptest.NewRequest(http.MethodPatch, fmt.Sprintf("/user/%d", tt.params.ID), bytes.NewReader(b))
				w := httptest.NewRecorder()
				r.ServeHTTP(w, req)
				assert.Equal(t, tt.expect.Code, w.Result().StatusCode)
				if w.Result().StatusCode != http.StatusOK {
					log.Println(w.Body.String())
				}
			})
		}
	}
}

func TestObjectNoFieldEdit(t *testing.T) {
	type User struct {
		ID       uint      `json:"uid" gorm:"primarykey"`
		Name     string    `json:"name" gorm:"size:100"`
		Age      int       `json:"age"`
		Enabled  bool      `json:"enabled"`
		Birthday time.Time `json:"birthday"`
	}

	db, _ := gorm.Open(sqlite.Open("file::memory:"), nil)
	db.AutoMigrate(User{})

	r := gin.Default()
	r.Use(WithGormDB(db))
	webobject := WebObject{
		Model:      User{},
		EditFields: []string{},
	}
	err := webobject.RegisterObject(&r.RouterGroup)
	assert.Nil(t, err)

	db.Create(&User{ID: 1, Name: "alice", Age: 9})

	var data = map[string]any{
		"name":    "updatename",
		"age":     11,
		"enabled": true,
		"birthay": "2022-02-02 11:11:11",
	}
	b, _ := json.Marshal(data)
	req := httptest.NewRequest(http.MethodPatch, fmt.Sprintf("/user/%d", 1), bytes.NewReader(b))
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Result().StatusCode)
}

func TestUpdatePtrTime(t *testing.T) {
	type User struct {
		ID       uint      `json:"uid" gorm:"primarykey"`
		Birthday time.Time `json:"birthday"`
	}

	db, _ := gorm.Open(sqlite.Open("file::memory:"), nil)
	db.AutoMigrate(User{})

	r := gin.Default()
	r.Use(WithGormDB(db))
	webobject := WebObject{
		Model:      User{},
		EditFields: []string{"Birthday"},
	}
	err := webobject.RegisterObject(&r.RouterGroup)
	assert.Nil(t, err)

	db.Create(&User{ID: 1})

	var data = map[string]any{
		"birthday": "2022-02-02 11:11:11",
	}
	b, _ := json.Marshal(data)
	req := httptest.NewRequest("PATCH", fmt.Sprintf("/user/%d", 1), bytes.NewReader(b))
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Result().StatusCode)

	var user User
	db.First(&user, 1)
	assert.Equal(t, "2022-02-02 11:11:11 +0000 UTC", user.Birthday.String())
}

func TestObjectRegister(t *testing.T) {
	type User struct {
		UUID     string    `json:"uid" gorm:"primarykey"`
		Name     string    `json:"name" gorm:"size:100"`
		Age      int       `json:"age"`
		Enabled  bool      `json:"enabled"`
		Birthday time.Time `json:"birthday"`
	}

	{
		type Param struct {
			Filterable []string
			Filters    []map[string]any
		}
		type Except struct {
			Total int
		}
		tests := []struct {
			name   string
			params Param
			expect Except
		}{
			{"filter by name and name is filterable",
				Param{
					[]string{"Name"},
					[]map[string]any{{"name": "name", "op": "=", "value": "alice"}},
				},
				Except{1},
			},
			{"filter by name but name is not filterable",
				Param{
					[]string{"Age"},
					[]map[string]any{{"name": "name", "op": "=", "value": "alice"}},
				},
				Except{4},
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {

				db, _ := gorm.Open(sqlite.Open("file::memory:"), nil)
				db.AutoMigrate(User{})

				r := gin.Default()
				r.Use(WithGormDB(db))

				webobject := WebObject{
					Model:        User{},
					FilterFields: tt.params.Filterable,
				}
				err := webobject.RegisterObject(&r.RouterGroup)
				assert.Nil(t, err)

				// Mock data
				{
					db.Create(&User{UUID: "1", Name: "alice", Age: 9})
					db.Create(&User{UUID: "2", Name: "bob", Age: 10})
					db.Create(&User{UUID: "3", Name: "clash", Age: 11})
					db.Create(&User{UUID: "4", Name: "duck", Age: 12})
				}

				data := map[string]any{
					"pos":     0,
					"limit":   5,
					"filters": tt.params.Filters,
				}

				b, _ := json.Marshal(data)
				req := httptest.NewRequest(http.MethodPost, "/user", bytes.NewReader(b))
				w := httptest.NewRecorder()
				r.ServeHTTP(w, req)
				assert.Equal(t, http.StatusOK, w.Result().StatusCode)

				var res QueryResult
				json.Unmarshal(w.Body.Bytes(), &res)
				assert.Equal(t, tt.expect.Total, res.TotalCount)
			})
		}
	}
}

type UnittestUser struct {
	ID   uint   `json:"id" gorm:"primarykey"`
	Name string `json:"name" gorm:"size:100"`
	Age  int    `json:"age"`
}

func initHookTest(t *testing.T) (TestClient, *gorm.DB) {
	db, _ := gorm.Open(sqlite.Open("file::memory:"), nil)
	db.AutoMigrate(UnittestUser{})

	db.Create(&UnittestUser{ID: 1, Name: "alice", Age: 9})
	db.Create(&UnittestUser{ID: 2, Name: "bob", Age: 10})
	db.Create(&UnittestUser{ID: 3, Name: "clash", Age: 11})

	webobject := WebObject{
		Name:         "user",
		Model:        UnittestUser{},
		EditFields:   []string{"Name"},
		FilterFields: []string{"Name, Age"},
		SearchFields: []string{"Name"},
		BeforeCreate: func(db *gorm.DB, ctx *gin.Context, vptr any) error {
			user := (vptr).(*UnittestUser)
			if user.Name == "dangerous" {
				return errors.New("dangerous is not allowed to create")
			}
			return nil
		},
		BeforeRender: func(db *gorm.DB, ctx *gin.Context, vptr any) (any, error) {
			user := (vptr).(*UnittestUser)
			if user.Name != "alice" {
				user.Age = 99
			}
			return vptr, nil
		},
		BeforeDelete: func(db *gorm.DB, ctx *gin.Context, vptr any) error {
			user := (vptr).(*UnittestUser)
			if user.Name == "alice" {
				return errors.New("alice is not allowed to delete")
			}
			return nil
		},
		BeforeUpdate: func(db *gorm.DB, ctx *gin.Context, vptr any, vals map[string]any) error {
			user := (vptr).(*UnittestUser)
			if user.Name == "alice" {
				return errors.New("alice is not allowed to update")
			}
			if vals["name"] == "dangerous" {
				return errors.New("this name is not allowed to update")
			}
			return nil
		},
	}

	r := gin.Default()
	r.Use(WithGormDB(db))
	err := webobject.RegisterObject(&r.RouterGroup)
	assert.Nil(t, err)

	return *NewTestClient(r), db
}

func TestOnRender(t *testing.T) {
	c, _ := initHookTest(t)

	var res QueryResult
	err := c.CallPost("/user", nil, &res)

	assert.Nil(t, err)
	assert.Equal(t, float64(9), res.Items[0].(map[string]any)["age"])
	assert.Equal(t, float64(99), res.Items[1].(map[string]any)["age"])
	assert.Equal(t, float64(99), res.Items[2].(map[string]any)["age"])
}

func TestOnDelete(t *testing.T) {
	c, _ := initHookTest(t)

	err := c.CallDelete("/user/1", nil, nil)
	assert.NotNil(t, err)

	err = c.CallDelete("/user/2", nil, nil)
	assert.Nil(t, err)
}

func TestOnCreate(t *testing.T) {
	c, _ := initHookTest(t)

	err := c.CallPut("/user", UnittestUser{Name: "dangerous"}, nil)
	assert.NotNil(t, err) // dangerous is not allowed to create

	err = c.CallPut("/user", UnittestUser{Name: "notdangerous"}, nil)
	assert.Nil(t, err)
}

func TestOnUpdate(t *testing.T) {
	c, _ := initHookTest(t)

	err := c.CallPatch("/user/1", map[string]any{"name": "notdangerous"}, nil)
	assert.NotNil(t, err) // alice is not allowed to update

	err = c.CallPatch("/user/2", map[string]any{"name": "dangerous"}, nil)
	assert.NotNil(t, err) // this name is not allowed to update

	err = c.CallPatch("/user/2", map[string]any{"name": "notdangerous"}, nil)
	assert.Nil(t, err)
}

func TestQueryViews(t *testing.T) {
	db, _ := gorm.Open(sqlite.Open("file::memory:"), nil)
	db.AutoMigrate(UnittestUser{})

	r := gin.Default()
	r.Use(WithGormDB(db))
	webobject := WebObject{
		Name:         "user",
		Model:        UnittestUser{},
		EditFields:   []string{"Name"},
		FilterFields: []string{"Name, Age"},
		SearchFields: []string{"Name"},
		Views: []QueryView{
			{
				Path:   "names",
				Method: http.MethodGet,
				Prepare: func(db *gorm.DB, ctx *gin.Context) (*gorm.DB, *QueryForm, error) {
					return db, &QueryForm{Limit: -1, ViewFields: []string{"ID", "Name"}}, nil
				},
			},
		},
	}
	err := webobject.RegisterObject(&r.RouterGroup)
	assert.Nil(t, err)

	// create 200 users
	var user [200]UnittestUser
	for i := 0; i < len(user); i++ {
		user[i] = UnittestUser{Name: fmt.Sprintf("user-%d", i), Age: i}
	}
	db.CreateInBatches(&user, len(user))

	client := NewTestClient(r)
	var result QueryResult
	err = client.CallGet("/user/names", nil, &result)
	assert.Nil(t, err)
	assert.Equal(t, 200, result.TotalCount)
	assert.Equal(t, 200, len(result.Items))
	assert.Equal(t, float64(0), result.Items[10].(map[string]any)["age"])
	assert.NotZero(t, result.Items[10].(map[string]any)["name"])
}

func TestPagination(t *testing.T) {
	// Pagination
	{
		db, _ := gorm.Open(sqlite.Open("file::memory:"), nil)
		db.AutoMigrate(UnittestUser{})

		r := gin.Default()
		r.Use(WithGormDB(db))

		RegisterObject(&r.RouterGroup, &WebObject{
			Name:  "user",
			Model: UnittestUser{},
		})

		{
			db.Create(&UnittestUser{Name: "user-1", Age: 1})
			db.Create(&UnittestUser{Name: "user-2", Age: 2})
			db.Create(&UnittestUser{Name: "user-3", Age: 3})
		}

		client := NewTestClient(r)

		var result QueryResult
		client.CallPost("/user", &QueryForm{Pos: 2, Limit: 1, Pagination: true}, &result)
		assert.Equal(t, 3, result.TotalCount)
		assert.Len(t, result.Items, 1)
		assert.Equal(t, "user-2", result.Items[0].(map[string]any)["name"])
	}
	// No Pagination
	{
		db, _ := gorm.Open(sqlite.Open("file::memory:"), nil)
		db.AutoMigrate(UnittestUser{})

		r := gin.Default()
		r.Use(WithGormDB(db))

		RegisterObject(&r.RouterGroup, &WebObject{
			Name:  "user",
			Model: UnittestUser{},
		})

		{
			db.Create(&UnittestUser{Name: "user-1", Age: 1})
			db.Create(&UnittestUser{Name: "user-2", Age: 2})
			db.Create(&UnittestUser{Name: "user-3", Age: 3})
		}

		client := NewTestClient(r)

		var result QueryResult
		client.CallPost("/user", &QueryForm{Pos: 2, Limit: 1}, &result)
		assert.Equal(t, 3, result.TotalCount)
		assert.Len(t, result.Items, 1)
		assert.Equal(t, "user-3", result.Items[0].(map[string]any)["name"])
	}
}

func TestColumnName(t *testing.T) {
	type User struct {
		ID   int64  `json:"id" gorm:"primarykey"`
		Name string `json:"name" gorm:"column:username"`
	}

	db, _ := gorm.Open(sqlite.Open("file::memory:"), nil)
	db.AutoMigrate(User{})

	r := gin.Default()
	r.Use(WithGormDB(db))

	RegisterObject(&r.RouterGroup, &WebObject{
		Name:         "user",
		Model:        User{},
		FilterFields: []string{"Name"},
	})

	{
		db.Create(&User{Name: "user-1"})
		db.Create(&User{Name: "user-2"})
		db.Create(&User{Name: "user-3"})

		var count int64
		db.Model(&User{}).Count(&count)
		assert.Equal(t, int64(3), count)
	}

	client := NewTestClient(r)

	var result QueryResult
	client.CallPost("/user", &QueryForm{
		Filters: []Filter{
			{Name: "name", Op: "=", Value: "user-2"},
		}}, &result)
	assert.Equal(t, 1, result.TotalCount)
	assert.Len(t, result.Items, 1)
	assert.Equal(t, "user-2", result.Items[0].(map[string]any)["name"])
}

func TestCreateTime(t *testing.T) {
	type User struct {
		ID       int64     `json:"id" gorm:"primarykey"`
		Birthday time.Time `json:"birthday"`
	}

	db, _ := gorm.Open(sqlite.Open("file::memory:"), nil)
	db.AutoMigrate(User{})

	r := gin.Default()
	r.Use(WithGormDB(db))

	err := RegisterObject(&r.RouterGroup, &WebObject{
		Name:       "user",
		Model:      User{},
		EditFields: []string{"Birthday"},
	})
	assert.Nil(t, err)

	{
		// RFC3339: 2006-01-02T15:04:05Z07:00
		// ISO8601: 2021-01-01T00:00:00Z
		json := `{"id":2, "birthday":"2021-01-01T00:00:00Z"}`
		req := httptest.NewRequest(http.MethodPut, "/user", strings.NewReader(json))
		req.Header.Set("Content-Type", "application/json") // !important
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Result().StatusCode)

		val, _ := GetByID[User](db, 2)
		assert.Equal(t, int64(2), val.ID)
		assert.Equal(t, "2021-01-01 00:00:00 +0000 UTC", val.Birthday.String())
	}

	// TODO: other time format
	{
		// now := time.Now()
		// str := now.Format(time.RFC3339) // 2024-04-16T15:41:05+08:00

		// json := `{"id":1, "birthday": "` + str + `"}`
		// req := httptest.NewRequest(http.MethodPut, "/user", strings.NewReader(json))
		// req.Header.Set("Content-Type", "application/json") // !important
		// w := httptest.NewRecorder()
		// r.ServeHTTP(w, req)
		// assert.Equal(t, http.StatusOK, w.Result().StatusCode)

		// now := time.Now()
		// b, _ := json.Marshal(User{ID: 1, Birthday: now})
		// req := httptest.NewRequest(http.MethodPut, "/user", bytes.NewReader(b))
		// w := httptest.NewRecorder()
		// r.ServeHTTP(w, req)
		// assert.Equal(t, http.StatusOK, w.Result().StatusCode)

		// val, _ := GetByID[User](db, 1)
		// assert.Equal(t, int64(1), val.ID)
		// assert.Equal(t, now.Format(time.RFC3339), val.Birthday.Format(time.RFC3339))
	}
}

func TestEditTime(t *testing.T) {
	type User struct {
		ID       int64     `json:"id" gorm:"primarykey"`
		Birthday time.Time `json:"birthday"`
	}

	db, _ := gorm.Open(sqlite.Open("file::memory:"), nil)
	db.AutoMigrate(User{})

	r := gin.Default()
	r.Use(WithGormDB(db))

	err := RegisterObject(&r.RouterGroup, &WebObject{
		Name:       "user",
		Model:      User{},
		EditFields: []string{"Birthday"},
	})
	assert.Nil(t, err)

	db.Create(&User{ID: 1, Birthday: time.Now()})

	{
		// RFC3339: 2006-01-02T15:04:05Z07:00
		// ISO8601: 2021-01-01T00:00:00Z
		json := `{"birthday":"2021-01-01T00:00:00Z"}`
		req := httptest.NewRequest("PATCH", "/user/1", strings.NewReader(json))
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Result().StatusCode)

		val, _ := GetByID[User](db, 1)
		assert.Equal(t, int64(1), val.ID)
		assert.Equal(t, "2021-01-01 00:00:00 +0000 UTC", val.Birthday.String())
	}

	{
		// input type="datetime-local" 2006-01-02T15:04
		json := `{"birthday":"2023-06-13T01:17"}`
		req := httptest.NewRequest("PATCH", "/user/1", strings.NewReader(json))
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Result().StatusCode)

		val, _ := GetByID[User](db, 1)
		assert.Equal(t, int64(1), val.ID)
		assert.Equal(t, "2023-06-13 01:17:00 +0000 UTC", val.Birthday.String())
	}
}

func TestPreloadGet(t *testing.T) {
	type Company struct {
		ID   int64 `gorm:"primarykey"`
		Name string
	}

	type User struct {
		ID        int64 `gorm:"primarykey"`
		CompanyID int64
		Company   Company `gorm:"foreignKey:CompanyID;references:ID"`
	}

	db, _ := gorm.Open(sqlite.Open("file::memory:"), nil)
	db.AutoMigrate(User{})

	r := gin.Default()
	r.Use(WithGormDB(db))

	err := RegisterObject(&r.RouterGroup, &WebObject{
		Name:  "user",
		Model: User{},
	})
	assert.Nil(t, err)

	db.Create(&User{ID: 1, Company: Company{ID: 1, Name: "company-1"}})
	db.Create(&User{ID: 2, Company: Company{ID: 2, Name: "company-2"}})

	client := NewTestClient(r)

	var u1 User
	err = client.CallGet("/user/1", nil, &u1)
	assert.Nil(t, err)

	assert.Equal(t, int64(1), u1.ID)
	assert.Equal(t, int64(1), u1.CompanyID)
	assert.Equal(t, int64(1), u1.Company.ID)
	assert.Equal(t, "company-1", u1.Company.Name)

	// db preload
	{
		var u2 User
		db.Preload("Company").Take(&u2)
		assert.Equal(t, int64(1), u2.ID)
		assert.Equal(t, int64(1), u2.CompanyID)
		assert.Equal(t, int64(1), u2.Company.ID)
		assert.Equal(t, "company-1", u2.Company.Name)
	}
}

func TestPreloadQuery(t *testing.T) {
	type Company struct {
		ID   int64 `gorm:"primarykey"`
		Name string
	}

	type User struct {
		ID        int64 `gorm:"primarykey"`
		CompanyID int64
		Company   Company `gorm:"foreignKey:CompanyID;references:ID"`
	}

	db, _ := gorm.Open(sqlite.Open("file::memory:"), nil)
	db.AutoMigrate(User{})
	r := gin.Default()
	r.Use(WithGormDB(db))

	err := RegisterObject(&r.RouterGroup, &WebObject{
		Name:  "user",
		Model: User{},
	})
	assert.Nil(t, err)

	db.Create(&User{ID: 1, Company: Company{ID: 1, Name: "company-1"}})
	db.Create(&User{ID: 2, Company: Company{ID: 2, Name: "company-2"}})

	client := NewTestClient(r)

	var us1 QueryResult
	err = client.CallPost("/user", nil, &us1)
	assert.Nil(t, err)
	assert.Equal(t, 2, len(us1.Items))
	fmt.Println((us1.Items[0].(map[string]any)["Company"]).(map[string]any)) // map[ID:1 Name:company-1]

	assert.Equal(t, float64(1), (us1.Items[0].(map[string]any)["Company"]).(map[string]any)["ID"])
	assert.Equal(t, "company-1", (us1.Items[0].(map[string]any)["Company"]).(map[string]any)["Name"])
	assert.Equal(t, float64(2), (us1.Items[1].(map[string]any)["Company"]).(map[string]any)["ID"])
	assert.Equal(t, "company-2", (us1.Items[1].(map[string]any)["Company"]).(map[string]any)["Name"])

	// db preload
	{
		var us2 []User
		db.Preload("Company").Find(&us2)
		assert.Len(t, us2, 2)
		assert.Equal(t, int64(1), us2[0].Company.ID)
		assert.Equal(t, "company-1", us2[0].Company.Name)
		assert.Equal(t, int64(2), us2[1].Company.ID)
		assert.Equal(t, "company-2", us2[1].Company.Name)
	}
}
