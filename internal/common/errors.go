package common

import "log"

func IgnoreErr(f func() error) {
	if err := f(); err != nil {
		log.Println("Error:", err)
	}
}
