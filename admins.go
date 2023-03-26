package gormpher

import (
	"embed"
	"fmt"
	"log"
	"net/http"
	"path"
	"reflect"
	"strings"

	"github.com/gin-gonic/gin"
)

type AdminObject struct {
	WebObject *WebObject
	Searchs   []string
	Filters   []string
	Edits     []string
	Orders    []string

	// Map fields to jsonTag. such as:
	// UUID string `json:"id"` => {"UUID": "id"}
	fieldsToJsons map[string]string
}

type AdminManager struct {
	AdminBojects []*AdminObject
	Names        []string
}

//go:embed web/dist/index.html web/dist/assets/*
var assets embed.FS

//go:embed web/dist/index.html
var indexHTML string

func RegisterObjectsWithAdmin(r *gin.RouterGroup, objs []WebObject) error {
	m := AdminManager{}

	for idx := range objs {
		obj := &objs[idx]
		if err := obj.RegisterObject((gin.IRoutes)(r)); err != nil {
			log.Fatalf("RegisterObject [%s] fail %v\n", obj.Name, err)
		}

		m.AdminBojects = append(m.AdminBojects, &AdminObject{
			WebObject: obj,
		})
	}

	if err := parseWebObjects(&m); err != nil {
		return err
	}

	go RegisterAdmin(r, &m)
	return nil
}

func RegisterAdmin(r *gin.RouterGroup, m *AdminManager) {
	r.GET("object_names", m.handleObjectNames)
	r.GET("object/:name", m.handleObjectFields)

	r.GET("/assets/*filepath", func(ctx *gin.Context) {
		p := path.Join("web/dist/", strings.TrimPrefix(ctx.Request.RequestURI, r.BasePath()))
		ctx.FileFromFS(p, http.FS(assets))
	})
	r.GET("/", func(ctx *gin.Context) {
		// handle Vite packaging static resources
		html := strings.ReplaceAll(indexHTML, "/assets/", "assets/")
		html = strings.ReplaceAll(html,
			`window.serverPrefix = '/admin'`, // url inject
			fmt.Sprintf(`window = serverPrefix = '%s'`, r.BasePath()),
		)
		ctx.Data(http.StatusOK, "text/html", []byte(html))
	})
}

func parseWebObjects(m *AdminManager) error {
	for _, obj := range m.AdminBojects {
		m.Names = append(m.Names, obj.WebObject.Name)

		fieldsToJsons := make(map[string]string)
		for k, v := range obj.WebObject.jsonToFields {
			fieldsToJsons[v] = k
		}
		obj.fieldsToJsons = fieldsToJsons

		obj.Filters = make([]string, 0)
		for _, f := range obj.WebObject.Filters {
			obj.Filters = append(obj.Filters, fieldsToJsons[f])
		}

		obj.Orders = make([]string, 0)
		for _, f := range obj.WebObject.Orders {
			obj.Orders = append(obj.Orders, fieldsToJsons[f])
		}

		obj.Edits = make([]string, 0)
		for _, f := range obj.WebObject.Editables {
			obj.Edits = append(obj.Edits, fieldsToJsons[f])
		}

		obj.Searchs = make([]string, 0)
		for _, f := range obj.WebObject.Searchs {
			obj.Searchs = append(obj.Searchs, fieldsToJsons[f])
		}
	}

	return nil
}

// Get object name list
/*
["user", "product"]
*/
func (m *AdminManager) handleObjectNames(c *gin.Context) {
	c.JSON(http.StatusOK, m.Names)
}

// Get object field and types
// Support javascript type: boolean, number, string
/*
{
	"fields": ["id", "name", "age", "enabled", "createdAt"],
	"types": ["number", "string", "number", "boolean", "string"]
	"goTypes" ["uint", "string", "int", "bool", "time.Time"],

	"searchs": ["name"],
	"filters": ["name", "age"],
	"orders": ["createdAt"],
	"edits": ["name"]
}
*/
func (m *AdminManager) handleObjectFields(c *gin.Context) {
	name := c.Param("name")
	if name == "" {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Need to specify name"})
		return
	}

	var res = make(map[string]any)

	fields := make([]string, 0)
	jsTypes := make([]string, 0)
	goTypes := make([]string, 0)

	for _, obj := range m.AdminBojects {
		if obj.WebObject.Name == name {
			rt := obj.WebObject.modelElem
			for i := 0; i < rt.NumField(); i++ {
				f := rt.Field(i)
				jsonName := obj.fieldsToJsons[f.Name]
				typeVal := convertGoTypeToJS(f.Type.Kind())

				fields = append(fields, jsonName)
				jsTypes = append(jsTypes, typeVal)
				goTypes = append(goTypes, f.Type.String())
			}

			res["searchs"] = obj.Searchs
			res["filters"] = obj.Filters
			res["orders"] = obj.Orders
			res["edits"] = obj.Edits
			break
		}
	}

	res["fields"] = fields
	res["goTypes"] = goTypes
	res["types"] = jsTypes

	c.JSON(http.StatusOK, res)
}

// Support javascript type: string, number, boolean, object, any
func convertGoTypeToJS(kind reflect.Kind) string {
	switch kind {
	case reflect.String:
		return "string"
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return "number"
	case reflect.Float32, reflect.Float64:
		return "number"
	case reflect.Bool:
		return "boolean"
	case reflect.Struct, reflect.Map, reflect.Slice:
		return "object"
	case reflect.Ptr, reflect.Interface:
		return "any"
	default:
		return "unknown"
	}
}
