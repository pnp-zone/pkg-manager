package frontend

import "github.com/labstack/echo/v4"

func (w *Wrapper) Register(c echo.Context) error {
	return c.Render(200, "register", nil)
}
