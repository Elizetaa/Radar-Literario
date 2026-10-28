package controller

import (
	"encoding/json"
	"errors"
	"github.com/gin-gonic/gin"
	"io"
	"log/slog"
	"radar-literario/backend/internal/model/core"
	"radar-literario/backend/internal/model/schemas"
	"strconv"
)

func Actor(c *gin.Context) core.Actor {
	a, _ := c.Get("actor")
	if v, ok := a.(core.Actor); ok {
		return v
	}
	return core.Actor{}
}
func Bind(c *gin.Context, dst any) error {
	d := json.NewDecoder(c.Request.Body)
	d.DisallowUnknownFields()
	if e := d.Decode(dst); e != nil {
		return core.Invalid("INVALID_JSON", "Informe JSON válido com os campos documentados.")
	}
	if e := d.Decode(new(any)); e != io.EOF {
		return core.Invalid("INVALID_JSON", "Envie apenas um objeto JSON.")
	}
	return nil
}
func ID(c *gin.Context, key string) (uint64, error) {
	id, e := strconv.ParseUint(c.Param(key), 10, 64)
	if e != nil || id == 0 {
		return 0, core.Invalid("INVALID_ID", "Identificador inválido.")
	}
	return id, nil
}
func Page(c *gin.Context) (schemas.Page, error) {
	p := schemas.Page{Page: 1, PageSize: 20}
	var e error
	if v, ok := c.GetQuery("page"); ok {
		p.Page, e = strconv.Atoi(v)
		if e != nil {
			return p, core.Invalid("INVALID_PAGE", "Página inválida.")
		}
	}
	if v, ok := c.GetQuery("page_size"); ok {
		p.PageSize, e = strconv.Atoi(v)
		if e != nil {
			return p, core.Invalid("INVALID_PAGE", "Tamanho de página inválido.")
		}
	}
	if p.Page < 1 || p.Page > 1000000 || p.PageSize < 1 || p.PageSize > 100 {
		return p, core.Invalid("INVALID_PAGE", "Use page >= 1 e page_size de 1 a 100.")
	}
	return p, nil
}
func Reply(c *gin.Context, status int, data any) {
	if status == 204 {
		c.Status(status)
		return
	}
	switch v := data.(type) {
	case schemas.List:
		c.JSON(status, v)
	case *schemas.List:
		c.JSON(status, v)
	default:
		c.JSON(status, gin.H{"data": data})
	}
}
func Fail(c *gin.Context, err error) {
	var e *core.Error
	if !errors.As(err, &e) {
		slog.Error("falha interna", "request_id", c.GetString("request_id"), "error_type", strconv.QuoteToASCII(typeName(err)))
		e = core.E(500, "INTERNAL_ERROR", "Não foi possível concluir a operação.")
	}
	c.AbortWithStatusJSON(e.Status, gin.H{"error": e})
}
func typeName(e error) string {
	if e == nil {
		return "nil"
	}
	return "repository_or_dependency"
}
