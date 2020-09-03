package jobcheck

import (
	"fmt"
	"testing"
	"time"
)

func TestJobCheck(t *testing.T) {
	jc := NewJobChecker(time.Second, time.Second*5, func(keys []string) error {
		fmt.Println("#. check ", keys)
		return nil
	})

	for {
		cap := jc.Capacity()
		cap <- []string{"1", "2", "3", "4", "5", "11", "12", "13", "14", "15"}
		// cap <- []string{"21", "22", "23", "24", "25", "211", "212", "213", "214", "215"}
		// cap <- []string{}
	}
	time.Sleep(time.Second * 100)
}
