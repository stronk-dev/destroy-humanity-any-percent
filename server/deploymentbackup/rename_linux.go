package deploymentbackup

import "golang.org/x/sys/unix"

// Do not fall back to replacing rename if the target filesystem cannot supply
// the atomic no-replace guarantee required for a committed backup identity.
func renameEnvelopeNoReplace(source, destination string) error {
	return unix.Renameat2(unix.AT_FDCWD, source, unix.AT_FDCWD, destination, unix.RENAME_NOREPLACE)
}
