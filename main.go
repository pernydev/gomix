package main

import (
	"fmt"

	"github.com/the-jonsey/pulseaudio"
	"gitlab.com/gomidi/midi/v2"
	_ "gitlab.com/gomidi/midi/v2/drivers/rtmididrv" // autoregisters driver
)

func main() {
	defer midi.CloseDriver()

	in, err := midi.FindInPort("nanoKONTROL2")
	if err != nil {
		fmt.Println("can't find midi")
		return
	}

	if err != nil {
		fmt.Println("can't find midi")
		return
	}

	// connect to the system's pulseaudio server
	client, err := pulseaudio.NewClient()

	if err != nil {
		fmt.Printf("ERROR: %s\n", err)
		return
	}

	bindings := map[uint8]string{
		0: "dc-source",
		1: "dc-output",
		2: "fun-toheadphones",
		3: "fun-tocord",
		4: "fun-tospeakers",
	}

	stop, err := midi.ListenTo(in, func(msg midi.Message, timestampms int32) {
		var bt []byte
		var ch, key, vel uint8
		switch {
		case msg.GetSysEx(&bt):
			fmt.Printf("got sysex: % X\n", bt)
		case msg.GetControlChange(&ch, &key, &vel):
			fmt.Printf("got control change: %d %d %d\n", ch, key, vel)
			if bindings[key] != "" {
				client.SetSinkVolume(bindings[key], float32(vel)/127)
				fmt.Printf("Setting %s to %f\n", bindings[key], float32(vel)/127)
			}
		}

	}, midi.UseSysEx())

	if err != nil {
		fmt.Printf("ERROR: %s\n", err)
		return
	}

	fmt.Scanln()

	stop()
}
