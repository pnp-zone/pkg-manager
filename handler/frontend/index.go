package frontend

import "github.com/labstack/echo/v4"

func (w *Wrapper) Index(c echo.Context) error {
	return c.Render(200, "index", nil)
}
