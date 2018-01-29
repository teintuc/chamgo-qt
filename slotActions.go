package main

import (
	"github.com/therecipe/qt/widgets"
	"log"
	"strconv"
	"strings"
	"time"
)

var temp string
var temp2 []string
var myTime time.Time

var GetSlotTicker *time.Ticker

func buttonClicked(btn int) {

	switch ActionButtons[btn] {

	case "Select All":
		selectAllSlots(true)
		if populated {
			GetSlotTicker.Stop()
		}

	case "Select None":
		selectAllSlots(false)
		if populated {
			GetSlotTicker.Stop()
		}

	case "Apply":
		applySlots()

	case "Clear":
		clearSlots()

	case "Refresh":
		refreshSlots()

	case "Set Active":
		activateSlots()

	case "mfkey32":
		mfkey32Slots()

	case "Upload":
		uploadSlots()

	case "Download":
		downloadSlots()

	default:
		log.Printf("Clicked on Button %s\n", ActionButtons[btn])
	}
}

func slotChecked(slot, state int) {
	log.Printf(" Checked %d - state: %d\n", slot, state)
	if state == 2 && Connected {
		if Device == Devices.name[1] {
			//RevG's first Slot is 1 and Last Slot is 8
			temp2 = serialCMD(DeviceActions.selectSlot + strconv.Itoa(slot+1))
		} else {
			//RevE's first Slot is 0 and Last Slot is 7
			temp2 = serialCMD(DeviceActions.selectSlot + strconv.Itoa(slot))
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

func applySlots() {
	GetSlotTicker.Stop()
	for i, s := range Slots {
		sel := s.slot.IsChecked()
		if sel {
			log.Printf("********************\nupdating %s\n", s.slotl.Text())
			hardwareSlot := i
			if Device == Devices.name[1]{
				hardwareSlot=i+1
			}
			temp2 = serialCMD(DeviceActions.selectSlot+strconv.Itoa(hardwareSlot))
			//select slot
			temp2 = serialCMD(Commands.config + "=" + s.mode.CurrentText())
			//set mode
			temp2 = serialCMD(Commands.config + "=" + s.mode.CurrentText())
			//set uid
			temp2 = serialCMD(Commands.uid + "=" + s.uid.Text())
			//set  button short
			temp2 = serialCMD(Commands.button + "=" + s.btns.CurrentText())
			//set button long
			temp2 = serialCMD(Commands.lbutton + "=" + s.btnl.CurrentText())
		}
	}
}

func clearSlots() {
	for i, s := range Slots {
		sel := s.slot.IsChecked()
		if sel {
			log.Printf("I should probably clear settings to Slot %d\n", i)
		}
	}
}

func refreshSlots() {
	//for i,s := range Slots {
	//	sel := s.slot.IsChecked()
	//	if sel {
	//		log.Printf("I should probably refresh settings to Slot %d\n", i)
	//	}
	//}
	// ToDo: bug! - curerently the first run has a offset - on the second run it looks OK
	populateSlots()
	populateSlots()
}

func activateSlots() {
	for i, s := range Slots {
		sel := s.slot.IsChecked()
		if sel {
			log.Printf("I should probably activate Slot %d\n", i)
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

func uploadSlots() {
	var filename string
	fileSelect := widgets.NewQFileDialog(nil, 0)
	filename = fileSelect.GetOpenFileName(nil, "open Dump", "", "", "", fileSelect.Options())

	for i, s := range Slots {
		sel := s.slot.IsChecked()
		if sel {
			log.Printf("I should probably upoload %s to Slot %d\n", filename, i)
		}
	}
}

func downloadSlots() {
	var filename string
	fileSelect := widgets.NewQFileDialog(nil, 0)
	filename = fileSelect.GetOpenFileName(nil, "save Dump", "", "", "", fileSelect.Options())

	for i, s := range Slots {
		sel := s.slot.IsChecked()
		if sel {
			log.Printf("I should probably download a dump from Slot %d into file %s\n", i, filename)
		}
	}

}

func populateSlots() {
	if !Connected {
		return
	}
	if populated == false {
		TagModes = strings.Split(serialCMD(DeviceActions.getModes)[2],",")
		TagButtons =  strings.Split(serialCMD(DeviceActions.getButtons)[2],",")
		//unselect all slots
		buttonClicked(1)
		populated = true
	}
	for _, s := range Slots {
		//select single slot
		s.slot.SetChecked(true)

		//get slot uid
		uid := serialCMD(DeviceActions.getUid)
		//set uid to lineedit
		s.uid.SetText(uid[2])

		size := serialCMD(DeviceActions.getSize)
		s.size.SetText(size[2])

		mode := serialCMD(DeviceActions.getMode)
		_, modeindex := getPosFromList(mode[2], TagModes)
		s.mode.Clear()
		s.mode.AddItems(TagModes)
		s.mode.SetCurrentIndex(modeindex)

		buttonl := serialCMD(DeviceActions.getButton)
		_, buttonlindex := getPosFromList(buttonl[2], TagButtons)
		s.btnl.Clear()
		s.btnl.AddItems(TagButtons)
		s.btnl.SetCurrentIndex(buttonlindex)

		// ToDo: currently mostly faked - currently not implemented in my revG
		buttons := serialCMD(DeviceActions.getButton)
		_, buttonsindex := getPosFromList(buttons[2], TagButtons)
		s.btns.Clear()
		s.btns.AddItems(TagButtons)
		s.btns.SetCurrentIndex(buttonsindex)

		s.slot.SetChecked(false)
	}
}

func checkCurrentSelection() {
	GetSlotTicker = time.NewTicker(time.Millisecond * 2000)
	var softSlot int
	go func() {
		for myTime = range GetSlotTicker.C {
			selected := serialCMD(DeviceActions.selectedSlot)
			if Device == Devices.name[1] {
				hardSlot, _ := strconv.Atoi(selected[2])
				softSlot = hardSlot - 1
			} else {
				hardSlot, _ := strconv.Atoi(strings.Replace(selected[2], "NO.", "", 1))
				softSlot = hardSlot
			}
			log.Printf("Tick at %s - Current Selected Slot: %d\n\n", myTime, softSlot+1)
			for i, s := range Slots {
				if s.slot.IsChecked() && i != softSlot {
					s.slot.SetChecked(false)
				} else {
					if !s.slot.IsChecked() && i == softSlot && populated {
						s.slot.SetChecked(true)
					}
				}
			}
		}
	}()
}

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
