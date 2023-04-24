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

		leds.SetColor("FLOCK", 255, 0, 0)
		leds.SetColor("SHIFT", 0, 255, 0)
		leds.SetColor("NUM", 0, 0, 255)

		leds.Wait()
	} else {
		// run as a cli
	}
}