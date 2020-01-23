package blockchain

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"fmt"
	"github.com/HydroProtocol/ethereum-watcher/utils"
	"github.com/btcsuite/btcd/btcec"
	"golang.org/x/crypto/sha3"
)

var bitCurve = btcec.S256()

func Keccak256(data ...[]byte) []byte {
	d := sha3.NewLegacyKeccak256()
	for _, b := range data {
		d.Write(b)
	}
	return d.Sum(nil)
}

func PersonalEcRecover(data []byte, sig []byte) (string, error) {
	if len(sig) != 65 {
		return "", fmt.Errorf("signature must be 65 bytes long")
	}
	if sig[64] >= 27 {
		sig[64] -= 27
	}

	rpk, err := SigToPub(hashPersonalMessage(data), sig)

	if err != nil {
		return "", err
	}

	if rpk == nil || rpk.X == nil || rpk.Y == nil {
		return "", fmt.Errorf("")
	}
	pubBytes := elliptic.Marshal(bitCurve, rpk.X, rpk.Y)
	return utils.Bytes2Hex(Keccak256(pubBytes[1:])[12:]), nil
}

func hashPersonalMessage(data []byte) []byte {
	msg := fmt.Sprintf("\x19Ethereum Signed Message:\n%d%s", len(data), data)
	return Keccak256([]byte(msg))
}

func SigToPub(hash, sig []byte) (*ecdsa.PublicKey, error) {
	// Convert to btcec input format with 'recovery id' v at the beginning.
	btcSig := make([]byte, 65)
	btcSig[0] = sig[64] + 27
	copy(btcSig[1:], sig)

	pub, _, err := btcec.RecoverCompact(btcec.S256(), btcSig, hash)
	return (*ecdsa.PublicKey)(pub), err
}
