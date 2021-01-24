# KeyboardMod

This program is for a Keyboard modification which consists of replacing the normal (mostly 3) keyboard leds by addressable RGB leds.
The original led outputs of the keyboard are instead used as input for the arduino (using simple optocouplers).
This still allows to read and display the original status from the keyboard, but you can freely set the colors used.

I inserted a USB hub circuit board extracted from the smallest USB hub I could find. So I can program and communicate with the arduino using the 
normal USB cable from the keyboard. 

The communication works over normal UART, currently at baud rate 9600.  
It accepts the following comands with <LED_NAME> replaced by the name of the led you want to access and <COLOR> replaced by a 24 bit hex number
(e.g. FF00FF) with always two hex characters being one color in the order R G B.

```
// Set the current color.
<LED_NAME> <COLOR>
// e.g.
// NUM 0000FF
// sets the color of the numlock led to blue

// Set the color used when the keyboard triggers the led. Enables the override mode.
// In this mode the led can be controlled by the keyboard itself and override the manually set color.
<LED_NAME> override <COLOR>
// e.g.
// NUM override 0000FF
// Sets the color of the numlock led to blue if the keyboard wants to enable it. 

// Disable the abillity of the keyboard to override the set color.
<LED_NAME> disable override
// e.g.
// NUM disable override
```

The response starts always with `OK: ` or `ERR: ` to indicate if the command was understood.