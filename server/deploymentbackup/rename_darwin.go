package deploymentbackup

import "golang.org/x/sys/unix"

// macOS development checks use the same no-replace guarantee. This does not
// extend the supported release-host contract beyond Linux/amd64.
func renameEnvelopeNoReplace(source, destination string) error {
	return unix.RenamexNp(source, destination, unix.RENAME_EXCL)
}
