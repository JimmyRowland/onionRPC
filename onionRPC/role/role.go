package role

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
)

type RoleConfig struct {
	ListenAddr        string
	TracingServerAddr string
	TracingIdentity   string
}

type Role struct {
	SessionKeys map[string]*ecdsa.PrivateKey
	FatalError  chan error
}

func InitRole() Role {
	return Role{
		SessionKeys: make(map[string]*ecdsa.PrivateKey),
		FatalError:  make(chan error),
	}
}

func GetPrivateAndPublicKey() (ecdsa.PrivateKey, ecdsa.PublicKey) {
	private, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	public := private.PublicKey
	return *private, public
}
