package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	"github.com/erikhallmark/jugg"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/pflag"
)

type cmdArgs struct {
	port       string
	outputFile string
	baudRate   int
	silent     bool
	verbose    bool
	help       bool
}

func main() {
	initCloseHandler()
	var args cmdArgs
	pflag.StringVarP(&args.port, "port", "p", "", "The com port you want to connect to")
	pflag.StringVarP(&args.outputFile, "output", "o", "", "The name of a file you'd like Jugg to output too")

	pflag.IntVarP(&args.baudRate, "baud", "b", 115200, "The baud rate of the device your connecting to (default is 115200)")

	pflag.BoolVarP(&args.silent, "silent", "s", false, "Used to silence the console output")
	pflag.BoolVarP(&args.help, "help", "h", false, "Show the help menu")
	pflag.BoolVarP(&args.verbose, "verbose", "v", false, "show more details")

	pflag.Parse()

	var mode = pflag.Arg(0)

	if args.help {
		printHelpMenu()
		os.Exit(0)
	}

	switch mode {
	case "monitor":
		monitorPort(args)

	case "list":
		listDevices(args)

	default:
		if mode == "" {
			fmt.Println("Please specify a mode")
		} else {
			fmt.Printf("%s is not a recognized mode", mode)
		}
	}

}

func monitorPort(args cmdArgs) {
	incoming := make(chan jugg.PortData)
	go jugg.MonitorPort(args.port, args.baudRate, incoming)

	for {
		update := <-incoming

		if update.Err != nil {
			log.Fatal(update.Err)
		}

		fmt.Printf("%s", update.Data)
	}
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
	table := tablewriter.NewWriter((os.Stdout))

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

	// for i := 0; i < len(devices); i++ {
	// 	var details = devices[i]

	// 	if args.verbose {

	// 	} else {
	// 		fmt.Printf("(%02d) %s - %s \r\n", i, details.Name, details.Product)
	// 	}
	// }

}

func printHelpMenu() {
	fmt.Println("Jugg Serial Port tool - V0.0.0 ")
	fmt.Println("Usage: jugg [mode] [arguments] ")
	fmt.Println("Modes:")
	fmt.Println("  list		List the available serial port")
	fmt.Println("  monitor	Monitors the activity on a serial port ")
	fmt.Println("Arguments:")
	fmt.Println("  --help	-h 	display this help menu")
	fmt.Println("  --port	-p	set the serial port")
	fmt.Println("  --baud	-b	set the baud rate")
	fmt.Println("  --output	-o	output file")
	fmt.Println("  --silent	-s	silent the output")
}

func initCloseHandler() {
	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		os.Exit(0)
	}()
}
