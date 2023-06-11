package main

import (
	"errors"
	"flag"
	"math/rand"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/restsend/gormpher"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type Product struct {
	UUID      string    `json:"id" gorm:"primarykey"`
	GroupID   int       `json:"-"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
	Enabled   bool      `json:"enabled"`
}

type User struct {
	ID        uint       `json:"id" gorm:"primarykey"`
	CreatedAt time.Time  `json:"createdAt"`
	UpdatedAt time.Time  `json:"updatedAt"`
	Name      string     `json:"name"`
	Age       int        `json:"age"`
	Enabled   bool       `json:"enabled"`
	LastLogin *time.Time `json:"lastLogin,omitempty"`
}

func main() {
	var dsn string
	var addr string

	flag.StringVar(&dsn, "n", "", "DB DSN")
	flag.StringVar(&addr, "a", ":8890", "Api Server Addr")
	flag.Parse()

	db, _ := gorm.Open(sqlite.Open(dsn), nil)
	db.AutoMigrate(Product{}, User{})

	r := gin.Default()

	objs := GetWebObjects(db)

	// visit API: http://localhost:8890/api
	gormpher.RegisterObjects(r, objs)
	// visit Admin: http://localhost:8890/admin
	gormpher.RegisterObjectsWithAdmin(r.Group("admin"), objs)

	r.Run(addr)
}

func GetWebObjects(db *gorm.DB) []gormpher.WebObject {
	return []gormpher.WebObject{
		// Basic Demo
		// Check API File: user.http
		// PUT 		http://localhost:8890/user
		// GET 		http://localhost:8890/user/:key
		// PATCH	http://localhost:8890/user/:key
		// POST 	http://localhost:8890/user
		// DELETE http://localhost:8890/user/:key
		// DELETE http://localhost:8890/user
		{
			Name:        "user",
			Model:       &User{},
			Searchables: []string{"Name", "Enabled"},
			Editables:   []string{"Name", "Age", "Enabled"},
			Filterables: []string{"Name", "CreatedAt", "Age", "Enabled"},
			Orderables:  []string{"CreatedAt", "Age", "Enabled"},
			GetDB: func(ctx *gin.Context, isCreate bool) *gorm.DB {
				return db
			},
		},
		// Advanced Demo
		// Check API File: product.http
		// PUT 		http://localhost:8890/product
		// GET 		http://localhost:8890/product/:key
		// PATCH	http://localhost:8890/product/:key
		// POST 	http://localhost:8890/product
		// DELETE http://localhost:8890/product/:key
		// DELETE http://localhost:8890/product
		{
			Name:        "product",
			Model:       &Product{},
			Searchables: []string{"Name"},
			Editables:   []string{"Name", "Enabled"},
			Filterables: []string{"Name", "CreatedAt", "Enabled"},
			Orderables:  []string{"CreatedAt"},
			GetDB: func(c *gin.Context, isCreate bool) *gorm.DB {
				return db
			},
			OnCreate: func(ctx *gin.Context, vptr any, vals map[string]any) error {
				p := (vptr).(*Product)
				p.UUID = MockUUID(8)
				p.GroupID = rand.Intn(5)
				return nil
			},
			OnDelete: func(ctx *gin.Context, vptr any) error {
				p := (vptr).(*Product)
				if p.Enabled {
					return errors.New("product is enabled, can not delete")
				}
				return nil
			},
			// Custom Query View
			// GET http://localhost:8890/product/all_enabled
			Views: []gormpher.QueryView{
				{
					Name:   "all_enabled",
					Method: "GET",
					Prepare: func(db *gorm.DB, c *gin.Context) (*gorm.DB, *gormpher.QueryForm, error) {
						// SELECT (id, name) FROM products WHERE enabled = true
						queryForm := &gormpher.QueryForm{
							Limit: -1,
							Filters: []gormpher.Filter{
								{Name: "enabled", Op: "=", Value: true}, // JSON format
							},
							ViewFields: []string{"UUID", "Name"},
						}
						return db, queryForm, nil
					},
				},
			},
		},
	}
}

func MockUUID(n int) string {
	source := []rune("0123456789abcdefghijklmnopqrstuvwxyz")
	b := make([]rune, n)
	for i := range b {
		b[i] = source[rand.Intn(len(source))]
	}
	return string(b)
}
