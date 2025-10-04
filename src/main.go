package main

import (
	"fmt"
	"os"
	"strconv"
	"tune-minal/src/metro"
	"tune-minal/src/tuner"
	"tune-minal/src/vis"
)

func main() {

	if len(os.Args) < 1 {
		panic("Usage: tuneminal [-tune [-semitone]] [-metronome [-bpm]]")
	}

	if os.Args[1] == "tune" {
		handleTuner()

	} else if os.Args[1] == "metronome" {
		bpm, err := strconv.Atoi(os.Args[2])
		if err != nil {
			panic("Usage: [-metronome [-bpm]] (bpm must be integer)")
		}

		metro.Start(bpm)
	} else if os.Args[1] == "vis" {
		vis.Start()
	} else {
		fmt.Println("Usage: tuneminal [-tune [-semitone]] [-metronome [-bpm]]")
	}
}

func handleTuner() {
	if len(os.Args) > 3 && os.Args[2] == "tonegen" {
		tuner.StartTonegen(os.Args[3])
	} else {
		tuner.StartTuner()
	}
}
