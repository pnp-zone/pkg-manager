package task

import (
	"fmt"
	"github.com/ProtonMail/gopenpgp/v2/crypto"
	jsoniter "github.com/json-iterator/go"
	"github.com/myOmikron/echotools/color"
	"github.com/pnp-zone/pkg-manager/conf"
	"github.com/pnp-zone/pkg-manager/models"
	"gorm.io/gorm"
	"io/ioutil"
	"os"
	"time"
)

var json = jsoniter.Config{
	EscapeHTML:    true,
	CaseSensitive: true,
}.Froze()

type IndexResponse struct {
	Packages []Package `json:"packages"`
}

type Package struct {
	Name            string           `json:"name"`
	FlagOrphaned    bool             `json:"flag_orphaned"`
	SourceURL       string           `json:"source_url"`
	Maintainer      []Maintainer     `json:"maintainer"`
	PackageVersions []PackageVersion `json:"package_versions"`
}

type Maintainer struct {
	Name        string `json:"name"`
	ContactMail string `json:"contact_mail"`
	Fingerprint string `json:"fingerprint"`
}

type PackageVersion struct {
	License      string `json:"license"`
	Description  string `json:"description"`
	Bytes        uint   `json:"bytes"`
	VersionMajor uint   `json:"version_major"`
	VersionMinor uint   `json:"version_minor"`
	VersionPatch uint   `json:"version_patch"`
	FlagYanked   bool   `json:"flag_yanked"`
	FlagLatest   bool   `json:"flag_latest"`
}

func BuildIndex(db *gorm.DB, keyring *crypto.KeyRing, config *conf.Config) {
	for {
		t1 := time.Now()
		packages := make([]models.Package, 0)
		packageVersionList := make([]models.PackageVersion, 0)
		packageVersions := make(map[uint][]models.PackageVersion)

		db.Preload("Maintainer.User").Preload("Maintainer").Find(&packages)
		db.Find(&packageVersionList)

		for _, p := range packageVersionList {
			if packageVersions[p.PackageID] == nil {
				packageVersions[p.PackageID] = make([]models.PackageVersion, 0)
			}
			packageVersions[p.PackageID] = append(packageVersions[p.PackageID], p)
		}

		res := IndexResponse{
			Packages: make([]Package, 0),
		}

		for _, p := range packages {
			pack := Package{
				Name:            p.Name,
				FlagOrphaned:    p.FlagOrphaned,
				Maintainer:      make([]Maintainer, 0),
				PackageVersions: make([]PackageVersion, 0),
			}

			if p.SourceURL != nil {
				pack.SourceURL = *p.SourceURL
			}

			for _, mt := range p.Maintainer {
				pack.Maintainer = append(pack.Maintainer, Maintainer{
					Name:        mt.User.Username,
					ContactMail: mt.ContactMail,
					Fingerprint: mt.Fingerprint,
				})
			}

			for _, pv := range packageVersions[p.ID] {
				pack.PackageVersions = append(pack.PackageVersions, PackageVersion{
					License:      pv.License,
					Description:  pv.Description,
					Bytes:        pv.Bytes,
					VersionMajor: pv.VersionMajor,
					VersionMinor: pv.VersionMinor,
					VersionPatch: pv.VersionPatch,
					FlagYanked:   pv.FlagYanked,
					FlagLatest:   pv.FlagLatest,
				})
			}

			res.Packages = append(res.Packages, pack)
		}

		if data, err := json.Marshal(res); err != nil {
			fmt.Println(err.Error())
			os.Exit(1)
		} else {
			if err := ioutil.WriteFile(config.Server.PkgDir+"index.json", data, 0640); err != nil {
				fmt.Println("Couldn't write index:", err.Error())
				os.Exit(1)
			}

			message := crypto.NewPlainMessage(data)
			dataSignature, err := keyring.SignDetached(message)
			if err != nil {
				fmt.Println("Couldn't sign index:", err.Error())
				os.Exit(1)
			}
			armored, err := dataSignature.GetArmored()
			if err != nil {
				fmt.Println("Couldn't get armored signature", err.Error())
				os.Exit(1)
			}

			if err := ioutil.WriteFile(config.Server.PkgDir+"index.json.asc", []byte(armored), 0640); err != nil {
				fmt.Println("Couldn't write index:", err.Error())
				os.Exit(1)
			}
		}

		color.Println(color.PURPLE, "Generated index in "+time.Now().Sub(t1).String())
		time.Sleep(time.Minute)
	}
}
