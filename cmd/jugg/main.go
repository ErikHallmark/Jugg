package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"

	"github.com/erikhallmark/jugg"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/pflag"
)

type cmdArgs struct {
	outputFile string
	inputFile  string
	silent     bool
	verbose    bool
	help       bool
}

func main() {
	initCloseHandler()
	var args cmdArgs
	pflag.StringVarP(&args.outputFile, "output", "o", "", "output file")
	pflag.StringVarP(&args.inputFile, "input", "i", "", "input file")

	pflag.BoolVarP(&args.silent, "silent", "s", false, "silence the output")
	pflag.BoolVarP(&args.help, "help", "h", false, "Show the help menu")
	pflag.BoolVarP(&args.verbose, "verbose", "v", false, "show more details when available")

	pflag.Parse()

	var mode = pflag.Arg(0)

	if args.help {
		printHelpMenu(pflag.Arg(0))
		os.Exit(0)
	}

	switch mode {
	case "monitor":
		monitorPort(args)

	case "list":
		listDevices(args)

	case "send":
		send(args)

	default:
		if mode == "" {
			fmt.Println("Please specify a mode")
		} else {
			fmt.Printf("%s is not a recognized mode. Use jugg --help for usage details \r\n ", mode)
		}
	}

}

func monitorPort(args cmdArgs) {
	port := pflag.Arg(1)
	baudRate, err := strconv.Atoi(pflag.Arg(2))
	if err != nil {
		e := fmt.Sprintf("%s is not a valid baud rate", pflag.Arg(2))
		log.Fatal(e)
	}

	var f *os.File
	if args.outputFile != "" {
		var err error
		f, err = os.Create(args.outputFile)
		if err != nil {
			log.Fatal(err)
		}
	}

	incoming := make(chan jugg.PortData)
	go jugg.MonitorPort(port, baudRate, incoming)
	for {
		update := <-incoming
		output := string(update.Data)
		if update.Err != nil {
			log.Fatal(update.Err)
		}
		if !args.silent {
			fmt.Print(output)
		}
		if args.outputFile != "" {
			_, err := f.WriteString(output)
			if err != nil {
				log.Fatal(err)
			}
		}
	}
}

func send(args cmdArgs) {
	port := pflag.Arg(1)
	baudRate, err := strconv.Atoi(pflag.Arg(2))
	data := []byte(pflag.Arg(3))

	if args.inputFile != "" {
		f, err := os.ReadFile(args.inputFile)
		if err != nil {
			log.Fatal(err)
		}

		data = f
	}

	if err != nil {
		log.Fatal(err)
	}

	jugg.SendData(port, baudRate, []byte(data))

}

func listDevices(args cmdArgs) {
	var devices, err = jugg.ListDevices()
	if err != nil {
		log.Fatal(err)
	}

	if len(devices) == 0 {
		fmt.Println("No serial devices found")
	}

	data := make([][]string, len(devices))
	tableString := &strings.Builder{}
	table := tablewriter.NewWriter((tableString))

	if args.verbose {
		for i := 0; i < len(devices); i++ {
			device := devices[i]
			data[i] = []string{device.Name,
				device.Product,
				device.VID,
				device.PID,
				device.SerialNumber,
				strconv.FormatBool(device.IsUSB)}
		}

		table.SetHeader([]string{"Name", "Product", "VID", "PID", "Serial#", "USB?"})
	} else {
		for i := 0; i < len(devices); i++ {
			device := devices[i]
			data[i] = []string{device.Name,
				device.Product}
		}

		table.SetHeader([]string{"Name", "Product"})
	}

	table.SetBorder(false)
	table.AppendBulk(data)
	table.Render()

	if !args.silent {
		fmt.Print(tableString.String())
	}

	if args.outputFile != "" {
		err := os.WriteFile(args.outputFile, []byte(tableString.String()), 0666)
		if err != nil {
			log.Fatal(err)
		}
	}

}

func printHelpMenu(mode string) {
	switch mode {
	case "":
		fmt.Println("Jugg Serial Port tool")
		fmt.Println("Usage: jugg [mode] <arguments> [options]")
		fmt.Println("Modes:")
		fmt.Println("  list		List the available serial ports")
		fmt.Println("  monitor	Monitors the activity on a serial port ")
		fmt.Println("  send		Send data to a serial port")
		fmt.Println("Arguments:")
		fmt.Println("  --help	-h 	display this help menu")
		fmt.Println("  --output	-o	output file")
		fmt.Println("  --input	-i	input file")
		fmt.Println("  --silent	-s	silent the output")
		fmt.Println("  --verbose	-v 	output more data when available")

	case "monitor":
		fmt.Println("Usage:	jugg monitor <port name> <baud rate> [options]")

	case "list":
		fmt.Println("Usage: jugg list [options]")

	case "send":
		fmt.Println("Usage: jugg send <port name> <baud rate> <data> [options]")

	default:
		fmt.Printf("%s is not recognized \r\n", mode)
		printHelpMenu("")
	}

}

func initCloseHandler() {
	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		os.Exit(0)
	}()
}
