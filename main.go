package main

import (
	"flag"
	"fmt"

	"github.com/aligator/keyboard-mod/core/led"
)

func main() {
	daemon := flag.Bool("d", false, "Run as a deamon")
	flag.Parse()

	if *daemon {
		leds, err := led.OpenLeds("/dev/ttyACM1")

		if err != nil {
			panic(err)
		}

		fmt.Println(leds)
		leds.Wait()
	} else {
		// run as a cli
	}
}
