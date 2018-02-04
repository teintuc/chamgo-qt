package main

import (
	"bytes"
	"github.com/therecipe/qt/widgets"
	"io/ioutil"
	"log"
	"os"
	"strconv"
	"strings"
	"runtime"
)

var temp2 []string

/*var temp string
var myTime time.Time
var GetSlotTicker *time.Ticker*/

func buttonClicked(btn int) {

	switch ActionButtons[btn] {

	case "Select All":
		selectAllSlots(true)
		if populated {
			//GetSlotTicker.Stop()
		}

	case "Select None":
		selectAllSlots(false)
		if populated {
			//GetSlotTicker.Stop()
		}

	case "Apply":
		applySlot()

	case "Clear":
		clearSlot()

	case "Refresh":
		refreshSlot()

	case "Set Active":
		activateSlots()

	case "mfkey32":
		mfkey32Slots()

	case "Upload":
		uploadSlots()

	case "Download":
		downloadSlots()

	default:
		log.Printf("clicked on Button: %s\n", ActionButtons[btn])
	}
}

func slotChecked(slot, state int) {
	//log.Printf(" Checked %d - state: %d\n", slot, state)
	if state == 2 && Connected {
		if Device == Devices.name[1] {
			//RevG's first Slot is 1 and Last Slot is 8
			sendSerialCmd(DeviceActions.selectSlot + strconv.Itoa(slot+1))
		} else {
			//RevE's first Slot is 0 and Last Slot is 7
			sendSerialCmd(DeviceActions.selectSlot + strconv.Itoa(slot))
		}
	}
	Slots[slot].slot.Repaint()
}

func selectAllSlots(b bool) {
	for _, s := range Slots {
		s.slot.SetChecked(b)
		s.slot.Repaint()
	}
}

func applySlot() {
	//GetSlotTicker.Stop()
	for i, s := range Slots {
		if s.slot.IsChecked() {
			log.Printf("********************\nupdating %s\n", s.slotl.Text())
			hardwareSlot := i
			if Device == Devices.name[1] {
				hardwareSlot = i + 1
			}
			sendSerialCmd(DeviceActions.selectSlot + strconv.Itoa(hardwareSlot))
			//select slot
			sendSerialCmd(Commands.config + "=" + s.mode.CurrentText())
			//set mode
			sendSerialCmd(Commands.config + "=" + s.mode.CurrentText())
			//set uid
			sendSerialCmd(Commands.uid + "=" + s.uid.Text())
			//set  button short
			sendSerialCmd(Commands.button + "=" + s.btns.CurrentText())
			//set button long
			sendSerialCmd(Commands.lbutton + "=" + s.btnl.CurrentText())
		}
	}
	populateSlots()
}

func countSelected() int {
	c := 0
	for _, s := range Slots {
		if s.slot.IsChecked() {
			c++
		}
	}
	return c
}

func clearSlot() {
	c1 := 0
	for i, s := range Slots {
		sel := s.slot.IsChecked()
		if sel {
			c1++
			log.Printf("clearing %s\n", s.slotl.Text())
			hardwareSlot := i
			if Device == Devices.name[1] {
				hardwareSlot = i + 1
			}
			sendSerialCmd(DeviceActions.selectSlot + strconv.Itoa(hardwareSlot))
			sendSerialCmd(DeviceActions.clearSlot)
		}
	}
	populateSlots()
}

func refreshSlot() {
	populateSlots()
}

func activateSlots() {
	if countSelected() > 1 {
		widgets.QMessageBox_Information(nil, "OK", "please select only one Slot",
			widgets.QMessageBox__Ok, widgets.QMessageBox__Ok)
		return
	}
	for i, s := range Slots {
		sel := s.slot.IsChecked()
		if sel {
			hardwareSlot := i
			if Device == Devices.name[1] {
				hardwareSlot = i + 1
			}
			sendSerialCmd(DeviceActions.selectSlot + strconv.Itoa(hardwareSlot))
		}
	}
}

func mfkey32Slots() {
	for i, s := range Slots {
		sel := s.slot.IsChecked()
		if sel {
			log.Printf("I should probably calc keys for Slot %d\n", i)
		}
	}
}

//type packet struct {
//	proto   byte
//	block   int
//	rblocks int
//	data    []byte
//	chk     byte
//}

func uploadSlots() bool {
	if countSelected() > 1 {
		widgets.QMessageBox_Information(nil, "OK", "please select only one Slot",
			widgets.QMessageBox__Ok, widgets.QMessageBox__Ok)
		return false
	}
	var filename string
	fileSelect := widgets.NewQFileDialog(nil, 0)
	filename = fileSelect.GetOpenFileName(nil, "open Dump", "", "", "", fileSelect.Options())
	if filename == "" {
		log.Println("no file selected")
		return false
	}

	for i, s := range Slots {
		if s.slot.IsChecked() {
			//log.Printf("********************\nupdating %s\n", s.slotl.Text())
			hardwareSlot := i
			if Device == Devices.name[1] {
				hardwareSlot = i + 1
			}
			sendSerialCmd(DeviceActions.selectSlot + strconv.Itoa(hardwareSlot))
			//log.Printf("upoload %s to Slot %d\n", filename, i)
			// Open file
			log.Printf("loading file %s\n", filename)
			fIn, err := os.Open(filename)
			if err != nil {
				log.Fatalln(err)
			}
			//readfile into buffer
			data, err := ioutil.ReadAll(fIn)
			if err != nil {
				log.Println(err)
			}
			fIn.Close()

			var p []xblock
			var p1 xblock
			oBuffer := make([]byte, 1)
			for _, d := range data {
				p1.payload = append(p1.payload, d)

				if len(p1.payload) == 128 {
					p1.proto = []byte{SOH}
					p1.packetNum = len(p)
					p1.packetInv = 255 - p1.packetNum
					p1.checksumm = int(checksum(p1.payload, 0))
					p = append(p, p1)
					p1.payload = []byte("")
				}
			}

			//set chameleon into receiver-mode
			sendSerialCmd(DeviceActions.startUpload)
			if SerialResponse.Code == 110 {
				//start uploading packets
				failure := 0
				success := 0
				//log.Printf("start sending %d Packets of %d bytes payload\n", len(p), len(p[0].payload))
				for _, sp := range p {
					var reSend = true
					for reSend {
						//log.Printf("send Packet: %d\n", sp.packetNum)
						sendPacket(sp)
						if _, err = serialPort.Read(oBuffer); err != nil {
							log.Println(err)
						} else {
							switch oBuffer[0] {
							case NAK: // NAK
								//receiver ask for retransmission of this block
								log.Printf("resend Packet %d\n", sp.packetNum)
								reSend = true
								failure++
							case ACK: // ACK
								//receiver accepted this block
								reSend = false
								success++
							case CAN: // CAN
								//receiver wants to quit session
								log.Printf("receiver aborted transmission at Packet %d\n", sp.packetNum)
								reSend = false
								failure++
							default:
								//should not happen
								log.Printf("unexspected answer(0x%X) for packet %d\n", oBuffer[0], sp.packetNum)
								reSend = false
							}
						}
					}
					//when receiver sends CAN - stop transmitting
					if oBuffer[0] == CAN {
						break
					}
				}
				log.Printf("upload done - Success: %d - Failures: %d\n", success, failure)

				//send EOT byte
				var eot []byte
				eot = append(eot, EOT)
				serialPort.Write(eot)
				n := 0
				for n == 1 {
					if n, err = serialPort.Read(oBuffer); err != nil {
						log.Println(err)
					}
					if oBuffer[0] != ACK {
						log.Printf("nexpectedanswer to EOT: 0x%X\n", oBuffer[0])
					} else {
						log.Println("end of transfer")

					}
				}
			}
		}
	}
	refreshSlot()
	return true
}

func sendPacket(p xblock) {

	var sp []byte
	sp = append(sp, p.proto[0])
	sp = append(sp, byte(p.packetNum)+1)
	sp = append(sp, byte(byte(255)-byte(p.packetNum)-1))
	for _, b := range p.payload {
		sp = append(sp, b)
	}
	sp = append(sp, byte(p.checksumm))
	serialPort.Write(sp)
}

func checksum(b []byte, cs byte) byte {
	for _, d := range b {
		cs = cs + d
	}
	return cs
}

//returns false if all payload-bytes are set to 0xff
func (p xblock) checkPaylod() bool {
	var counter = 0
	for _, b := range p.payload {
		if b == 0xff {
			counter++
		}
	}
	if counter == len(p.payload) {
		return false
	}
	return true
}

func downloadSlots() {
	if countSelected() > 1 {
		widgets.QMessageBox_Information(nil, "OK", "please select only one Slot",
			widgets.QMessageBox__Ok, widgets.QMessageBox__Ok)
		return
	}

	var filename string
	var data bytes.Buffer
	for i, s := range Slots {
		hardwareSlot := i
		if Device == Devices.name[1] {
			hardwareSlot = i + 1
		}
		sel := s.slot.IsChecked()
		if sel {
			fileSelect := widgets.NewQFileDialog(nil, 0)
			filename = fileSelect.GetSaveFileName(nil, "save Data from "+s.slotl.Text()+" to File", "", "", "", fileSelect.Options())
			if filename == "" {
				log.Println("no file seleted")
				return
			}
			log.Printf("download a dump from Slot %d into file %s\n", i, filename)
			sendSerialCmd(DeviceActions.selectSlot + strconv.Itoa(hardwareSlot))

			//set chameleon into receiver-mode
			sendSerialCmd(DeviceActions.startDownload)
			if SerialResponse.Code == 110 {

				oBuffer := make([]byte, 1)
				dBuffer := make([]byte, 1024)

				//log.Println("prepare")
				var protocmd []byte
				protocmd = append(protocmd, NAK)
				var (
					transfered = 0
					success    = 0
					failed     = 0
				)

				var getBytes = true
				for getBytes {

					// init tranafer
					if _, err := serialPort.Write(protocmd); err != nil {
						log.Println(err)
						break
					}
					if protocmd[0] == EOT || protocmd[0] == EOF || protocmd[0] == CAN {
						log.Printf("tranfer end.")
						break
					}
					for oBuffer[0] == SOH || oBuffer[0] == ACK {
						if _, err := serialPort.Read(oBuffer); err != nil {
							log.Println(err)
							break
						}
					}
					//rotocmd[0] = oBuffer[0]
					log.Printf("Anser to 0x%X -> 0x%X\n", protocmd[0], oBuffer[0])
					transfered++

					//start receiving blocks
					if getBytes {
						myPacket := xblock{}
						bytesReceived := 0
						blockReceived := false
						for !blockReceived {
							n, err := serialPort.Read(dBuffer)
							bytesReceived = bytesReceived + n
							if err != nil {
								log.Println("Read failed:", err)
							}

							log.Printf("bytesReceived: %d - n: %d\n", bytesReceived, n)


							//on linux (debian 9.3.9 in a viratualbox) I get one responsebyte more
							minBytes:=0
							offset:=0
							switch runtime.GOOS {
							case "linux":
								minBytes=132
							default:
								minBytes=131

							}

							if bytesReceived >= minBytes {
								//offset = bytesReceived - minBytes
								//offset=0
								//if bytesReceived >= 132 {
								//	log.Println("got more as expected ...")
								//	protocmd[0] = NAK
								//	getBytes = false
								//	blockReceived = true
								//}

								log.Printf("Received: (offset: 0)\n%X\n", dBuffer[:bytesReceived])
								myPacket.proto = oBuffer
								myPacket.packetNum = int(dBuffer[offset+0])
								myPacket.packetInv = int(dBuffer[offset+1])
								myPacket.payload = dBuffer[offset+2:offset+130]
								myPacket.checksumm = int(dBuffer[offset+130])
								log.Printf("Received: (offset: %d)\n%X\n",offset, dBuffer[offset:bytesReceived])

								CHK := int(checksum(myPacket.payload, 0))
								if CHK == myPacket.checksumm && myPacket.checkPaylod() {
									//packet OK
									log.Printf("Checksum OK for Packet: %d\n", myPacket.packetNum)
									protocmd[0] = ACK
									success++
									data.Write(myPacket.payload)
								} else {
									//something went wrong
									if !myPacket.checkPaylod() && failed < 10 {

										if byte(myPacket.packetNum) == EOF || byte(myPacket.packetNum) == EOT {
											//EOT & EOF are no failures
											log.Printf("EOF or EOT received (0x%X)\n",byte(myPacket.packetNum))
											failed--
										} else {
											//message for sender
											failed++
											blockReceived = true
											protocmd[0] = NAK

										}
									}
									//stop transfer
									log.Printf("Failed Packet (%d)\n len: %d\nData: %X\n",myPacket.packetNum,bytesReceived,dBuffer[offset:bytesReceived])
									failed-- //the last packet checksum must have missmatched - no error!
									protocmd[0] = CAN
									getBytes = false
								}
								blockReceived=true
							}
						}
						log.Printf("received %d bytes\n", bytesReceived)
					}
				}
				log.Printf("Success: %d - failed: %d\n", success, failed)
			}
			if _, err := serialPort.Write([]byte{CAN}); err != nil {
				log.Println(err)
				break
			}

			slotsize,_:=strconv.Atoi(s.size.Text())
			if data.Len() == slotsize {
				log.Printf("got %d bytes to write to %s... ", data.Len(), filename)
				// Write file
				fOut, err := os.Create(filename)
				if err != nil {
					log.Println(filename, " - write failed")
					log.Fatalln(err)
				}
				fOut.Write(data.Bytes())
				fOut.Close()

				log.Println(filename, " - write successful")
			} else {
				log.Printf("got only %d from %d expected bytes - file not written",data.Len(),slotsize)
			}
		}
	}

}

func populateSlots() {
	if !Connected {
		return
	}

	if populated == false {
		//ToDo: error-handling
		sendSerialCmd(DeviceActions.getModes)
		TagModes = strings.Split(SerialResponse.Payload, ",")
		//ToDo: error-handling
		sendSerialCmd(DeviceActions.getButtons)
		TagButtons = strings.Split(SerialResponse.Payload, ",")
		populated = true
	}

	c := 0
	all := countSelected()

	hardwareSlot := 0
	myProgressBar.zero()
	myProgressBar.widget.SetRange(c, all)
	for sn, s := range Slots {
		//update single slot
		if s.slot.IsChecked() {
			c++
			myProgressBar.update(c)
			if Device == Devices.name[1] {
				hardwareSlot = sn + 1
			} else {
				hardwareSlot = sn
			}

			log.Printf("read data for Slot %d\n", sn+1)
			sendSerialCmd(DeviceActions.selectSlot + strconv.Itoa(hardwareSlot))
			//get slot uid
			sendSerialCmd(DeviceActions.getUid)
			uid := SerialResponse.Payload
			//set uid to lineedit
			s.uid.SetText(uid)

			sendSerialCmd(DeviceActions.getSize)
			size := SerialResponse.Payload

			s.size.SetText(size)

			sendSerialCmd(DeviceActions.getMode)
			mode := SerialResponse.Payload
			_, modeindex := getPosFromList(mode, TagModes)
			s.mode.Clear()
			s.mode.AddItems(TagModes)
			s.mode.SetCurrentIndex(modeindex)
			s.mode.Repaint()

			sendSerialCmd(DeviceActions.getButton)
			buttonl := SerialResponse.Payload
			_, buttonlindex := getPosFromList(buttonl, TagButtons)
			s.btnl.Clear()
			s.btnl.AddItems(TagButtons)
			s.btnl.SetCurrentIndex(buttonlindex)
			s.btnl.Repaint()

			// ToDo: currently mostly faked - currently not implemented in my revG
			//unlear about RButton & LButton short and long -> 4 scenarios?
			sendSerialCmd(DeviceActions.getButton)
			buttons := SerialResponse.Payload
			_, buttonsindex := getPosFromList(buttons, TagButtons)
			s.btns.Clear()
			s.btns.AddItems(TagButtons)
			s.btns.SetCurrentIndex(buttonsindex)
			s.btns.Repaint()
		}
	}
}

//func checkCurrentSelection() {
//	//GetSlotTicker = time.NewTicker(time.Millisecond * 2000)
//	var softSlot int
//	//go func() {
//		//for myTime = range GetSlotTicker.C {
//			sendSerialCmd(DeviceActions.selectedSlot)
//			selected := SerialResponse.Payload
//			if Device == Devices.name[1] {
//				hardSlot, _ := strconv.Atoi(selected)
//				softSlot = hardSlot - 1
//			} else {
//				hardSlot, _ := strconv.Atoi(strings.Replace(selected, "NO.", "", 1))
//				softSlot = hardSlot
//			}
//			log.Printf("Current Selected Slot: %d\n\n", softSlot+1)
//			for i, s := range Slots {
//				if s.slot.IsChecked() && i != softSlot {
//					s.slot.SetChecked(false)
//				} else {
//					if !s.slot.IsChecked() && i == softSlot && populated {
//						s.slot.SetChecked(true)
//					}
//				}
//			}
//		//}
//	//}()
//}

func getPosFromList(val string, array []string) (exists bool, index int) {
	exists = false
	index = -1

	for i, v := range array {
		if val == v {
			index = i
			exists = true
			return
		}
	}

	return
}

func (pb *progressBar) update(c int) {
	pb.widget.SetValue(c)
}

func (pb *progressBar) zero() {
	pb.widget.Reset()
	//pb.widget.SetValue(0)
	//pb.widget.Repaint()
}
