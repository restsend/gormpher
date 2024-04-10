package gormpher

import (
	"embed"
	"fmt"
	"log"
	"net/http"
	"os"
	"path"
	"reflect"
	"strings"

	"github.com/gin-gonic/gin"
)

// old admin

//go:embed web/dist/index.html web/dist/assets/* html/resources/*
var assets embed.FS

//go:embed web/dist/index.html
var indexHTML string

// new admin

//go:embed html/index.html
var newHTML string

type AdminObject struct {
	webObject *WebObject

	// All elements are json format string
	Searchs []string
	Filters []string
	Edits   []string
	Orders  []string

	// Map fields to jsonTag. such as:
	// UUID string `json:"id"` => {"UUID": "id"}
	fieldToJSONs map[string]string

	PrimaryKeyField    string
	PrimaryKeyJsonName string
}

type AdminManager struct {
	AdminObjects []AdminObject
	Names        []string
}

// RegisterObjectsWithAdmin quickly Register Admin by webobjects
func RegisterObjectsWithAdmin(r *gin.RouterGroup, objs []WebObject) {
	m := AdminManager{}
	for _, obj := range objs {
		m.RegisterObject(r, obj)
	}
	RegisterAdminHandler(r, &m)
}

func (m *AdminManager) RegisterObject(r *gin.RouterGroup, obj WebObject) {
	if err := obj.RegisterObject((gin.IRouter)(r)); err != nil {
		log.Fatalf("RegisterObjectWithAdmin [%s] fail %v\n", obj.Name, err)
	}

	m.Names = append(m.Names, obj.Name)
	m.AdminObjects = append(m.AdminObjects, woToAo(obj))
}

// convert WebObject to AdminObject
func woToAo(wo WebObject) AdminObject {
	ao := AdminObject{webObject: &wo}

	rt := reflect.TypeOf(wo.Model)
	if rt.Kind() == reflect.Ptr {
		rt = rt.Elem()
	}

	fieldsToJsons := make(map[string]string)
	for i := 0; i < rt.NumField(); i++ {
		f := rt.Field(i)
		jsonTag := f.Tag.Get("json")
		if jsonTag == "-" {
			continue
		} else if jsonTag == "" {
			fieldsToJsons[f.Name] = f.Name
		} else {
			var jsonField string = jsonTag
			if strings.Contains(jsonTag, ",") {
				jsonField = strings.Split(jsonTag, ",")[0]
			}
			fieldsToJsons[f.Name] = jsonField
		}
	}
	ao.fieldToJSONs = fieldsToJsons

	ao.Filters = make([]string, 0)
	for _, f := range wo.FilterFields {
		ao.Filters = append(ao.Filters, fieldsToJsons[f])
	}

	ao.Orders = make([]string, 0)
	for _, f := range wo.OrderFields {
		ao.Orders = append(ao.Orders, fieldsToJsons[f])
	}

	ao.Edits = make([]string, 0)
	for _, f := range wo.EditFields {
		ao.Edits = append(ao.Edits, fieldsToJsons[f])
	}

	ao.Searchs = make([]string, 0)
	for _, f := range wo.SearchFields {
		ao.Searchs = append(ao.Searchs, fieldsToJsons[f])
	}

	return ao
}

// RegisterAdmin resolve adminManager & register admin handler
func RegisterAdminHandler(r *gin.RouterGroup, m *AdminManager) {
	r.GET("object_names", m.handleObjectNames)
	r.GET("object/:name", m.handleObjectFields)

	// old admin assets
	r.GET("/assets/*filepath", func(ctx *gin.Context) {
		p := path.Join("web/dist/", strings.TrimPrefix(ctx.Request.RequestURI, r.BasePath()))
		ctx.FileFromFS(p, http.FS(assets))
	})

	// new admin assets
	r.GET("/resources/*filepath", func(ctx *gin.Context) {
		p := path.Join("html/", strings.TrimPrefix(ctx.Request.RequestURI, r.BasePath()))
		ctx.FileFromFS(p, http.FS(assets))
	})

	r.GET("/v1", func(ctx *gin.Context) {
		// handle Vite packaging static resources
		html := strings.ReplaceAll(indexHTML, "/assets/", "assets/")
		html = strings.ReplaceAll(html,
			`window.serverPrefix = '/admin'`, // url inject
			fmt.Sprintf(`window = serverPrefix = '%s'`, r.BasePath()),
		)
		ctx.Data(http.StatusOK, "text/html", []byte(html))
	})

	r.GET("/v2", func(ctx *gin.Context) {
		// for development
		b, err := os.ReadFile("../html/index.html")
		if err != nil {
			panic(err)
		}
		ctx.Data(http.StatusOK, "text/html", b)

		// ctx.Data(http.StatusOK, "text/html", []byte(newHTML))
	})
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
	"edits": ["name"],

	"primaryKey": "id",
}
*/
func (m *AdminManager) handleObjectFields(c *gin.Context) {
	name := c.Param("name")
	if name == "" {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Need to specify name"})
		return
	}

	var result = make(map[string]any)

	fields := make([]string, 0)
	jsTypes := make([]string, 0)
	goTypes := make([]string, 0)

	for _, obj := range m.AdminObjects {
		if obj.webObject.Name == name {
			rt := obj.webObject.modelElem
			for i := 0; i < rt.NumField(); i++ {
				f := rt.Field(i)

				jsonTag := f.Tag.Get("json")
				if jsonTag == "-" {
					continue
				}

				jsonName := obj.fieldToJSONs[f.Name]
				typeVal := convertGoTypeToJS(f.Type.Kind())

				fields = append(fields, jsonName)
				jsTypes = append(jsTypes, typeVal)
				goTypes = append(goTypes, f.Type.String())
			}

			result["searchs"] = obj.Searchs
			result["filters"] = obj.Filters
			result["orders"] = obj.Orders
			result["edits"] = obj.Edits
			result["primaryKey"] = obj.webObject.jsonPKName
			break
		}
	}

	result["fields"] = fields
	result["goTypes"] = goTypes
	result["types"] = jsTypes

	c.JSON(http.StatusOK, result)
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
