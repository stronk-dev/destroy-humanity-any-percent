package deploymentrelease

import "testing"

func TestRestartCourtesyRequiresExactSystemEnvelope(t *testing.T) {
	if !isRestartCourtesy([]byte(`{"kind":"system","payload":{"code":"server_restarting"}}`)) {
		t.Fatal("exact restart courtesy frame rejected")
	}
	for _, invalid := range [][]byte{
		[]byte(`{"kind":"system","payload":{"code":"server_restart"}}`),
		[]byte(`{"kind":"game","payload":{"code":"server_restarting"}}`),
		[]byte(`{"kind":"system","payload":{}}`),
	} {
		if isRestartCourtesy(invalid) {
			t.Fatalf("non-courtesy frame accepted: %s", invalid)
		}
	}
}
