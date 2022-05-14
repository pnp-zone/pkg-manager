package frontendapi

import (
	"github.com/ProtonMail/gopenpgp/v2/crypto"
	"github.com/pnp-zone/pkg-manager/conf"
	"gorm.io/gorm"
)

type Wrapper struct {
	DB      *gorm.DB
	Keyring *crypto.KeyRing
	Config  *conf.Config
}
