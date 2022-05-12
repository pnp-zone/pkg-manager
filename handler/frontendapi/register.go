package frontendapi

import (
	"github.com/ProtonMail/gopenpgp/v2/crypto"
	"github.com/labstack/echo/v4"
	"github.com/myOmikron/echotools/database"
	"github.com/myOmikron/echotools/utilitymodels"
	"github.com/pnp-zone/pkg-manager/models"
	"io/ioutil"
)

type RegisterRequest struct {
	Username    string
	Password    string
	Password2   string
	Mail        string
	ContactMail string
	PGPKey      string
}

func (w *Wrapper) Register(c echo.Context) error {
	var form RegisterRequest

	err := echo.FormFieldBinder(c).
		String("username", &form.Username).
		String("password", &form.Password).
		String("password2", &form.Password2).
		String("mail", &form.Mail).
		String("contact_mail", &form.ContactMail).
		String("pgp", &form.PGPKey).
		BindError()

	if err != nil {
		return c.String(400, err.Error())
	}

	if form.Password != form.Password2 {
		return c.String(400, "Passwords must be identical")
	}

	armored, err := crypto.NewKeyFromArmored(form.PGPKey)
	if err != nil {
		return c.String(400, "Invalid PGP key")
	}

	// Check if key is valid
	if armored.IsExpired() {
		return c.String(400, "Key is expired")
	}

	if armored.IsRevoked() {
		return c.String(400, "Key is revoked")
	}

	fpr := armored.GetFingerprint()

	var count int64
	w.DB.Find(&utilitymodels.User{}, "email = ? OR username = ?", form.Mail, form.Username).Count(&count)
	if count != 0 {
		return c.String(409, "username or email is already in use")
	}

	w.DB.Find(&models.Maintainer{}, "contact_mail = ?", form.ContactMail).Count(&count)
	if count != 0 {
		return c.String(409, "username or email is already in use")
	}

	w.DB.Find(&models.Maintainer{}, "fingerprint = ?", fpr).Count(&count)
	if count != 0 {
		return c.String(409, "fingerprint is already in use")
	}

	err = ioutil.WriteFile(w.Config.Server.PGPDir+fpr+".asc", []byte(form.PGPKey), 0600)
	if err != nil {
		return c.String(500, "Error processing PGP key")
	}

	user, err := database.CreateUser(w.DB, form.Username, form.Password, &form.Mail, true)
	if err != nil || user == nil {
		return c.String(500, "Database error")
	}

	w.DB.Create(&models.Maintainer{
		UserID:      user.ID,
		ContactMail: form.ContactMail,
		Fingerprint: fpr,
	})

	return c.String(200, "yey")
}
