package main

import (
	"os"

	"github.com/WolfgangMau/chamgo-qt/config"
	"github.com/alecthomas/kong"
	"github.com/sirupsen/logrus"
	"github.com/therecipe/qt/widgets"
)

//Global Variables - StateStorage
const AppName = "Chameleon"
const AppDescription = "Gui for the chameleon mini"

var Cfg *config.Config
var Statusbar *widgets.QStatusBar
var DeviceActions config.DeviceActions
var MyTabs *widgets.QTabWidget
var TagA QTbytes
var TagB QTbytes
var Params *Cli

type Cli struct {
	Debug  bool   `help:"Enable debug logging."`
	Config string `optional name:"config" help:"Custom configuration file."`
}

func initcfg(configfile string) {
	tmp, err := config.NewConfigReader().Load(configfile)
	if err != nil {
		logrus.Fatal(err)
	}
	Cfg = tmp
	dn := Cfg.Device[SelectedDeviceId].Name
	DeviceActions.Load(Cfg.Device[SelectedDeviceId].CmdSet, dn)

	if _, err = getSerialPorts(); err != nil {
		logrus.Fatal(err)
	}
}

func initParameters() *Cli {
	clirsc := new(Cli)
	kname := kong.Name(AppName)
	kdescription := kong.Description(AppDescription)
	kong.Parse(clirsc, kname, kdescription)

	return clirsc
}

func main() {
	// Parse parameters
	Params = initParameters()
	if Params.Debug == true {
		logrus.SetLevel(logrus.DebugLevel)
	}
	// Parse configuration file
	initcfg(Params.Config)

	Connected = false

	app := widgets.NewQApplication(len(os.Args), os.Args)

	window := widgets.NewQMainWindow(nil, 0)
	window.SetWindowTitle(AppName)
	window.SetFixedSize2(1100, 600)

	mainlayout := widgets.NewQVBoxLayout()

	MyTabs = widgets.NewQTabWidget(nil)
	MyTabs.AddTab(allSlots(), "Tags")
	MyTabs.AddTab(serialTab(), "Device")
	MyTabs.AddTab(dataTab(), "Data")
	MyTabs.SetCurrentIndex(2)

	mainlayout.AddWidget(MyTabs, 0, 0x0020)

	mainWidget := widgets.NewQWidget(nil, 0)
	mainWidget.SetLayout(mainlayout)
	window.SetCentralWidget(mainWidget)

	Statusbar = widgets.NewQStatusBar(window)
	Statusbar.ShowMessage("Disconnected", 0)
	window.SetStatusBar(Statusbar)

	checkForDevices()
	// Show the window
	window.Show()

	// Execute app
	app.Exec()
}
