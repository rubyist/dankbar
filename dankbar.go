package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"time"
)

var (
	textColor = "#d3d3d3"
	errColor  = "FF0000"

	batteryCapacity = "/sys/class/power_supply/BAT0/capacity"
	batteryStatus   = "/sys/class/power_supply/BAT0/status"
	batteryFull     = ""
	battery75       = ""
	battery50       = ""
	battery25       = ""
	battery0        = ""
	batteryCharging = ""

	wifiUp   = textColor
	wifiDown = "#4e4a4b"

	evtError = Event{Name: "error", FullText: "E", Color: errColor}

	separatorWidth = 35
)

var things = map[string]func() Event{
	"wifi":    Wifi,
	"battery": Battery,
	"time":    Time,
}

func main() {
	if _, err := os.Stat("/home/scott/.config/dankbar/config.json"); os.IsNotExist(err) {
		os.MkdirAll("/home/scott/.config/dankbar/", 0755)
		ioutil.WriteFile("/home/scott/.config/dankbar/config.json", []byte(`["time"]`), 0644)
	}

	f, _ := os.Open("/home/scott/.config/dankbar/config.json")
	var show []string
	json.NewDecoder(f).Decode(&show)

	fmt.Fprintf(os.Stdout, "{\"version\":1}\n[[]")

	enc := json.NewEncoder(os.Stdout)
	for {
		fmt.Fprintf(os.Stdout, ",") // TODO streaming?
		var events []Event

		for _, thing := range show {
			events = append(events, things[thing]())
		}

		enc.Encode(events)
		time.Sleep(time.Second * 3)
	}
}

func Wifi() Event {
	out, err := exec.Command("iwconfig").Output()
	if err != nil {
		return evtError
	}

	re, err := regexp.Compile("ESSID:(.+)")
	if err != nil {
		return evtError
	}

	ssid := re.Find(out)

	wifiColor := wifiUp
	if len(ssid) == 0 || bytes.Contains(ssid, []byte(":off/any")) {
		wifiColor = wifiDown
	}

	return Event{
		Name:                "wifi",
		FullText:            "",
		Color:               wifiColor,
		Separator:           false,
		SeparatorBlockWidth: separatorWidth,
	}
}

func Time() Event {
	t := time.Now().Format("Jan 2 15:04 ")
	return Event{
		Name:                "time",
		FullText:            t,
		Color:               textColor,
		Separator:           false,
		SeparatorBlockWidth: separatorWidth,
	}
}

func Battery() Event {
	capBy, err := ioutil.ReadFile(batteryCapacity)
	if err != nil {
		return evtError
	}

	cap, err := strconv.Atoi(strings.TrimSpace(string(capBy)))
	if err != nil {
		return evtError
	}

	var batteryText string

	switch {
	case cap > 75:
		batteryText = batteryFull
	case cap > 50:
		batteryText = battery75
	case cap > 25:
		batteryText = battery50
	case cap > 15:
		batteryText = battery25
	default:
		batteryText = battery0
	}

	statBy, err := ioutil.ReadFile(batteryStatus)
	if err != nil {
		return evtError
	}
	if strings.TrimSpace(string(statBy)) == "Charging" {
		batteryText += "  " + batteryCharging
	}

	return Event{
		Name:                "battery",
		FullText:            batteryText,
		Color:               textColor,
		Separator:           false,
		SeparatorBlockWidth: separatorWidth,
	}

}

type Event struct {
	Name                string `json:"name"`
	FullText            string `json:"full_text"`
	Color               string `json:"color"`
	Separator           bool   `json:"separator"`
	SeparatorBlockWidth int    `json:"separator_block_width"`
}
