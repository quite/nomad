//+build darwin dragonfly freebsd netbsd openbsd solaris windows

package driver

import (
	"github.com/hashicorp/nomad/client/fingerprint"
	"github.com/hashicorp/nomad/helper"
)

func (d *ExecDriver) Fingerprint(req *fingerprint.FingerprintRequest, resp *fingerprint.FingerprintResponse) error {
	d.fingerprintSuccess = helper.BoolToPtr(false)
	resp.Detected = false
	return nil
}
