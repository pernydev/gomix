package main

import (
	"os"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/log"
	"github.com/the-jonsey/pulseaudio"
	"gitlab.com/gomidi/midi/v2"
	_ "gitlab.com/gomidi/midi/v2/drivers/rtmididrv" // autoregisters driver

	"github.com/charmbracelet/bubbles/progress"
	tea "github.com/charmbracelet/bubbletea"
)

const (
	padding  = 2
	maxWidth = 80
)

type channel struct {
	Sink string
	Vol  float32
	Prog progress.Model
	Name string
}

var helpStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#626262")).Render

type tickMsg time.Time

var channels = []channel{
	0: {Sink: "dc_source", Vol: 0, Name: "Mic          "},
	1: {Sink: "dc_output", Vol: 0, Name: "DC Output    "},
	2: {Sink: "fun_to_headphones", Vol: 0, Name: "Headphones   "},
	3: {Sink: "fun_to_dc", Vol: 0, Name: "Desktop to DC"},
	4: {Sink: "fun_to_speakers", Vol: 0, Name: "Speakers     "},
}

type model struct {
}

func (m model) Init() tea.Cmd {
	return tickCmd()
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		return m, tea.Quit

	case tea.WindowSizeMsg:
		/*
			m.progress.Width = msg.Width - padding*2 - 4
			if m.progress.Width > maxWidth {
				m.progress.Width = maxWidth
			}
		*/

		for ch, channel := range channels {
			channel.Prog.Width = msg.Width - padding*2 - 4
			if channel.Prog.Width > maxWidth {
				channel.Prog.Width = maxWidth
			}
			channels[ch] = channel
		}

		return m, nil
	case tickMsg:
		return m, tickCmd()
	default:
		return m, nil
	}
}

func (m model) View() string {
	pad := strings.Repeat(" ", padding)
	var b strings.Builder
	for _, channel := range channels {
		b.WriteString(pad)
		b.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("#626262")).Render(channel.Name))
		b.WriteString(pad)
		b.WriteString(channel.Prog.ViewAs(float64(channel.Vol)))
		b.WriteString("\n\n")
	}
	b.WriteString(helpStyle("Press any key to quit"))
	return b.String()

}

func tickCmd() tea.Cmd {
	return tea.Tick(time.Millisecond*50, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

func main() {
	logger := log.New(os.Stdout)

	logger.Info("Starting up...")
	defer midi.CloseDriver()

	in, err := midi.FindInPort("nanoKONTROL2")
	if err != nil {
		logger.Fatal("Can't find midi. Don't mean to be rude but.. Do you have a nanoKONTROL2? Is it plugged in? This program only works with a nanoKONTROL2", "error", err)
	}
	// connect to the system's pulseaudio server
	client, err := pulseaudio.NewClient()

	if err != nil {
		logger.Fatal("Can't connect to pulseaudio. Is the program running on the wrong user?", "error", err)
	}

	for ch, channel := range channels {
		channel.Prog = progress.NewModel(progress.WithDefaultGradient())
		channel.Prog.ShowPercentage = true
		channels[ch] = channel
	}

	stop, err := midi.ListenTo(in, func(msg midi.Message, timestampms int32) {
		var bt []byte
		var ch, key, vel uint8
		switch {
		case msg.GetSysEx(&bt):
			logger.Debug("Got SysEx", bt)
		case msg.GetControlChange(&ch, &key, &vel):
			if int(key) < len(channels) {
				channels[key].Vol = float32(vel) / 127
				client.SetSinkVolume(channels[key].Sink, channels[key].Vol)
			}
		}
	}, midi.UseSysEx())

	if err != nil {
		logger.Fatal("Can't listen to midi", "error", err)
	}

	if err := tea.NewProgram(model{}).Start(); err != nil {
		logger.Fatal("Can't start program", "error", err)
	}

	stop()
}
