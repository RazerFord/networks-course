package logging

import "log"

func init() {
	log.SetFlags(log.Flags() ^ (log.Ldate | log.Ltime))
}

func Info(format string, args ...any) {
	log.Printf("[ INFO ] "+format, args...)
}

func Warn(format string, args ...any) {
	log.Printf("[ ERROR ] "+format, args...)
}
