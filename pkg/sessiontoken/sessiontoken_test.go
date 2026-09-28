package sessiontoken

import "testing"

func TestHashMatchesMigrationDigest(t *testing.T) {
	// SELECT encode(sha256(convert_to('abc', 'UTF8')), 'hex');
	const want = "ba7816bf8f01cfea414140de5dae2223b00361a396177a9cb410ff61f20015ad"
	if got := Hash("abc"); got != want {
		t.Fatalf("Hash(abc) = %s, want %s", got, want)
	}
}

func TestNewIsRandomAndNotItsHash(t *testing.T) {
	a, b := New(), New()
	if len(a) != 64 || a == b {
		t.Fatalf("tokens must be 64 hex chars and unique: %q %q", a, b)
	}
	if Hash(a) == a {
		t.Fatal("hash must differ from the token")
	}
}
