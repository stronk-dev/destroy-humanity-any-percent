package operations

import "syscall"

// FilesystemUsage reads the used fraction of the filesystem containing path.
// The host observer and its real-filesystem integration use this same producer.
func FilesystemUsage(path string) (float64, error) {
	var state syscall.Statfs_t
	if syscall.Statfs(path, &state) != nil || state.Blocks == 0 || state.Bavail > state.Blocks {
		return 0, ErrInvalid
	}
	return float64(state.Blocks-state.Bavail) / float64(state.Blocks), nil
}
