package utils

import (
	"sync"
)

/**
 * @author: HuaiAn xu
 * @date: 2024-03-12 10:39:41
 * @file: volume.go
 * @description: 卷锁
 */

type VolumeLocks struct {
	lock map[string]struct{} //nolint:staticcheck
	mu   sync.Mutex
}

func NewVolumeLocks() *VolumeLocks {
	return &VolumeLocks{
		lock: make(map[string]struct{}),
	}
}

// Lock locks the volume.
func (v *VolumeLocks) Lock(volumeID string) {
	v.mu.Lock()
	v.lock[volumeID] = struct{}{}
	v.mu.Unlock()
}

// Unlock unlocks the volume.
func (v *VolumeLocks) Unlock(volumeID string) {
	v.mu.Lock()
	defer v.mu.Unlock()
	delete(v.lock, volumeID)
}

// IsLocked returns true if the volume is locked.
func (v *VolumeLocks) IsLocked(volumeID string) bool {
	v.mu.Lock()
	defer v.mu.Unlock()
	_, exists := v.lock[volumeID]
	return exists
}

// TryAcquire tries to acquire the lock for the volume. If the lock is already
func (v *VolumeLocks) TryAcquire(volumeID string) bool {
	v.mu.Lock()
	defer v.mu.Unlock()
	if _, exists := v.lock[volumeID]; exists {
		return false
	}
	v.lock[volumeID] = struct{}{}
	return true
}

// Release releases the lock for the volume.
func (v *VolumeLocks) Release(volumeID string) {
	v.mu.Lock()
	defer v.mu.Unlock()
	delete(v.lock, volumeID)
}
