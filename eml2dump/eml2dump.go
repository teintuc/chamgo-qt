package eml2dump

import (
	"encoding/hex"
	"io/ioutil"
	"os"

	"github.com/sirupsen/logrus"
)

func Bytes2File(filename string, filebytes []byte) (err error) {
	//open file for write
	outFile, err := os.OpenFile(filename, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		logrus.Error(err)
		return err
	}
	defer outFile.Close()

	//write data to file
	if _, err = outFile.Write(filebytes); err != nil {
		logrus.Error(err)
		return err
	}
	logrus.Debugf("wrote %d bytes to file: %s", len(filebytes), filename)
	return nil
}

func File2Bytes(filename string) (filedata []byte, err error) {

	//does the fikle exists?
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		logrus.Error(err)
		return filedata, err
	}

	//read filedata
	filedata, _ = ioutil.ReadFile(filename)
	logrus.Debugf("read %d bytes from file: %s", len(filedata), filename)
	return filedata, err
}

func Bytes2Emul(filename string, data []byte) bool {
	encoded := hex.EncodeToString(data)
	//open file for write
	outFile, err := os.OpenFile(filename, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		logrus.Error(err)
		return false
	}
	defer outFile.Close()
	//write 32char-lines
	for i := 1; i < len(encoded)+1; i++ {
		outFile.WriteString(string(encoded[i-1]))
		if err != nil {
			logrus.Error(err)
			return false
		}
		if i%32 == 0 && i < len(encoded)-1 {
			outFile.WriteString("\n")
		}
	}
	return true
}
