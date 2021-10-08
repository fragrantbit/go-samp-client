package main

import (
	"time"
)

func NewTask(callback func(), channel *chan bool, 
		b *bool, ms time.Duration, slowing bool) {

	ticker := time.NewTicker(ms * time.Millisecond)
	var i int
	go func() {
		for {		
			select {
			case v := <-*channel:
				if b != nil {
					*b = v
				}
				ticker.Stop()
				return
			// verify it.
			case <-ticker.C:
				callback()
				if slowing {
					i++
					ticker = time.NewTicker(ms + time.Duration(i) * time.Millisecond)
				}
			}	
		}
    }()
}
