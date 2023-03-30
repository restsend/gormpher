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

//go:embed web/dist/index.html web/dist/assets/*
var assets embed.FS

//go:embed web/dist/index.html
var indexHTML string

type AdminObject struct {
	// If you set webObject, it will be resolved.
	// Or you can manually construct the following fields.
	webObject *WebObject

	// All elements are json format string
	Searchs []string
	Filters []string
	Edits   []string
	Orders  []string

	// Map fields to jsonTag. such as:
	// UUID string `json:"id"` => {"UUID": "id"}
	fieldsToJsons map[string]string
}

type AdminManager struct {
	AdminBojects []*AdminObject
	Names        []string
}

// RegisterObjectsWithAdmin quickly Regiter Admin by webobjects
func RegisterObjectsWithAdmin(r *gin.RouterGroup, objs []WebObject) {
	m := AdminManager{}
	for _, obj := range objs {
		m.RegisterObject(r, &obj)
	}
	RegisterAdmin(r, &m)
}

// RegiterAdmin resolve adminManager & register admin handler
func RegisterAdmin(r *gin.RouterGroup, m *AdminManager) {
	m.parseWebObjects()

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

func (m *AdminManager) RegisterObject(r *gin.RouterGroup, obj *WebObject) {
	if err := obj.RegisterObject((gin.IRouter)(r)); err != nil {
		log.Fatalf("RegisterObjectWithAdmin [%s] fail %v\n", obj.Name, err)
	}
	m.AdminBojects = append(m.AdminBojects, &AdminObject{
		webObject: obj,
	})
}

func (m *AdminManager) parseWebObjects() {
	for _, obj := range m.AdminBojects {

		if obj.webObject == nil {
			continue
		}

		m.Names = append(m.Names, obj.webObject.Name)

		fieldsToJsons := make(map[string]string)
		for k, v := range obj.webObject.jsonToFields {
			fieldsToJsons[v] = k
		}
		obj.fieldsToJsons = fieldsToJsons

		obj.Filters = make([]string, 0)
		for _, f := range obj.webObject.Filters {
			obj.Filters = append(obj.Filters, fieldsToJsons[f])
		}

		obj.Orders = make([]string, 0)
		for _, f := range obj.webObject.Orders {
			obj.Orders = append(obj.Orders, fieldsToJsons[f])
		}

		obj.Edits = make([]string, 0)
		for _, f := range obj.webObject.Editables {
			obj.Edits = append(obj.Edits, fieldsToJsons[f])
		}

		obj.Searchs = make([]string, 0)
		for _, f := range obj.webObject.Searchs {
			obj.Searchs = append(obj.Searchs, fieldsToJsons[f])
		}
	}
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
		if obj.webObject.Name == name {
			rt := obj.webObject.modelElem
			for i := 0; i < rt.NumField(); i++ {
				f := rt.Field(i)

				if f.Tag.Get("json") == "-" {
					continue
				}

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
