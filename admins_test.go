package gormpher

import (
	"reflect"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestConvertTypeToJS(t *testing.T) {
	testCases := []struct {
		input reflect.Kind
		want  string
	}{
		{reflect.String, "string"},
		{reflect.Int, "number"},
		{reflect.Float64, "number"},
		{reflect.Bool, "boolean"},
		{reflect.Struct, "object"},
		{reflect.Map, "object"},
		{reflect.Slice, "object"},
		{reflect.Ptr, "any"},
		{reflect.Interface, "any"},
		{reflect.Complex128, "unknown"},
	}

	for _, tc := range testCases {
		got := convertGoTypeToJS(tc.input)
		if got != tc.want {
			t.Errorf("convertTypeToJS(%v) = %v, want %v", tc.input, got, tc.want)
		}
	}
}

func TestWoToAo(t *testing.T) {
	type User struct {
		ID        uint   `json:"uid" gorm:"primarykey"`
		Name      string `json:"name" gorm:"size:100"`
		Age       int
		Body      string     `json:"-" gorm:"-"`
		LastLogin *time.Time `json:"lastLogin,omitempty"`
	}

	wo := WebObject{
		Model:        User{},
		EditFields:   []string{"Name"},
		FilterFields: []string{"Name", "Age", "ID"},
		SearchFields: []string{"Name", "ID"},
		OrderFields:  []string{"LastLogin"},
	}

	ao := woToAo(wo)
	assert.Equal(t, 1, len(ao.Orders))
	assert.Equal(t, 1, len(ao.Edits))
	assert.Equal(t, 3, len(ao.Filters))
	assert.Equal(t, 2, len(ao.Searchs))

	assert.Equal(t, "name", ao.Edits[0])
	assert.Equal(t, "lastLogin", ao.Orders[0])
}

func TestRegisterSingleObject(t *testing.T) {
	type User struct {
		ID        uint   `json:"uid" gorm:"primarykey"`
		Name      string `json:"name" gorm:"size:100"`
		Age       int
		Body      string     `json:"-" gorm:"-"`
		LastLogin *time.Time `json:"lastLogin,omitempty"`
	}

	type Product struct {
		UUID string `json:"uuid" gorm:"primaryKey"`
		Name string `json:"name"`
	}

	db, _ := gorm.Open(sqlite.Open("file::memory:"), nil)

	ar := gin.Default().Group("admin")

	wo1 := WebObject{
		Path:         "user",
		Model:        User{},
		EditFields:   []string{"Name"},
		FilterFields: []string{"Name", "Age", "ID"},
		SearchFields: []string{"Name", "ID"},
		OrderFields:  []string{"LastLogin"},
		GetDB: func(c *gin.Context, isCreate bool) *gorm.DB {
			return db
		},
	}

	wo2 := WebObject{
		Path:  "product",
		Model: Product{},
		GetDB: func(c *gin.Context, isCreate bool) *gorm.DB {
			return db
		},
	}

	am := AdminManager{}
	am.RegisterObject(ar, wo1)
	am.RegisterObject(ar, wo2)

	assert.Equal(t, 2, len(am.AdminObjects))
	assert.Equal(t, 2, len(am.Names))
}

func TestRegisterObjects(t *testing.T) {
	type User struct {
		ID        uint   `json:"uid" gorm:"primarykey"`
		Name      string `json:"name" gorm:"size:100"`
		Age       int
		Body      string     `json:"-" gorm:"-"`
		LastLogin *time.Time `json:"lastLogin,omitempty"`
	}

	db, _ := gorm.Open(sqlite.Open("file::memory:"), nil)
	db.AutoMigrate(User{})

	r := gin.Default()

	wos := []WebObject{
		{
			Path:         "user1",
			Model:        User{},
			EditFields:   []string{"Name"},
			FilterFields: []string{"Name", "Age", "ID"},
			SearchFields: []string{"Name", "ID"},
			OrderFields:  []string{"LastLogin"},
			GetDB: func(c *gin.Context, isCreate bool) *gorm.DB {
				return db
			},
		},
		{
			Path:         "user2",
			Model:        User{},
			EditFields:   []string{"Name"},
			FilterFields: []string{"Name", "Age", "ID"},
			SearchFields: []string{"Name", "ID"},
			OrderFields:  []string{"LastLogin"},
			GetDB: func(c *gin.Context, isCreate bool) *gorm.DB {
				return db
			},
		},
	}

	RegisterObjectsWithAdmin(r.Group("admin"), wos)
}
