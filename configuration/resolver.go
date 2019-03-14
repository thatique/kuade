package configuration

import (
	"errors"
	"fmt"
	"path/filepath"

	"github.com/shibukawa/configdir"
)

var (
	ConfigDoesnotExist = errors.New("config doesn't exists.")

	CfgDir configdir.ConfigDir
)

const (
	// Directory contains below files/directories for HTTPS configuration.
	certsDir = "certs"

	// Directory contains all CA certificates other than system defaults for HTTPS.
	certsCADir = "CAs"

	// Public certificate file for HTTPS.
	publicCertFile = "public.crt"

	// Private key file for HTTPS.
	privateKeyFile = "private.key"
)

func init() {
	CfgDir = configdir.New("thatiq", "kuade")
	CfgDir.LocalPath, _ = filepath.Abs(".")
}

func Get() (*Configuration, error) {
	folder := CfgDir.QueryFolderContainsFile("kuade-config.yml")
	if folder == nil {
		return nil, ConfigDoesnotExist
	}

	fp, err := folder.Open("kuade-config.yml")
	if err != nil {
		return nil, err
	}

	defer fp.Close()

	config, err := Parse(fp)
	if err != nil {
		return nil, fmt.Errorf("error parsing %s: %v", filepath.Join(folder.Path, "kuade-config.yml"), err)
	}

	return config, nil
}

func GetDefaultCertsDir() string {
	folders := CfgDir.QueryFolders(configdir.System)
	return filepath.Join(folders[0].Path, certsDir)
}

func GetDefaultCertsCADir() string {
	return filepath.Join(GetDefaultCertsDir(), certsCADir)
}

func GetPublicCertFile() string {
	return filepath.Join(GetDefaultCertsDir(), publicCertFile)
}

func GetPrivateKeyFile() string {
	return filepath.Join(GetDefaultCertsDir(), privateKeyFile)
}
