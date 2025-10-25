package main

import (
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/log"
	"github.com/the-jonsey/pulseaudio"
	"gitlab.com/gomidi/midi/v2"
	"gitlab.com/gomidi/midi/v2/drivers"
	_ "gitlab.com/gomidi/midi/v2/drivers/rtmididrv" // autoregisters driver

	"github.com/charmbracelet/bubbles/progress"
)

const (
	padding  = 2
	maxWidth = 80
)

type channel struct {
	Sink         string
	Vol          float32
	Prog         progress.Model
	IgnoreMaster bool

	OnColor  uint8
	OffColor uint8
}

var helpStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#626262")).Render

type tickMsg time.Time

var channels = map[int]*channel{
	0: {Sink: "general-output", Vol: 0, OnColor: 0x31},
	1: {Sink: "discord-output", Vol: 0, OnColor: 0x38},
	2: {Sink: "tidal-output", Vol: 0, OnColor: 0x15},

	7: {Sink: "mic", Vol: 0, IgnoreMaster: true, OnColor: 0x14},
}

var channel_values = make([]float32, 8)
var muted_channels = map[int]bool{}
var muted_previous_values = map[int]float32{}
var boosted_channels = map[int]bool{}

var master = float32(1)

var out drivers.Out
var client *pulseaudio.Client

func sliderToVolume(slider float32, isBoosted bool) float32 {
	minVol := float32(0.15)
	maxVol := float32(1.1)
	if isBoosted {
		minVol = 0.3
		maxVol = 5.0
	}

	if slider == 0 {
		return 0
	}

	return minVol + (maxVol-minVol)*slider
}

var previousStates = map[uint8][2]uint8{}

var outMut = sync.Mutex{}

func sendMidi(channel uint8, key uint8, velocity uint8) {
	outMut.Lock()
	defer outMut.Unlock()

	if d, ok := previousStates[key]; ok && d[0] == channel && d[1] == velocity {
		return
	}
	out.Send(midi.NoteOn(channel, key, velocity))
	previousStates[key] = [2]uint8{channel, velocity}
}

func updateVolumes() {
	for ch := 0; ch < 8; ch++ {

		if v := channels[ch]; v == nil {
			continue
		}
		val := channel_values[ch]
		if !channels[ch].IgnoreMaster {
			val = val * master
		}
		prog := int(val * 8)

		c := channels[ch]
		colorOn := c.OnColor

		if m, ok := muted_channels[ch]; ok && m {
			colorOn = 0x5
		}

		for i := 0; i < 8; i++ {
			if i == prog && i != 0 {
				sendMidi(6, uint8(i*8)+uint8(ch), 0x3)
			} else {
				brightness := uint8(1)
				color := colorOn
				if i < prog || (prog == 0 && i == 0) {
					brightness = 6
					color = c.OnColor
				}
				sendMidi(brightness, uint8(i*8)+uint8(ch), color)

			}
		}

		isBoosted := false
		if v, ok := boosted_channels[ch]; v && ok {
			isBoosted = true
		}

		channels[ch].Vol = val
		if v, ok := muted_channels[ch]; v && ok {
			val = 0
		}
		client.SetSinkVolume(channels[ch].Sink, sliderToVolume(val, isBoosted))
	}
}

var micMutedByMaster = true

var muteMutexes = map[int]*sync.Mutex{}

func muteChannel(key int) {
	muted_channels[int(key)] = true
	muted_previous_values[int(key)] = channel_values[key]

	go func() {
		for i := 0; i < 12; i++ {
			if !muted_channels[int(key)] {
				return
			}
			channel_values[key] = channel_values[key] * (1 - float32(i)/11)
			time.Sleep(15 * time.Millisecond)
			go updateVolumes()
		}
	}()
}

func unmuteChannel(key int) {
	muted_channels[int(key)] = false
	previous := muted_previous_values[int(key)]
	fmt.Println("going back to", previous)

	go func() {
		for i := 0; i < 12; i++ {
			if muted_channels[int(key)] {
				return
			}
			channel_values[key] = float32(previous) * float32(i) / 11
			go updateVolumes()

			time.Sleep(15 * time.Millisecond)
		}
	}()
}

func evalMaster(vel uint8) {
	fmt.Println("master", vel)
	master = float32(vel) / 127
	if master == 0 {
		fmt.Println("mastermute")
		go muteChannel(7)
		micMutedByMaster = true
	} else if micMutedByMaster {
		fmt.Println("masterunmute")
		unmuteChannel(7)
		micMutedByMaster = false
	}

}

func main() {
	logger := log.New(os.Stdout)

	logger.Info("Starting up...")
	defer midi.CloseDriver()

	in, err := midi.FindInPort("APC mini mk2")
	if err != nil {
		logger.Fatal("Can't find midi. Don't mean to be rude but.. Do you have a nanoKONTROL2? Is it plugged in? This program only works with a nanoKONTROL2", "error", err)
	}
	// connect to the system's pulseaudio server
	client, err = pulseaudio.NewClient()

	if err != nil {
		logger.Fatal("Can't connect to pulseaudio. Is the program running on the wrong user?", "error", err)
	}

	out, err = midi.FindOutPort("APC mini mk2")
	if err != nil {
		logger.Fatal("Can't find midi. Don't mean to be rude but.. Do you have a nanoKONTROL2? Is it plugged in? This program only works with a nanoKONTROL2", "error", err)
	}
	out.Open()
	defer out.Close()

	_, err = midi.ListenTo(in, func(msg midi.Message, timestampms int32) {
		var bt []byte
		var ch, key, vel uint8
		switch {
		case msg.GetSysEx(&bt):
			fmt.Println("Got SysEx", bt[6:15])
			for i, v := range bt[6:14] {
				channel_values[i] = float32(v) / 127
			}
			evalMaster(bt[14])
			go updateVolumes()

		case msg.GetNoteOn(&ch, &key, &vel):
			{
				if key < 8 {
					m, ok := muted_channels[int(key)]
					if !ok {
						m = false
					}
					if m {
						unmuteChannel(int(key))
						return
					}
					muteChannel(int(key))
				}

				if key >= 100 && key < 109 {
					ch := int(key - 100)
					fmt.Println("boost", ch)
					if v, ok := boosted_channels[ch]; !ok || !v {
						boosted_channels[ch] = true
						out.Send(midi.NoteOn(0, key, 127))
						go updateVolumes()
						break
					}
					boosted_channels[ch] = false
					out.Send(midi.NoteOn(0, key, 0))
					go updateVolumes()
				}
			}

		case msg.GetControlChange(&ch, &key, &vel):
			c := key - 48
			if c > 7 {
				evalMaster(vel)
			} else {
				muted_channels[int(c)] = false
				val := float32(vel) / 127
				fmt.Println(val)
				channel_values[c] = val
			}
			fmt.Println(c, vel)
			go updateVolumes()
		}
	}, midi.UseSysEx())

	fmt.Println("listening...")
	if err != nil {
		logger.Fatal("Can't listen to midi", "error", err)
	}

	err = out.Send([]byte{
		0xF0,
		0x47,
		0x7F,
		0x4F,
		0x60,
		0x00,
		0x04,
		0x00,

		0x00,
		0x00,
		0x00,

		0xF7,
	})
	fmt.Println(err)

	go updateVolumes()

	select {}
}
