package frontendapi

import (
	"github.com/pnp-zone/pkg-manager/conf"
	"gorm.io/gorm"
)

type Wrapper struct {
	DB     *gorm.DB
	Config *conf.Config
}
