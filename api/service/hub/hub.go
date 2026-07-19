package hub

import "sync"

type Hub struct {
	mu sync.Mutex
	
}