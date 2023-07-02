package rest

import (
	"fmt"
	"net/http"

	goccy "github.com/goccy/go-json"
	"github.com/labstack/echo/v4"
)

type goccyEchoSerializer struct{}

func NewGoccyEchoSerializer() *goccyEchoSerializer {
	return &goccyEchoSerializer{}
}

func (g *goccyEchoSerializer) Serialize(c echo.Context, i any, indent string) error {
	enc := goccy.NewEncoder(c.Response())
	if indent != "" {
		enc.SetIndent("", indent)
	}
	return enc.Encode(i)
}

func (g *goccyEchoSerializer) Deserialize(c echo.Context, i interface{}) error {

	err := goccy.NewDecoder(c.Request().Body).Decode(i)
	if ute, ok := err.(*goccy.UnmarshalTypeError); ok {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Unmarshal type error: expected=%v, got=%v, field=%v, offset=%v", ute.Type, ute.Value, ute.Field, ute.Offset)).SetInternal(err)
	} else if se, ok := err.(*goccy.SyntaxError); ok {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Syntax error: offset=%v, error=%v", se.Offset, se.Error())).SetInternal(err)
	}
	return err

}
