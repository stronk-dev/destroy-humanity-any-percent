package deploymentbackup

import (
	"crypto/sha256"
	"encoding/hex"
	"testing"
)

func TestRecoveryIdentityPopulationContractsRejectVacuousRows(t *testing.T) {
	domain := RecoveryDomainIdentity{Rows: 1, SHA256: "sha256:" + string(make([]byte, 64))}
	// Use a real hexadecimal digest rather than letting format rejection satisfy
	// the population checks.
	sum := sha256.Sum256([]byte("fixture"))
	domain.SHA256 = "sha256:" + hex.EncodeToString(sum[:])
	empty := RecoveryDomainIdentity{SHA256: domain.SHA256}
	identity := RecoveryIdentity{SchemaVersion: 1, DatabaseMigration: 74, Database: domain, Epoch: domain,
		Player: empty, Founder: empty, Company: empty, Events: empty, Board: empty}
	if err := ValidateEmptyRecoveryIdentity(identity); err != nil {
		t.Fatalf("valid empty identity rejected: %v", err)
	}
	if err := ValidatePopulatedRecoveryIdentity(identity); err == nil {
		t.Fatal("empty identity accepted as populated")
	}
	identity.Player, identity.Founder, identity.Company, identity.Events, identity.Board = domain, domain, domain, domain, domain
	if err := ValidatePopulatedRecoveryIdentity(identity); err != nil {
		t.Fatalf("valid populated identity rejected: %v", err)
	}
	if err := ValidateEmptyRecoveryIdentity(identity); err == nil {
		t.Fatal("populated identity accepted as empty")
	}
}

func TestRecoveryIdentityDigestChangesForSameCountContentMutation(t *testing.T) {
	digest := func(row string) string {
		hasher := sha256.New()
		writeIdentityValue(hasher, []byte("accounts"))
		writeIdentityValue(hasher, []byte(row))
		return hex.EncodeToString(hasher.Sum(nil))
	}
	before := digest(`{"account_id":"01986666-c001-7000-8000-000000000001","recovery_hash":"before"}`)
	after := digest(`{"account_id":"01986666-c001-7000-8000-000000000001","recovery_hash":"after"}`)
	if before == after {
		t.Fatal("same-count row-content mutation did not change recovery identity")
	}
}

func TestRecoveryIdentityComparisonRequiresEveryDomain(t *testing.T) {
	sum := sha256.Sum256([]byte("fixture"))
	domain := RecoveryDomainIdentity{Rows: 1, SHA256: "sha256:" + hex.EncodeToString(sum[:])}
	left := RecoveryIdentity{SchemaVersion: 1, DatabaseMigration: 74, Database: domain, Player: domain,
		Founder: domain, Company: domain, Events: domain, Board: domain, Epoch: domain}
	right := left
	if err := CompareRecoveryIdentity(left, right); err != nil {
		t.Fatal(err)
	}
	right.Board.Rows++
	if err := CompareRecoveryIdentity(left, right); err == nil {
		t.Fatal("changed board identity accepted")
	}
}
