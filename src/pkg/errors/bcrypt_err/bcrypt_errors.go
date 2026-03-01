package bcrypt_err

import (
	"errors"

	"golang.org/x/crypto/bcrypt"
)

var (
	ErrPasswordSalah = errors.New("Password tidak sesuai")
)

func TranslateBcryptError(err error) error {
	if err == nil {
		return nil
	}

	switch {
	case errors.Is(err, bcrypt.ErrMismatchedHashAndPassword):
		return ErrPasswordSalah
	default:
		return err
	}
}
