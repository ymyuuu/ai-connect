package common

import "log"

func IgnoreErr(err error) {
	log.Printf("err: %v", err)
}
