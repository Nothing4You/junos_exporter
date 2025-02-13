// SPDX-License-Identifier: MIT

package connector

import (
	"io"

	"github.com/pkg/errors"
	"golang.org/x/crypto/ssh"
)

func loadPrivateKey(r io.Reader) (ssh.AuthMethod, error) {
	b, err := io.ReadAll(r)
	if err != nil {
		return nil, errors.Wrap(err, "could not read from reader")
	}

	key, err := ssh.ParsePrivateKey(b)
	if err != nil {
		return nil, errors.Wrap(err, "could not parse private key")
	}

	return ssh.PublicKeys(key), nil
}
