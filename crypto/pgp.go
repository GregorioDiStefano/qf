package crypto

import (
	"crypto"
	"errors"
	"io"

	"golang.org/x/crypto/openpgp"
	"golang.org/x/crypto/openpgp/packet"
)

var packetConfigCompression, packetConfigNoCompression packet.Config

func init() {
	packetConfigCompression = packet.Config{
		DefaultHash:            crypto.SHA512,
		DefaultCipher:          packet.CipherAES256,
		DefaultCompressionAlgo: packet.CompressionZIP,
		CompressionConfig: &packet.CompressionConfig{
			Level: packet.BestSpeed,
		},
	}
}

func EncryptFile(src io.Reader, w io.Writer, password []byte) (written int64, err error) {
	var packetConfig packet.Config

	packetConfig = packetConfigCompression

	cipherText, err := openpgp.SymmetricallyEncrypt(w, password, nil, &packetConfig)

	if err != nil {
		return
	}
	defer cipherText.Close()

	if err != nil {
		return
	}

	written, err = io.Copy(cipherText, src)

	if err != nil {
		return
	}
	return
}

func DecryptFile(r io.Reader, w io.Writer, password []byte) (err error) {
	var packetConfig packet.Config
	failed := false
	packetConfig = packetConfigCompression

	prompt := func(keys []openpgp.Key, symmetric bool) ([]byte, error) {
		// If the given passphrase isn't correct, the function will be called again, forever.
		if failed {
			return nil, errors.New("decryption failed")
		}
		failed = true
		return password, nil
	}

	md, err := openpgp.ReadMessage(r, nil, prompt, &packetConfig)

	if err != nil {
		return
	}

	io.Copy(w, md.UnverifiedBody)
	return
}
