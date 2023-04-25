package main

import (
	"flag"
	"fmt"

	"github.com/aligator/keyboard-mod/daemon/led"
	"github.com/aligator/keyboard-mod/daemon/web"
)

func main() {
	daemon := flag.Bool("d", false, "Run as a daemon")
	host := flag.String("host", "", "Host to listen on for the web ui and the rest api. If empty, the web server will not be started. Example: localhost:8080")
	flag.Parse()

	if *daemon {

		device := flag.Arg(0)
		if device == "" {
			fmt.Fprintf(flag.CommandLine.Output(), "You have to provide a device path as argument\n")
			return
		}

		leds, err := led.OpenLeds(device)

		if err != nil {
			panic(err)
		}

		fmt.Println(leds)

		if *host != "" {
			// Start the embedded web ui
			go func() {
				panic(web.ListenAndServe(*host, leds))
			}()
		}

		panic(leds.Run())
	} else {
		// run as a cli
	}
}
