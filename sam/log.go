package main

import "github.com/aobco/log"

func main() {
	size()
	date()
}

func size() {
	log.Init("sam.log", "WARN", 1, 2, 2, log.RollingBySize, true)
	log.Debugf("debug")
	log.Infof("info")
	log.Warnf("warn")
	log.Errorf("error")
	println("done")
}

func date() {
	log.Init("sam.log", "WARN", 1, 2, 2, log.RollingByDate)
	log.Debugf("debug")
	log.Infof("info")
	log.Warnf("warn")
	log.Errorf("error")
	println("done")
}
