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

type User struct {
	ID        uint       `json:"id" gorm:"primarykey"`
	CreatedAt time.Time  `json:"createdAt"`
	UpdatedAt time.Time  `json:"updatedAt"`
	Name      string     `json:"name"`
	Age       int        `json:"age"`
	Enabled   bool       `json:"enabled"`
	LastLogin *time.Time `json:"lastLogin"`
}

type Product struct {
	UUID      string    `json:"id" gorm:"primarykey"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
	GroupID   int       `json:"-"`
	Group     Group     `json:"group" gorm:"foreignKey:GroupID;references:ID"` // association
	Name      string    `json:"name"`
	Enabled   bool      `json:"enabled"`
}

type Group struct {
	ID   uint   `json:"id" gorm:"primarykey"`
	Name string `json:"name"`
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
	// visit Admin: http://localhost:8890/admin/v1
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
			Name:         "user",
			Model:        &User{},
			SearchFields: []string{"Name", "Enabled"},
			EditFields:   []string{"Name", "Age", "Enabled", "LastLogin"},
			FilterFields: []string{"Name", "CreatedAt", "UpdatedAt", "Age", "Enabled"},
			OrderFields:  []string{"CreatedAt", "Age", "Enabled"},
			GetDB:        func(ctx *gin.Context, isCreate bool) *gorm.DB { return db },
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
			Name:         "product",
			Model:        &Product{},
			SearchFields: []string{"Name"},
			EditFields:   []string{"Name", "Enabled", "Model"},
			FilterFields: []string{"Name", "CreatedAt", "Enabled"},
			OrderFields:  []string{"CreatedAt"},
			GetDB:        func(c *gin.Context, isCreate bool) *gorm.DB { return db },
			BeforeCreate: func(ctx *gin.Context, vptr any, vals map[string]any) error {
				p := (vptr).(*Product)
				p.UUID = MockUUID(8)

				// create group
				group := Group{Name: "group" + MockUUID(4)}
				if err := db.Create(&group).Error; err != nil {
					return err
				}

				p.GroupID = int(group.ID)
				return nil
			},
			BeforeDelete: func(ctx *gin.Context, vptr any) error {
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
