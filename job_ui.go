package main

import "time"

// ui-ideas
//
// * https://godoc.org/github.com/jroimartin/gocui
//
//  global-progress [..............]
//  plugin1         [....]
//  plugin2         [............]
//  plugin3         [..............]
//
// cons: vertical space
//
// ui-option2:
//   <-> global progress
//  [....|.....|.....|....|....|....]
//   ^
//   | plugin-progress via _-=#░█▓▒░█
//   v
//
// cons: horizontal space
//        plugin-name fehlt

// JobUI defines the interface for all vopher-UIs
type JobUI interface {
	Start()
	Stop()

	Refresh()

	AddJob(jobID string)
	Print(jobID, msg string)
	JobDone(jobID string)
	Wait() // wait for all jobs to be .Done()
}

func newUI(ui string) JobUI {
	switch ui {
	case "oneline":
		return &UIOneLine{
			pt:       newProgressTicker(0),
			prefix:   "vopher",
			duration: 25 * time.Millisecond,
		}
	case "simple":
		return &UISimple{jobs: make(map[string]Runtime)}
	case "quiet":
		return &UIQuiet{}
	}
	return nil
}
