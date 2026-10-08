//go:build !linux && !darwin

package deploymentbackup

import "errors"

func renameEnvelopeNoReplace(_, _ string) error {
	return errors.ErrUnsupported
}
