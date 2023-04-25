package web

import (
	"embed"
	"encoding/json"
	"fmt"
	"io/fs"
	"net/http"
	"strings"

	"github.com/aligator/checkpoint"
	"github.com/aligator/keyboard-mod/daemon/led"
)

//go:embed public
var public embed.FS

func ListenAndServe(host string, leds *led.Leds) error {
	mux := http.NewServeMux()

	staticFS, err := fs.Sub(public, "public")
	if err != nil {
		return checkpoint.From(err)
	}
	static := http.FileServer(http.FS(staticFS))
	mux.Handle("/", static)
	mux.HandleFunc("/api/", api("/api/", leds))

	fmt.Println("Listening on", host)
	return http.ListenAndServe(host, mux)
}

func api(prefix string, leds *led.Leds) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			switch r.URL.Path[len(prefix):] {
			case "leds":
				leds, err := leds.GetStatus()
				if err != nil {
					respond(w, http.StatusInternalServerError, nil)
					return
				}
				respond(w, http.StatusOK, leds)
			default:
				respond(w, http.StatusNotFound, nil)
			}
		case http.MethodPatch:
			if strings.HasPrefix(r.URL.Path, prefix+"leds/") {
				ledId := strings.TrimPrefix(r.URL.Path, prefix+"leds/")
				if ledId == "" {
					respond(w, http.StatusBadRequest, nil)
					return
				}

				var ledStatus led.Led
				err := json.NewDecoder(r.Body).Decode(&ledStatus)
				if err != nil {
					respond(w, http.StatusBadRequest, nil)
					return
				}

				ledStatus.Id = ledId

				err = leds.SetLed(ledStatus)
				if err != nil {
					respond(w, http.StatusInternalServerError, nil)
					return
				}

				leds, err := leds.GetStatus()
				if err != nil {
					respond(w, http.StatusInternalServerError, nil)
					return
				}
				respond(w, http.StatusOK, leds)
				return
			}
		}
	}
}

func respond(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	if data == nil {
		return
	}

	err := json.NewEncoder(w).Encode(data)
	if err != nil {
		fmt.Println(err)
	}
}
