package gormpher

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

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

func TestHandleGet(t *testing.T) {
	r := gin.Default()
	db, _ := gorm.Open(sqlite.Open("file::memory:"), nil)
	db.AutoMigrate(&testUser{})

	r.GET("/user/:key", func(ctx *gin.Context) {
		HandleGet(ctx, db, reflect.TypeOf(testUser{}), func(ctx *gin.Context, vptr any) error {
			v := vptr.(*testUser)
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

func TestHandleCreate(t *testing.T) {
	r := gin.Default()
	db, _ := gorm.Open(sqlite.Open("file::memory:"), nil)
	db.AutoMigrate(&testUser{})

	r.PUT("/user", func(ctx *gin.Context) {
		HandleCreate(ctx, db, reflect.TypeOf(testUser{}), func(ctx *gin.Context, vptr any, vals map[string]any) error {
			v := vptr.(*testUser)
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
		HandleDelete(ctx, db, reflect.TypeOf(testUser{}), func(ctx *gin.Context, vptr any) error {
			v := vptr.(*testUser)
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
