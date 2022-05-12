package models

import "github.com/myOmikron/echotools/utilitymodels"

type Maintainer struct {
	utilitymodels.CommonID
	UserID      uint
	User        utilitymodels.User
	ContactMail string
	Fingerprint string
}
