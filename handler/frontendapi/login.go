package frontendapi

import (
	"github.com/labstack/echo/v4"
	"github.com/myOmikron/echotools/auth"
	"github.com/myOmikron/echotools/middleware"
)

type LoginRequest struct {
	Username string
	Password string
}

func (w *Wrapper) Login(c echo.Context) error {
	var form LoginRequest

	err := echo.FormFieldBinder(c).
		String("username", &form.Username).
		String("password", &form.Password).
		BindError()

	if err != nil {
		return c.String(400, err.Error())
	}

	user, err := auth.Authenticate(w.DB, form.Username, form.Password)
	if err != nil {
		return c.String(401, err.Error())
	}

	if err := middleware.Login(w.DB, user, c); err != nil {
		return c.String(500, err.Error())
	}

	return c.String(200, "Login successful")
}
