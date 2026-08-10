package account

import (
	"bytes"
	"crypto/rand"
	"testing"
)

func TestBootstrapKeyRequiresThirtyTwoRandomBytesAsLowercaseHex(t *testing.T) {
	valid := "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"
	if _, ok := bootstrapDigest(valid); !ok {
		t.Fatal("valid bootstrap key rejected")
	}
	for _, invalid := range []string{"", "a", valid[:63], valid + "0", "G" + valid[1:], "A" + valid[1:]} {
		if _, ok := bootstrapDigest(invalid); ok {
			t.Fatalf("invalid bootstrap key accepted: %q", invalid)
		}
	}
}

func TestBootstrapReceiptAuthenticatedEncryptionBindsDigestAndKeyID(t *testing.T) {
	keys := BootstrapReceiptKeys{CurrentID: "receipt-v1", Current: bytes.Repeat([]byte{0x41}, 32)}
	digest, _ := bootstrapDigest("0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef")
	nonce, ciphertext, err := encryptBootstrapReceipt(rand.Reader, keys, digest, []byte(`{"secret":"credential"}`))
	if err != nil {
		t.Fatal(err)
	}
	rotated := BootstrapReceiptKeys{CurrentID: "receipt-v2", Current: bytes.Repeat([]byte{0x42}, 32), Previous: map[string][]byte{"receipt-v1": keys.Current}}
	plaintext, err := decryptBootstrapReceipt(rotated, digest, "receipt-v1", nonce, ciphertext)
	if err != nil || string(plaintext) != `{"secret":"credential"}` {
		t.Fatalf("rotated decrypt=%s err=%v", plaintext, err)
	}
	tampered := append([]byte(nil), ciphertext...)
	tampered[len(tampered)-1] ^= 1
	if _, err := decryptBootstrapReceipt(rotated, digest, "receipt-v1", nonce, tampered); err == nil {
		t.Fatal("tampered receipt authenticated")
	}
	otherDigest := digest
	otherDigest[0] ^= 1
	if _, err := decryptBootstrapReceipt(rotated, otherDigest, "receipt-v1", nonce, ciphertext); err == nil {
		t.Fatal("receipt authenticated under another request digest")
	}
}
