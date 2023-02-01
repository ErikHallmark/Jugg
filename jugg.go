package jugg

import (
	"fmt"
	"log"
	"os"

	"github.com/tarm/serial"
	"go.bug.st/serial/enumerator"
)

type PortDetails = enumerator.PortDetails

type PortData struct {
	Data []byte
	Err  error
}

func ListDevices() ([]*PortDetails, error) {
	//TODO: Deal with potential errors
	var list, err = enumerator.GetDetailedPortsList()
	return list, err
}

func MonitorPort(port string, baud int, output chan PortData) {
	c := &serial.Config{Name: port, Baud: baud}
	s, err := serial.OpenPort(c)

	if err != nil {
		output <- PortData{Data: nil, Err: err}
	}

	for {
		buf := make([]byte, 128)
		n, err := s.Read(buf)

		if err != nil {
			log.Fatal(err)
		}

		output <- PortData{Data: buf[:n], Err: nil}
	}
}

func Monitor(port string, baud *int, file string, beSilent bool) {
	var doOutput = file != ""
	var f *os.File
	if doOutput {
		var err error
		f, err = os.Create(file)

		if err != nil {
			log.Fatal(err)
		}
	}

	c := &serial.Config{Name: port, Baud: *baud}
	s, err := serial.OpenPort(c)

	if err != nil {
		log.Fatal(err)
	}

	for true {
		buf := make([]byte, 128)
		n, err := s.Read(buf)

		if err != nil {
			log.Fatal(err)
		}
		var output = fmt.Sprintf("%s", buf[:n])

		if !beSilent {
			fmt.Print(output)
		}

		if doOutput {
			_, err = f.WriteString(output)
			if err != nil {
				log.Fatal(err)
			}
		}
	}
}
