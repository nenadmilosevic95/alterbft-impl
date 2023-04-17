package net

import (
	"fmt"
	"os"
	"time"
)

// Log allows printing string to the standard output prefixed by a timestamp.
// If a Prefix string is configured, the timestamp is followed by the Prefix.
type Log struct {
	Prefix    string
	StartTime time.Time
}

func StartLog(prefix string) Log {
	log := Log{
		Prefix: prefix,
	}
	log.Println("Log started")
	log.StartTime = time.Now()
	return log
}

func (l Log) prefix() string {
	var timestamp string
	if l.StartTime.IsZero() {
		timestamp = time.Now().Format("15:04:05.000000")
	} else {
		duration := time.Now().Sub(l.StartTime)
		timestamp = fmt.Sprintf("%09.6f", duration.Seconds())
	}
	if len(l.Prefix) > 0 {
		return fmt.Sprint(timestamp, " ", l.Prefix)

	}
	return timestamp
}

// Printf is a wrapper to fmt.Printf() method, incremented by a log prefix.
func (l Log) Printf(format string, a ...interface{}) (n int, err error) {
	format = fmt.Sprint(l.prefix(), " ", format)
	return fmt.Fprintf(os.Stdout, format, a...)
}

// Println is a wrapper to fmt.Println() method, incremented by a log prefix.
func (l Log) Println(a ...interface{}) (n int, err error) {
	pars := make([]interface{}, 0, len(a)+1)
	pars = append(pars, l.prefix())
	return fmt.Fprintln(os.Stdout, append(pars, a...)...)
}
