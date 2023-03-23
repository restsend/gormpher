package main

import (
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
	GroupID   int       `json:"groupId"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
	Enabled   bool      `json:"enabled"`
}

func (p *Product) BeforeCreate(tx *gorm.DB) (err error) {
	if groupId, ok := tx.Get("groupId"); ok {
		p.GroupID = groupId.(int)
	}
	p.UUID = MockUUID(8)
	return
}

type User struct {
	ID        uint       `json:"id" gorm:"primarykey"`
	CreatedAt time.Time  `json:"createdAt"`
	UpdatedAt time.Time  `json:"updatedAt"`
	Name      string     `json:"name"`
	Age       int        `json:"age"`
	Enabled   bool       `json:"enabled"`
	LastLogin *time.Time `json:"lastLogin"`
}

func main() {
	var dsn string
	var webServerAddr string
	var apiServerAddr string

	flag.StringVar(&dsn, "n", "", "DB DSN")
	flag.StringVar(&webServerAddr, "web", ":5000", "Web Server Addr")
	flag.StringVar(&apiServerAddr, "api", ":8890", "Api Server Addr")
	flag.Parse()

	db, _ := gorm.Open(sqlite.Open(dsn), nil)
	db.AutoMigrate(Product{}, User{})

	r := gin.Default()
	RegisterWebObjectHandler(r, db, webServerAddr)

	// visit Admin: http://localhost:5000/
	// visit API: http://localhost:8890/
	r.Run(apiServerAddr)
}

func RegisterWebObjectHandler(r gin.IRoutes, db *gorm.DB, webServerAddr string) {
	objs := []gormpher.WebObject{
		// Check API File: product.http
		// PUT 		http://localhost:8890/user
		// GET 		http://localhost:8890/user/:key
		// PATCH	http://localhost:8890/user/:key
		// POST 	http://localhost:8890/user
		// DELETE http://localhost:8890/user/:key
		// DELETE http://localhost:8890/user
		{
			Model:     &Product{},
			Searchs:   []string{"Name"},
			Editables: []string{"Name", "Enabled"},
			Filters:   []string{"Name", "CreatedAt", "Enabled"},
			Orders:    []string{"CreatedAt"},
			GetDB: func(c *gin.Context, isCreate bool) *gorm.DB {
				return db.Set("groupId", rand.Intn(5)) // Mock groupId
			},
		},
		// Check API File: user.http
		// PUT 		http://localhost:8890/user
		// GET 		http://localhost:8890/user/:key
		// PATCH	http://localhost:8890/user/:key
		// POST 	http://localhost:8890/user
		// DELETE http://localhost:8890/user/:key
		// DELETE http://localhost:8890/user
		{
			Model:     &User{},
			Searchs:   []string{"Name", "Enabled"},
			Editables: []string{"Name", "Age", "Enabled"},
			Filters:   []string{"Name", "CreatedAt", "Age", "Enabled"},
			Orders:    []string{"CreatedAt", "Age", "Enabled"},
			GetDB: func(c *gin.Context, isCreate bool) *gorm.DB {
				return db
			},
		},
	}

	// Register API
	gormpher.RegisterObjects(r, objs)

	// Register Admin (Optional)
	gormpher.RegisterObjectsWithAdmin(objs, webServerAddr)
}

func MockUUID(n int) string {
	source := []rune("0123456789abcdefghijklmnopqrstuvwxyz")
	b := make([]rune, n)
	for i := range b {
		b[i] = source[rand.Intn(len(source))]
	}
	return string(b)
}
