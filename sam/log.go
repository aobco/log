package main

import "github.com/aobco/log"

func main() {
	log.InitZapLog("sam.log", "WARN", 100000, 100000, 10000000, 1000000)
	log.Debugf("debug")
	log.Infof("info")
	log.Warnf("warn")
	log.Errorf("error")
}
