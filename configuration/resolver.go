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
