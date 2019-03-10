package configuration

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/shibukawa/configdir"
)

var CfgDir configdir.ConfigDir

func init() {
	CfgDir = configdir.New("thatiq", "kuade")
}

func Resolve() (*Configuration, error) {

}