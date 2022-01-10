package main

import (
	"context"
	"log"
	"time"

	"github.com/mmqbaba/zeus/ratelimit/leaky"
	"github.com/mmqbaba/zeus/ratelimit/token"
)

func main() {
	leaky()
	token()
}

func leaky() {
	rl := leakylimiter.New(100, false)
	prev := time.Now()
	for i := 0; i < 10; i++ {
		now := rl.Take()
		log.Println(i, now.Sub(prev))
		prev = now
	}
}

func token() {
	rl := tokenlimiter.New(100, 10)
	for i := 0; i < 10; i++ {
		if err := rl.Wait(context.Background()); err != nil {
			log.Println("[wait] not allow", err)
		} else {
			log.Println("[wait] allow")
		}
	}

	for i := 0; i < 10; i++ {
		if rl.Allow() {
			log.Println("[allow] allow")
		} else {
			log.Println("[allow] not allow")
		}
	}
}
