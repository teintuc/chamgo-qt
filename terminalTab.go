package main

import (
	"fmt"
	"github.com/WolfgangMau/chamgo-qt/config"
	"github.com/WolfgangMau/chamgo-qt/nonces"
	"github.com/therecipe/qt/widgets"
	"log"
	"os"
	"strconv"
	"strings"
	"time"
)

var (
	serialSendButton *widgets.QPushButton
	serialMonitor    *widgets.QPlainTextEdit
	serialPortSelect *widgets.QComboBox
	deviceSelect     *widgets.QComboBox
)

func serialTab() *widgets.QWidget {
	serialTabLayout := widgets.NewQHBoxLayout()
	leftTabLayout := widgets.NewQVBoxLayout()
	serialTabPage := widgets.NewQWidget(nil, 0)

	/********************************************** Serial Connect *********************************************/

	serialPorts, _ := getSerialPorts()

	serConLayout := widgets.NewQFormLayout(nil)

	deviceSelect = widgets.NewQComboBox(nil)
	deviceSelect.AddItems(getDeviceNames())
	deviceSelect.SetCurrentIndex(SelectedDeviceId)
	deviceSelect.SetFixedWidth(160)

	serialPortSelect = widgets.NewQComboBox(nil)
	serialPortSelect.AddItems(serialPorts)
	serialPortSelect.SetFixedWidth(160)
	serialPortSelect.SetCurrentIndex(SelectedPortId)

	serialConnectButton := widgets.NewQPushButton2("Connect", nil)

	serialDeviceInfo := widgets.NewQLabel(nil, 0)
	serialDeviceInfo.SetText("not Connected")

	serConLayout.AddWidget(deviceSelect)
	serConLayout.AddWidget(serialPortSelect)
	serConLayout.AddWidget(serialConnectButton)
	serConLayout.AddWidget(serialDeviceInfo)

	serialConnectGroup := widgets.NewQGroupBox2("Serial Connection", nil)
	serialConnectGroup.SetLayout(serConLayout)
	serialConnectGroup.SetFixedSize2(220, 180)
	leftTabLayout.AddWidget(serialConnectGroup, 1, 0x0020)

	macrodir := config.Apppath() + string(os.PathSeparator) + "macros" + string(os.PathSeparator)
	log.Println("checking for macrodir: ", macrodir)
	var macros []string
	if macrodir != "" {
		macros = config.GetFilesInFolder(macrodir, ".cmds")
	}
	if len(macros) > 0 {
		log.Println("Macro-Files found: ", len(macros))

		macroGroupLayout := widgets.NewQHBoxLayout()
		macroGroup := widgets.NewQGroupBox2("Command Macros", nil)
		macroGroup.SetFixedWidth(220)
		macroSelect := widgets.NewQComboBox(macroGroup)
		macroSelect.AddItems(macros)
		macroGroupLayout.AddWidget(macroSelect, 1, 0x0020)
		macroSend := widgets.NewQPushButton2("execute", nil)
		macroGroupLayout.AddWidget(macroSend, 1, 0x0020)

		macroGroup.SetLayout(macroGroupLayout)
		leftTabLayout.AddWidget(macroGroup, 1, 0x0020)

		macroSend.ConnectClicked(func(checked bool) {

			if Connected {

				log.Println("execute macro ", macroSelect.CurrentText())
				cmds := config.ReadFileLines(config.Apppath() + string(os.PathSeparator) + "macros" + string(os.PathSeparator) + macroSelect.CurrentText())
				if len(cmds) > 0 {
					for _, c := range cmds {
						if strings.Contains(strings.ToLower(c), "detectionmy?") {
							serialMonitor.AppendPlainText("-> " + strings.Replace(strings.Replace(c, "\r", "", -1), "\n", "", -1))

							// send cmd ang get the expected 218 bytes (208 nonce + 2 crc + 8 cmd-response (100:OK\n\r)
							SerialSendOnly(c)
							buff := GetSpecificBytes(218)
							//buffer should be empty - only to get sure
							SerialPort.ResetInputBuffer()

							responsecode := strings.Replace(strings.Replace(string(buff[len(buff)-8:]), "\r", "", -1), "\n", "", -1)
							log.Println("len enc: ", len(buff))
							buff = nonces.DecryptData(buff[0:len(buff)-10], 123321, 208)
							uid := buff[0:4]
							serialMonitor.AppendPlainText(fmt.Sprintf("uid: %04X\n", uid))

							noncemap := nonces.ExtractNonces(buff)

							if len(noncemap) > 0 {
								serialMonitor.AppendPlainText(fmt.Sprintf("found %d nonces\n\t#     NT     NR     AR", len(noncemap)))
								for i, n := range noncemap {
									serialMonitor.AppendPlainText(fmt.Sprintf("nonce #%d: %X %X %X", i+1, n.Nt, n.Nr, n.Ar))
								}
							}
							serialMonitor.AppendPlainText(fmt.Sprintf("<- %s\nuid: %x\nbuff (%d): %X", responsecode, uid, len(buff), buff))
						} else {
							sendSerialCmd(c)
							time.Sleep(time.Millisecond * time.Duration(Cfg.Device[SelectedDeviceId].Config.Serial.WaitForReceive))
							serialMonitor.AppendPlainText(fmt.Sprintf("-> cmd: %s response: %s", c, SerialResponse.Payload))
						}
						serialMonitor.Repaint()
					}
				}
			}
		})
	}

	serialTabLayout.AddLayout(leftTabLayout, 0)

	serialConnectButton.ConnectClicked(func(checked bool) {
		Commands := Cfg.Device[SelectedDeviceId].CmdSet
		//for n,c := range Commands {
		//	log.Printf("Command Name: %s -> %s\n",n,c)
		//}
		//Commands.load(deviceSelect.CurrentText())

		if serialConnectButton.Text() == "Connect" {

			err := connectSerial(SerialDevice)
			if err != nil {
				widgets.QMessageBox_Information(nil, "OK", "can't connect to Serial\n"+string(err.Error()),
					widgets.QMessageBox__Ok, widgets.QMessageBox__Ok)
				log.Println("error on connect: ", err)
			} else {
				initcfg()
				if len(DeviceActions.GetUid) <= 0 {
					log.Println("no action for 'getUid!?' ", DeviceActions.GetUid)
				}

				//ask for the device-version
				sendSerialCmd(Commands["version"] + "?")

				if SerialResponse.Code >= 100 {
					serialConnectButton.SetText("Disconnect")
					serialSendButton.SetDisabled(false)
					serialSendButton.Repaint()
					//checkForDevices()
				}
				//web got a expected answer from the VERSION(MY) cmd
				if SerialResponse.Code == 101 {
					serialDeviceInfo.SetText("Connected\n" + deviceInfo(SerialResponse.Payload))
					Connected = true
					Statusbar.ShowMessage("Connected to Port: "+serialPortSelect.CurrentText()+" - Device: "+Cfg.Device[SelectedDeviceId].Cdc+" - Firmware: "+deviceInfo(SerialResponse.Payload), 0)
					if SerialResponse.Code == 101 {
						buttonClicked(0)
						buttonClicked(4)
						buttonClicked(1)
						MyTabs.SetCurrentIndex(0)
					}
				} else {
					widgets.QMessageBox_Information(nil, "OK", "no Version Response from Device!",
						widgets.QMessageBox__Ok, widgets.QMessageBox__Ok)
				}
			}
		} else {
			err := SerialPort.Close()
			if err == nil {
				Cfg.Save()
				serialConnectButton.SetText("Connect")
				serialSendButton.SetDisabled(true)
				serialSendButton.Repaint()
				serialDeviceInfo.SetText("not Connected")
				Statusbar.ShowMessage("not Connected", 0)
				Connected = false
				SerialPort.Close()

			}
		}

	})

	/********************************************** Serial Monitor *********************************************/

	serMonitorLayout := widgets.NewQVBoxLayout()
	serSendLayout := widgets.NewQHBoxLayout()

	serialMonitor = widgets.NewQPlainTextEdit(nil)
	serialMonitor.AppendPlainText("")
	serialMonitor.SetFixedHeight(380)
	serialMonitor.SetReadOnly(true)

	serialSendButton = widgets.NewQPushButton2("send", nil)
	serialSendButton.SetDisabled(true)
	serialSendButton.SetFixedWidth(80)

	serialSendTxt := widgets.NewQLineEdit(nil)
	serialSendTxt.ConnectReturnPressed(serialSendButton.Click)
	serialSendTxt.SetTabOrder(serialSendButton, nil)

	serSendLayout.AddWidget(serialSendTxt, 0, 0)
	serSendLayout.AddWidget(serialSendButton, 1, 0)

	serMonitorLayout.AddWidget(serialMonitor, 0, 0)
	serMonitorLayout.AddLayout(serSendLayout, 0)

	serialSendGroup := widgets.NewQGroupBox2("Serial Terminal", nil)
	serialSendGroup.SetLayout(serMonitorLayout)

	serialTabLayout.AddWidget(serialSendGroup, 0, 0x0020)
	serialTabLayout.SetAlign(33)

	serialSendButton.ConnectClicked(func(checked bool) {
		if serialSendTxt.Text() != "" {
			sendSerialCmd(serialSendTxt.Text())
			if SerialResponse.Code >= 100 {
				serialMonitor.AppendPlainText("-> " + SerialResponse.Cmd)
				serialMonitor.AppendPlainText("<- " + strconv.Itoa(SerialResponse.Code) + " " + SerialResponse.String)
				if SerialResponse.Payload != "" {
					serialMonitor.AppendPlainText("<- " + SerialResponse.Payload)
				}
				serialMonitor.Repaint()
			} else {
				widgets.QMessageBox_Information(nil, "OK", "no Response for Cmd:\n"+serialSendTxt.Text(),
					widgets.QMessageBox__Ok, widgets.QMessageBox__Ok)
			}
		}
	})

	serialTabPage.SetLayout(serialTabLayout)

	return serialTabPage
}