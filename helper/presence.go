package helper

import "sync"

var (
	onlineUsers = make(map[string]bool)
	mu          sync.RWMutex
)

// Set user online
func SetOnline(userID string) {
	mu.Lock()
	defer mu.Unlock()
	onlineUsers[userID] = true
}

// Set user offline
func SetOffline(userID string) {
	mu.Lock()
	defer mu.Unlock()
	delete(onlineUsers, userID)
}

// Cek apakah user online
func IsOnline(userID string) bool {
	mu.RLock()
	defer mu.RUnlock()
	return onlineUsers[userID]
}
