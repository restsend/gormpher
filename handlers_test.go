package gormpher

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type testUser struct {
	ID      uint   `json:"id" gorm:"primarykey"`
	Name    string `json:"name"`
	Age     int    `json:"age"`
	Enabled bool   `json:"enabled"`
}

func TestHandleEdit(t *testing.T) {
	r := gin.Default()
	db, _ := gorm.Open(sqlite.Open("file::memory:"), nil)
	db.AutoMigrate(&testUser{})

	db.Create(&testUser{ID: 1, Name: "demo", Age: 12, Enabled: false})

	r.PATCH("/user/:key", func(ctx *gin.Context) {
		HandleEdit[testUser](ctx, db, []string{"Name", "Age", "Enabled"}, nil)
	})

	testCases := []struct {
		name           string
		requestBody    interface{}
		expectedStatus int
		expectedName   string
		expectedAge    int
		expectedEnable bool
	}{
		{
			name:           "update name and age",
			requestBody:    testUser{Name: "updated", Age: 12, Enabled: false},
			expectedStatus: http.StatusOK,
			expectedName:   "updated",
			expectedAge:    12,
			expectedEnable: false,
		},
		{
			name:           "update name, age and enabled",
			requestBody:    testUser{Name: "updated2", Age: 22, Enabled: true},
			expectedStatus: http.StatusOK,
			expectedName:   "updated2",
			expectedAge:    22,
			expectedEnable: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			b, err := json.Marshal(tc.requestBody)
			assert.Nil(t, err)

			req := httptest.NewRequest(http.MethodPatch, "/user/1", bytes.NewReader(b))
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)
			assert.Equal(t, tc.expectedStatus, w.Result().StatusCode)

			var u testUser
			db.Take(&u, 1)
			assert.Equal(t, tc.expectedName, u.Name)
			assert.Equal(t, tc.expectedAge, u.Age)
			assert.Equal(t, tc.expectedEnable, u.Enabled)
		})
	}
}

func TestHandleEdit2(t *testing.T) {
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
			// TODO: 能否将数据库中的值更新为 null?
			{"ptr_case_2",
				Param{
					1, map[string]any{
						"ptrTime": nil,
					}},
				Except{http.StatusOK},
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {

				db, _ := gorm.Open(sqlite.Open("file::memory:"), nil)
				db.AutoMigrate(User{})

				r := gin.Default()
				r.PATCH("/user/:key", func(ctx *gin.Context) {
					HandleEdit[User](ctx, db, []string{"Name", "Age", "Enabled", "Birthday", "PtrTime"}, nil)
				})

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

func TestHandleEditFunc(t *testing.T) {
	type User struct {
		ID  uint `json:"id" gorm:"primarykey"`
		Age int  `json:"age"`
	}

	r := gin.Default()
	db, _ := gorm.Open(sqlite.Open("file::memory:"), nil)
	db.AutoMigrate(&User{})

	r.PATCH("/user/:key", func(ctx *gin.Context) {
		HandleEdit[User](ctx, db, []string{"Age"},
			func(ctx *gin.Context, u *User, vals map[string]any) error {
				if vals["age"].(float64) > 100 {
					return errors.New("age too old")
				}
				return nil
			})
	})

	{
		db.Create(&User{ID: 1, Age: 12})
	}

	// can edit
	{
		b, _ := json.Marshal(User{Age: 99})
		req := httptest.NewRequest("PATCH", "/user/1", bytes.NewReader(b))
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Result().StatusCode)
	}

	// can not edit
	{
		b, _ := json.Marshal(User{Age: 101})
		req := httptest.NewRequest("PATCH", "/user/1", bytes.NewReader(b))
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		assert.Equal(t, http.StatusBadRequest, w.Result().StatusCode)
	}
}

func TestExecuteEdit(t *testing.T) {
	db, _ := gorm.Open(sqlite.Open("file::memory:"), nil)
	db.AutoMigrate(&testUser{})

	db.Create(&testUser{Name: "alice", Age: 12})

	v, err := ExecuteEdit[testUser](db, 1, map[string]any{
		"name": "bob",
		"age":  99,
	})
	assert.Equal(t, nil, err)
	assert.Equal(t, "bob", v.Name)
	assert.Equal(t, 99, v.Age)

	v, err = ExecuteEdit[testUser](db, "1", map[string]any{
		"name": "clash",
		"age":  66,
	})
	assert.Equal(t, nil, err)
	assert.Equal(t, "clash", v.Name)
	assert.Equal(t, 66, v.Age)
}

func TestHandleCreate(t *testing.T) {
	r := gin.Default()
	db, _ := gorm.Open(sqlite.Open("file::memory:"), nil)
	db.AutoMigrate(&testUser{})

	r.PUT("/user", func(ctx *gin.Context) {
		HandleCreate[testUser](ctx, db, func(ctx *gin.Context, v *testUser, vals map[string]any) error {
			if v.Age >= 100 {
				return errors.New("age too old")
			}
			return nil
		})
	})

	// onCreate
	{
		u := testUser{Name: "name", Age: 101}
		b, _ := json.Marshal(u)
		req := httptest.NewRequest(http.MethodPut, "/user", bytes.NewReader(b))
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
	}

	testCases := []struct {
		name string
		age  int
	}{
		{name: "Alice", age: 25},
		{name: "Bob", age: 30},
		{name: "Charlie", age: 35},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			u := testUser{Name: tc.name, Age: tc.age}
			b, _ := json.Marshal(u)
			req := httptest.NewRequest(http.MethodPut, "/user", bytes.NewReader(b))
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			assert.Equal(t, http.StatusOK, w.Code)

			var val testUser
			err := json.Unmarshal(w.Body.Bytes(), &val)
			assert.Nil(t, err)
			assert.Equal(t, tc.name, val.Name)
			assert.Equal(t, tc.age, val.Age)
		})
	}
}

func TestHandleDelete(t *testing.T) {
	r := gin.Default()
	db, _ := gorm.Open(sqlite.Open("file::memory:"), nil)
	db.AutoMigrate(&testUser{})

	r.DELETE("/user/:key", func(ctx *gin.Context) {
		HandleDelete[testUser](ctx, db, func(ctx *gin.Context, v *testUser) error {
			if v.Name == "admin" {
				return errors.New("cannot delete admin")
			}
			return nil
		})
	})

	db.Create(&testUser{ID: 1, Name: "demo", Age: 12})
	db.Create(&testUser{ID: 2, Name: "admin", Age: 20})

	// exist, delete
	{
		req := httptest.NewRequest(http.MethodDelete, "/user/1", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "true", w.Body.String())

		var count int64
		db.Model(testUser{}).Count(&count)
		assert.Equal(t, int64(1), count)
	}

	// not exist, no error
	{
		req := httptest.NewRequest(http.MethodDelete, "/user/99", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	}

	// onDelete, admin is not allowed to delete
	{
		req := httptest.NewRequest(http.MethodDelete, "/user/2", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
	}
}

func TestHandleGet(t *testing.T) {
	r := gin.Default()
	db, _ := gorm.Open(sqlite.Open("file::memory:"), nil)
	db.AutoMigrate(&testUser{})

	r.GET("/user/:key", func(ctx *gin.Context) {
		HandleGet[testUser](ctx, db, func(ctx *gin.Context, v *testUser) error {
			v.Name = "[Name]" + v.Name
			return nil
		})
	})

	db.Create(&testUser{Name: "demo", Age: 12})

	// exist
	{
		req := httptest.NewRequest("GET", "/user/1", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var val testUser
		err := json.Unmarshal(w.Body.Bytes(), &val)
		assert.Nil(t, err)
		assert.Equal(t, "[Name]demo", val.Name)
		assert.Equal(t, 12, val.Age)
	}

	// not exist
	{
		req := httptest.NewRequest("GET", "/user/2", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
	}
}

func TestHandleQueryNil(t *testing.T) {
	r := gin.Default()
	db, _ := gorm.Open(sqlite.Open("file::memory:"), nil)
	db.AutoMigrate(&testUser{})

	r.POST("/user", func(ctx *gin.Context) {
		HandleQuery[testUser](ctx, db, nil)
	})

	db.Create(&testUser{Name: "alice", Age: 12})
	db.Create(&testUser{Name: "bob", Age: 13})
	db.Create(&testUser{Name: "clash", Age: 14})

	client := NewTestClient(r)
	{
		var result QueryResult[[]testUser]
		err := client.CallPost("/user", nil, &result)
		assert.Nil(t, err)
		assert.Equal(t, 3, result.Total)
		assert.Equal(t, "alice", result.Items[0].Name)
		assert.Equal(t, 12, result.Items[0].Age)
	}
}

func TestHandleQuery(t *testing.T) {
	r := gin.Default()
	db, _ := gorm.Open(sqlite.Open("file::memory:"), nil)
	db.AutoMigrate(&testUser{})

	r.POST("/user", func(ctx *gin.Context) {
		HandleQuery[testUser](ctx, db, &QueryOption{
			Filterables: []string{"Name", "Age", "Enabled"},
			Searchables: []string{"Name"},
			Orderables:  []string{"Age"},
		})
	})

	db.Create(&testUser{Name: "alice", Age: 12})
	db.Create(&testUser{Name: "bob", Age: 13})
	db.Create(&testUser{Name: "clash", Age: 14})

	client := NewTestClient(r)
	// basic
	{
		var result QueryResult[[]testUser]
		err := client.CallPost("/user", &QueryForm{}, &result)
		assert.Nil(t, err)
		assert.Equal(t, 3, result.Total)
	}
	// filter
	{
		var result QueryResult[[]testUser]
		err := client.CallPost("/user", &QueryForm{
			Filters: []Filter{
				{Name: "name", Op: "=", Value: "alice"},
			},
		}, &result)
		assert.Nil(t, err)
		assert.Equal(t, 1, result.Total)
		assert.Equal(t, "alice", result.Items[0].Name)
	}
	// search
	{
		var result QueryResult[[]testUser]
		err := client.CallPost("/user", &QueryForm{Keyword: "bob"}, &result)
		assert.Nil(t, err)
		assert.Equal(t, 1, result.Total)
		assert.Equal(t, "bob", result.Items[0].Name)
	}
	// order
	{
		// desc
		var result QueryResult[[]testUser]
		err := client.CallPost("/user", &QueryForm{
			Orders: []Order{
				{Name: "age", Op: "desc"},
			},
		}, &result)
		assert.Nil(t, err)
		assert.Equal(t, 3, result.Total)
		assert.Equal(t, "clash", result.Items[0].Name)
		// asc
		err = client.CallPost("/user", &QueryForm{
			Orders: []Order{
				{Name: "age", Op: "asc"},
			},
		}, &result)
		assert.Nil(t, err)
		assert.Equal(t, 3, result.Total)
		assert.Equal(t, "alice", result.Items[0].Name)
	}
}

func TestExecuteQuery(t *testing.T) {
	db, _ := gorm.Open(sqlite.Open("file::memory:"), nil)
	db.AutoMigrate(&testUser{})

	db.Create(&testUser{Name: "alice", Age: 12})
	db.Create(&testUser{Name: "bob", Age: 13})
	db.Create(&testUser{Name: "clash", Age: 14})

	list, count, err := ExecuteQuery[testUser](db, QueryForm{
		Pos:   0,
		Limit: 10,
		Filters: []Filter{
			{Name: "name", Op: "=", Value: "alice"},
		},
	}, false)
	assert.Nil(t, err)
	assert.Equal(t, 1, count)
	assert.Equal(t, "alice", list[0].Name)
}
