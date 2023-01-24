package main

import (
	"github.com/tarm/serial"
	flag "github.com/spf13/pflag"
	"go.bug.st/serial/enumerator"
	"log"
	"fmt"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	InitCloseHandler()
	var port = flag.StringP("port", "p", "COM5", "The com port you wish to access")
	var baud = flag.IntP("baud", "b", 9600,"Baud Rate")
	var help = flag.BoolP("help", "h", false, "Display help menu")
	var file = flag.StringP("output", "o", "", "Output file")

	flag.Parse();

	var command = flag.Arg(0);

	
	if *help {
		printHelpMenu();
		os.Exit(0);	
	}

	switch command{

	case "monitor":
		monitor(*port, baud, *file);
		break;

	case "list":
		listDevices();
		break;

	default:
		if command == "" {
			fmt.Printf("Please specify a what you want to do")
		} else {
			fmt.Printf("%s is not recognized", command)
		}
	}
}

func printHelpMenu() {
	fmt.Print("Jugg Serial Port tool - V0.0 \r\n\r\nUsage: jugg [mode] [arguments] \r\n\r\nModes: \r\n  list		List the available serial port \r\n  monitor	Monitors the activity on a serial port \r\n\r\nArguments:\r\n\r\n  --help	-h 	display this help menu\r\n  --port	-p	set the serial port\r\n  --baud	-b	set the baud rate\r\n  --output	-o	output file\r\n")
}

func listDevices() {
	//TODO: Deal with potential errors
	var list, _ = enumerator.GetDetailedPortsList()

	fmt.Println(list[0].Name)



	for i := 0; i < len(list); i++ {
		var details = list[i]
		var name = details.Name
		var product = details.Product
		fmt.Printf("%s - %s  \r\n", name, product)
	}
}

func monitor(port string, baud *int, file string) {
	var output = file != "";
	var f *os.File
	if output {
		var err error
		f, err = os.Create(file);

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
		fmt.Print(output)

		_, err = f.WriteString(output)
		if err != nil {
			log.Fatal(err)
		}

	}
}

func InitCloseHandler() {
	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		os.Exit(0)
	}()
}
