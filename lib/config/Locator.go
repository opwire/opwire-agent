package config

import (
	"path/filepath"
	"os"
	"os/user"
	"github.com/opwire/opwire-agent/lib/storages"
	"github.com/opwire/opwire-agent/lib/utils"
)

const BLANK = ""
const configDir = "opwire"
const configFileName = "opwire-agent"

var configFileExts []string = []string{ ".cfg", ".conf", ".json" }

type Locator struct{}

func NewLocator() (*Locator) {
	return &Locator{}
}

func (l *Locator) GetConfigPath(argConfigPath string) (string, string) {
	fs := storages.GetFs()
	for _, pos := range getConfigSeries() {
		if pos == "arg" && len(argConfigPath) > 0 {
			_, err := fs.Stat(argConfigPath)
			if err == nil {
				return argConfigPath, pos
			}
		}
		var cfgpath string
		var cfgdir string
		switch (pos) {
		case "env":
			cfgdir = findEnvConfigDir()
		case "bin":
			cfgdir = findProgramDir()
		case "cwd":
			cfgdir = findWorkingDir()
		case "xdg":
			cfgdir = findXdgConfigDir()
		case "home":
			cfgdir = findUserHomeDir()
		case "etc":
			cfgdir = "/etc"
		}
		if cfgdir != BLANK {
			for _, ext := range configFileExts {
				cfgfile := configFileName + ext
				if pos == "home" {
					cfgfile = "." + cfgfile
				}
				cfgpath = filepath.Join(cfgdir, cfgfile)
				_, err := fs.Stat(cfgpath)
				if err == nil {
					return cfgpath, pos
				}
			}
		}
	}
	return BLANK, BLANK
}

func getConfigSeries() []string {
	series := utils.Split(os.Getenv("OPWIRE_AGENT_CONFIG_SERIES"), ",")
	if len(series) == 0 {
		series = []string { "arg", "env", "bin", "cwd", "xdg", "home", "etc" }
	}
	return series
}

func findEnvConfigDir() string {
	return os.Getenv("OPWIRE_AGENT_CONFIG_DIR")
}

func findProgramDir() string {
	var dir string
	ex, err := os.Executable()
	if err == nil {
		dir = filepath.Dir(ex)
	}
	return dir
}

func findWorkingDir() string {
	dir, err := os.Getwd()
	if err != nil {
		return BLANK
	}
	return dir
}

func findUserHomeDir() string {
	var dir string
	if usr, err := user.Current(); err == nil {
		dir = usr.HomeDir
	} else {
		// Fallback to reading $HOME
		dir = os.Getenv("HOME")
	}
	return dir
}

func findXdgConfigDir() string {
	var dir string
	// See https://specifications.freedesktop.org/basedir-spec/latest/),
	xdgdir := os.Getenv("XDG_CONFIG_HOME")
	if xdgdir != BLANK {
		// $XDG_CONFIG_HOME/opwire
		dir = filepath.Join(xdgdir, configDir)
	} else {
		homedir := findUserHomeDir()
		if homedir != BLANK {
			// $HOME/.config/opwire
			dir = filepath.Join(homedir, ".config", configDir)
		}
	}
	return dir
}
