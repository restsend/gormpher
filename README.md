Gormpher - Generate restful APIs by defining struct model, based on Gin and Gorm
====

# Quick Start

```go
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
 var addr string

 flag.StringVar(&dsn, "n", "", "DB DSN")
 flag.StringVar(&addr, "a", ":8890", "Api Server Addr")
 flag.Parse()

 db, _ := gorm.Open(sqlite.Open(dsn), nil)
 db.AutoMigrate(Product{}, User{})

 r := gin.Default()

 objs := GetWebObjects(db)

 // visit API: http://localhost:8890/api
 gormpher.RegisterObjects(r.Group("api"), objs)
 // visit Admin: http://localhost:8890/admin
 gormpher.RegisterObjectsWithAdmin(r.Group("admin"), objs)

 r.Run(addr)
}

func GetWebObjects(db *gorm.DB) []gormpher.WebObject {
 return []gormpher.WebObject{
  // Check API File: product.http
  // PUT   http://localhost:8890/user
  // GET   http://localhost:8890/user/:key
  // PATCH http://localhost:8890/user/:key
  // POST  http://localhost:8890/user
  // DELETE http://localhost:8890/user/:key
  // DELETE http://localhost:8890/user
  {
   Name:      "product",
   Model:     &Product{},
   Searchs:   []string{"Name"},
   Editables: []string{"Name", "Enabled"},
   Filters:   []string{"Name", "CreatedAt", "Enabled"},
   Orders:    []string{"CreatedAt"},
   GetDB: func(c *gin.Context, isCreate bool) *gorm.DB {
    return db
   },
   Init: func(ctx *gin.Context, vptr any) error {
    p := (vptr).(*Product)
    p.UUID = MockUUID(8)
    p.GroupID = rand.Intn(5)
    return nil
   },
  },
  // Check API File: user.http
  // PUT   http://localhost:8890/user
  // GET   http://localhost:8890/user/:key
  // PATCH http://localhost:8890/user/:key
  // POST  http://localhost:8890/user
  // DELETE http://localhost:8890/user/:key
  // DELETE http://localhost:8890/user
  {
   Name:      "user",
   Model:     &User{},
   Searchs:   []string{"Name", "Enabled"},
   Editables: []string{"Name", "Age", "Enabled"},
   Filters:   []string{"Name", "CreatedAt", "Age", "Enabled"},
   Orders:    []string{"CreatedAt", "Age", "Enabled"},
   GetDB: func(ctx *gin.Context, isCreate bool) *gorm.DB {
    return db
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
```

Run the project and visit [http://localhost:8890/admin](http://localhost:8890/admin) , you can see a web interface for administrator.

![image](https://user-images.githubusercontent.com/48195906/227763154-c13addc4-28bf-4572-8fce-1110e6cdae8e.png)
