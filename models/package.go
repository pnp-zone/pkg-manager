package models

import (
	"github.com/myOmikron/echotools/utilitymodels"
	"time"
)

type Package struct {
	utilitymodels.CommonSoftDelete
	Name         string       `gorm:"not null;unique"`
	Maintainer   []Maintainer `gorm:"many2many:package_maintainer;"`
	FlagOrphaned bool         `gorm:"default:false"`
	SourceURL    *string
}

type PackageVersion struct {
	utilitymodels.CommonID
	CreatedAt    time.Time
	Downloads    uint `gorm:"default:0"`
	PackageID    uint `gorm:"not null"`
	Package      Package
	VersionMajor uint   `gorm:"not null"`
	VersionMinor uint   `gorm:"not null"`
	VersionPatch uint   `gorm:"not null"`
	Bytes        uint   `gorm:"not null"`
	Description  string `gorm:"not null;size:1024"`
	License      string `gorm:"size:32"`
	FlagYanked   bool   `gorm:"default:false"`
}
