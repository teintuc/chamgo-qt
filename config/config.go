package config

import (
	"io/ioutil"
	"log"
	"os"
	"os/user"
	"path"
	"path/filepath"

	"gopkg.in/yaml.v2"
)

var configfile = "config.yaml"

type Config struct {
	currentfile string
	Device      []struct {
		Vendor  string            `yaml:"vendor"`
		Product string            `yaml:"product"`
		Name    string            `yaml:"name"`
		Cdc     string            `yaml:"cdc"`
		CmdSet  map[string]string `yaml:"cmdset"`
		Config  struct {
			Slot struct {
				Selected int `yaml:"selected"`
				Offset   int `yaml:"offset"`
				First    int `yaml:"first"`
				Last     int `yaml:"last"`
			} `yaml:"slot"`
			Serial struct {
				Baud              int  `yaml:"baud"`
				WaitForReceive    int  `yaml:"waitforreceive"`
				ConeectionTimeout int  `yaml:"connectiontimeout"`
				Autoconnect       bool `yaml:"autoconnect"`
			} `yaml:"serial"`
			MfkeyBin string `yaml:"mfkeybin"`
		} `yaml:"config"`
	} `yaml:"device"`
}

func (c *Config) Load(cfgfile string) *Config {
	// If the given config file is empty, use the default one
	if len(cfgfile) == 0 {
		cfgfile = c.findConfigFile()
	}
	// Save the current config file for save function
	c.currentfile = cfgfile

	log.Printf("Loading configFile: %s\n", cfgfile)
	yamlFile, err := ioutil.ReadFile(cfgfile)
	if err != nil {
		log.Printf("error reading config (%s) err   #%v ", cfgfile, err)
		os.Exit(2)
	}

	err = yaml.Unmarshal(yamlFile, c)
	if err != nil {
		log.Fatalf("Unmarshal: %v", err)
		return nil
	}
	return c
}

func (c *Config) Save() bool {
	if len(c.currentfile) > 0 {
		if data, err := yaml.Marshal(c); err != nil {
			log.Printf("error Marshall yaml (%s)\n", err)
			return false
		} else {
			err := ioutil.WriteFile(c.currentfile, data, 0644)
			if err != nil {
				log.Fatal(err)
			}
			return true
		}
	}
	return false
}

func (c *Config) findConfigFile() string {
	var foundpath string

	usr, err := user.Current()
	if err != nil {
		log.Fatal(err)
	}

	pathstotest := []string{
		usr.HomeDir,
		".",
	}

	// Loop through possible path. The last as prority
	for _, p := range pathstotest {
		currentpath := path.Join(p, configfile)
		_, err := os.Stat(currentpath)
		if err == nil {
			foundpath = currentpath
		}
	}
	return foundpath
}

func Apppath() string {

	appdir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		log.Printf("no executable-path found!?!\n%s\n", err)
		return ""
	}
	return appdir
}

type DeviceActions struct {
	//config info
	GetModes    string
	GetButtons  string
	GetButtonsl string
	//slot info
	GetMode    string
	GetUid     string
	GetButton  string
	GetButtonl string
	GetSize    string
	//actions
	SelectSlot    string
	SelectedSlot  string
	ClearSlot     string
	StartUpload   string
	StartDownload string

	GetRssi string
}

func (d *DeviceActions) Load(commands map[string]string, device string) {
	switch device {

	case "Chameleon RevE-Rebooted":
		d.GetModes = commands["config"]
		d.GetButtons = commands["button"]

		d.GetMode = commands["config"] + "?"
		d.GetUid = commands["uid"] + "?"
		d.GetButton = commands["button"] + "?"
		d.GetButtonl = commands["buttonl"] + "?"
		d.GetSize = commands["memory"] + "?"

		d.SelectSlot = commands["setting"] + "="
		d.SelectedSlot = commands["setting"] + "?"
		d.StartUpload = commands["upload"]
		d.StartDownload = commands["download"]
		d.ClearSlot = commands["clear"]

		d.GetRssi = commands["rssi"] + "?"

	case "Chameleon RevG":
		d.GetModes = commands["config"] + "=?"
		d.GetButtons = commands["button"] + "=?"

		d.GetMode = commands["config"] + "?"
		d.GetUid = commands["uid"] + "?"
		d.GetButton = commands["button"] + "?"
		d.GetButtonl = commands["buttonl"] + "?"
		d.GetSize = commands["memory"] + "?"

		d.SelectSlot = commands["setting"] + "="
		d.SelectedSlot = commands["setting"] + "?"
		d.StartUpload = commands["upload"]
		d.StartDownload = commands["download"]
		d.ClearSlot = commands["clear"]

		d.GetRssi = commands["rssi"] + "?"
	}
}
