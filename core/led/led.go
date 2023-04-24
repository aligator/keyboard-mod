package led

import (
	"bufio"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/aligator/checkpoint"
	"github.com/tarm/serial"
)

type Led struct {
	Id string `json:"id"`
}

type ledConfig struct {
	Name    string `json:"name"`
	Version string `json:"version"`
	Leds    []Led  `json:"leds"`
}

type Response struct {
	Ok   bool
	Type string
	Msg  string
}

type Leds struct {
	stream *serial.Port

	message chan Response
	done    chan struct{}

	Leds []Led
}

func OpenLeds(device string) (*Leds, error) {
	leds := &Leds{}

	config := &serial.Config{
		Name: device,
		Baud: 9600,
		Size: 8,
	}
	stream, err := serial.OpenPort(config)
	if err != nil {
		return nil, checkpoint.From(err)
	}
	leds.stream = stream

	leds.message = make(chan Response, 10)
	leds.done = make(chan struct{})
	go leds.handleMessage()
	go leds.listen()

	return leds, checkpoint.From(err)
}

func (l *Leds) Wait() {
	<-l.done
	l.close()
}

func (l *Leds) close() {
	_ = l.stream.Close()
	close(l.message)
}

func (l *Leds) listen() {
	scanner := bufio.NewScanner(l.stream)
	scanner.Split(bufio.ScanLines)

	err := l.hello()
	if err != nil {
		panic(err)
	}

	for {
		open := scanner.Scan()
		if !open {
			if scanner.Err() != nil {
				l.message <- Response{
					Ok:  false,
					Msg: scanner.Err().Error(),
				}
			}
			continue
		}

		line := scanner.Text()

		if strings.HasPrefix(line, "OK: ") {
			line = strings.TrimPrefix(line, "OK: ")

			// Get the type of the message.
			splitted := strings.SplitN(line, " ", 2)
			msgType := splitted[0]
			l.message <- Response{
				Ok:   true,
				Type: msgType,
				Msg:  splitted[1],
			}
		}

		if strings.HasPrefix(line, "ERR: ") {
			l.message <- Response{
				Ok:  false,
				Msg: strings.TrimPrefix(line, "ERR: "),
			}
		}

		// Drop all other messages.
	}
}

// handleMessage reads from the message channel and handles the messages.
func (l *Leds) handleMessage() {
	for {
		msg, open := <-l.message
		if !open {
			break
		}

		fmt.Printf("received message: %v\n", msg)

		if !msg.Ok {
			fmt.Printf("received error: %v\n", msg.Msg)
			continue
		}

		switch msg.Type {
		case "hello":
			cfg := &ledConfig{}
			err := json.Unmarshal([]byte(msg.Msg), cfg)
			if err != nil {
				fmt.Printf("could not parse led config: %v\n", err)
				continue
			}
			fmt.Printf("received config: %v\n", cfg)
			l.Leds = cfg.Leds
		default:
			fmt.Printf("unknown message type: %v\n", msg.Type)
		}
	}
}

func (l *Leds) hello() error {
	return l.Send("hello\n")
}

func (l *Leds) Send(cmd string) error {
	_, err := l.stream.Write([]byte(cmd))
	return checkpoint.From(err)
}

func (l *Leds) SetColor(name string, r, g, b int) error {
	return nil
}
