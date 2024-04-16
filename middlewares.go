package gormpher

import (
	"fmt"
	"log"
	"runtime"
	"strings"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

var DbField = "gormpher_db_field"

func WithGormDB(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set(DbField, db)
		c.Next()
	}
}

type ErrorWithCode interface {
	StatusCode() int
}

func AbortWithJSONError(c *gin.Context, code int, err error) {
	var errWithFileNum error = err

	if log.Flags()&(log.Lshortfile|log.Llongfile) != 0 {
		var ok bool
		_, file, line, ok := runtime.Caller(1)
		if !ok {
			file = "???"
			line = 0
		}
		pos := strings.LastIndex(file, "/")
		if log.Flags()&log.Lshortfile != 0 && pos >= 0 {
			file = file[1+pos:]
		}
		errWithFileNum = fmt.Errorf("%s:%d: %v", file, line, err)
	}
	c.Error(errWithFileNum)

	if e, ok := err.(ErrorWithCode); ok {
		code = e.StatusCode()
	}

	if c.IsAborted() {
		c.JSON(code, gin.H{"error": err.Error()})
	} else {
		c.AbortWithStatusJSON(code, gin.H{"error": err.Error()})
	}
}
