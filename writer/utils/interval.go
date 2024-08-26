package utils

import "time"

func SetInterval(intervalAction func(), interval time.Duration) {
	ticker := time.Tick(interval)

	for range ticker {
		intervalAction()
	}
}
