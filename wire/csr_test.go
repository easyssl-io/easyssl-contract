package wire

import (
	"crypto"
	"crypto/ecdsa"
	"crypto/ed25519"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/rsa"
	"testing"
)

func TestKeyTypeOf(t *testing.T) {
	ec := func(c elliptic.Curve) crypto.PublicKey {
		k, err := ecdsa.GenerateKey(c, rand.Reader)
		if err != nil {
			t.Fatal(err)
		}
		return k.Public()
	}
	rsaKey := func(bits int) crypto.PublicKey {
		k, err := rsa.GenerateKey(rand.Reader, bits)
		if err != nil {
			t.Fatal(err)
		}
		return k.Public()
	}
	edPub, _, _ := ed25519.GenerateKey(rand.Reader)
	for _, c := range []struct {
		pub  crypto.PublicKey
		want string
	}{
		{ec(elliptic.P256()), KeyTypeECP256},
		{ec(elliptic.P384()), KeyTypeECP384},
		{ec(elliptic.P521()), ""},
		{rsaKey(2048), KeyTypeRSA2048},
		{rsaKey(3072), KeyTypeRSA3072},
		{rsaKey(1024), ""},
		{edPub, ""},
	} {
		if got := KeyTypeOf(c.pub); got != c.want {
			t.Errorf("KeyTypeOf(%T) = %q, want %q", c.pub, got, c.want)
		}
	}
	for _, k := range KeyTypes {
		if !ValidKeyType(k) {
			t.Errorf("ValidKeyType(%q) = false", k)
		}
	}
	if ValidKeyType("ec_p521") {
		t.Error("ValidKeyType accepted an unlisted type")
	}
}
