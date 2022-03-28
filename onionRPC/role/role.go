package role

import (
	"crypto/cipher"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"encoding/json"
	"io"
)

type RoleConfig struct {
	ListenAddr        string
	TracingServerAddr string
	TracingIdentity   string
}

type Role struct {
	SessionKeys map[string]cipher.Block
	FatalError  chan error
}

type ReqExitLayer struct {
	Args          interface{}
	ServiceMethod string
	ServerAddr    string
	Res           interface{}
}
type ReqRelayLayer struct {
	ExitListenAddr string
	ExitSessionId  string
	Encrypted      []byte
}
type ReqGuardLayer struct {
	RelayListenAddr string
	RelaySessionId  string
	Encrypted       []byte
}

func InitRole() Role {
	return Role{
		SessionKeys: make(map[string]cipher.Block),
		FatalError:  make(chan error),
	}
}

func GetPrivateAndPublicKey() (ecdsa.PrivateKey, ecdsa.PublicKey) {
	private, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	public := private.PublicKey
	return *private, public
}

func encode(payload interface{}) []byte {
	//var buf bytes.Buffer
	//gob.NewEncoder(&buf).Encode(payload)
	//return buf.Bytes()
	buf, _ := json.Marshal(payload)
	return buf
}

func encryptAES(payload []byte, cipherBlock cipher.Block) []byte {
	gcm, _ := cipher.NewGCM(cipherBlock)
	nonce := make([]byte, gcm.NonceSize())
	io.ReadFull(rand.Reader, nonce)
	return gcm.Seal(nonce, nonce, payload, nil)
}

func Encrypt(payload interface{}, cipherBlock cipher.Block) []byte {
	return encryptAES(encode(payload), cipherBlock)
}

func decryptAES(payload []byte, cipherBlock cipher.Block) []byte {
	gcm, _ := cipher.NewGCM(cipherBlock)
	nonceSize := gcm.NonceSize()
	nonce, ciphertext := payload[:nonceSize], payload[nonceSize:]
	plaintext, _ := gcm.Open(nil, nonce, ciphertext, nil)
	return plaintext
}
func decode(payload []byte, data interface{}) error {
	//return gob.NewDecoder(bytes.NewBuffer(payload)).Decode(data)
	return json.Unmarshal(payload, data)
}

func Decrypt(payload []byte, cipherBlock cipher.Block, data interface{}) error {
	return decode(decryptAES(payload, cipherBlock), data)
}
