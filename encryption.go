package main

import (
	"crypto/rand"
	"io"
	"math/big"

	"github.com/septianw/jas/common"
	"golang.org/x/crypto/nacl/secretbox"
)

type Keypair struct {
	Private *[32]byte
	Public  *[32]byte
}

// Encrypt using nacl seal
func Encrypt(in []byte) (out []byte) {
	var random = make([]byte, 32)
	var key [32]byte
	var nonce [24]byte
	var err error

	_, err = rand.Read(random)
	common.ErrHandler(err)
	if err != nil {
		return
	}

	copy(key[:], random)

	_, err = io.ReadFull(rand.Reader, nonce[:])
	common.ErrHandler(err)
	if err != nil {
		return
	}

	encrypted := secretbox.Seal(nonce[:], in, &nonce, &key)

	// generate random gap
	i, _ := rand.Int(rand.Reader, big.NewInt(10))
	gap := uint8(i.Uint64() + 1)
	gapBytes := make([]byte, gap)
	rand.Read(gapBytes)

	out = append(out, key[:]...)
	out = append(out, encrypted...)

	return
}

// decrypt using seal
func Decrypt(in []byte) (out []byte) {
	var key [32]byte
	var nonce [24]byte

	keyIn := in[:32]
	nonceIn := in[32 : 32+24]
	encMsg := in[32+24:]

	copy(key[:], keyIn)
	copy(nonce[:], nonceIn)

	out, ok := secretbox.Open(nil, encMsg, &nonce, &key)
	if !ok {
		panic("decryption error")
		return
	}
	// dari sini sudah decrypted.

	return
}
