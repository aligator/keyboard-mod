package led

import (
	"bufio"
	"encoding/json"
	"fmt"
	"strings"
	"sync"

	"github.com/aligator/checkpoint"
	"github.com/tarm/serial"
)

type Led struct {
	Id            string `json:"id"`
	Color         int    `json:"color"`
	OverrideColor int    `json:"overrideColor"`
	IsForced      bool   `json:"isForced"`
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

	// err is a channel which will make Run to return an error.
	err chan error

	message     chan Response
	initialized chan struct{}

	ledStatusMutex sync.RWMutex
	ledStatus      []Led
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
	leds.err = make(chan error)
	leds.initialized = make(chan struct{})
	go leds.handleMessage()
	go leds.listen()

	<-leds.initialized
	return leds, checkpoint.From(err)
}

// Run will block until an error occurs.
// The error will never be nil.
func (l *Leds) Run() error {
	err := <-l.err
	l.close()
	return err
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
		case "status":
			cfg := &ledConfig{}
			err := json.Unmarshal([]byte(msg.Msg), cfg)
			if err != nil {
				fmt.Printf("could not parse led config: %v\n", err)
				continue
			}
			fmt.Printf("received config: %v\n", cfg)

			l.ledStatusMutex.Lock()
			l.ledStatus = cfg.Leds
			l.ledStatusMutex.Unlock()

			close(l.initialized)
		case "set":
			status := &Led{}
			err := json.Unmarshal([]byte(msg.Msg), status)
			if err != nil {
				fmt.Printf("could not parse led status: %v\n", err)
				continue
			}

			l.ledStatusMutex.Lock()
			for i, led := range l.ledStatus {
				if led.Id == status.Id {
					l.ledStatus[i] = *status
					break
				}
			}
			l.ledStatusMutex.Unlock()

			fmt.Println("set", status)
		default:
			fmt.Printf("unknown message type: %v\n", msg.Type)
		}
	}
}

func (l *Leds) hello() error {
	return l.Send("status\n")
}

func (l *Leds) Send(cmd string) error {
	// TODO: change that the send blocks until it got an OK or ERR response. (with timeout!)
	// Return that response.
	// Also use a mutex to avoid concurrent writes.

	fmt.Print("sending message: ", cmd)

	_, err := l.stream.Write([]byte(cmd))
	if err != nil {
		return checkpoint.From(err)
	}
	return checkpoint.From(l.stream.Flush())
}

func (l *Leds) GetStatus() ([]Led, error) {
	l.ledStatusMutex.RLock()
	defer l.ledStatusMutex.RUnlock()

	return l.ledStatus, nil
}

func (l *Leds) SetLed(led Led) error {

	// TODO: communicate with json
	return l.Send(fmt.Sprintf("%v %x\n", led.Id, led.Color))
}
