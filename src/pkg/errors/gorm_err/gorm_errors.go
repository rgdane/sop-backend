package gorm_err

import (
	"errors"
	"strings"

	"gorm.io/gorm"
)

var (
	ErrDataTidakDitemukan       = errors.New("Data tidak ditemukan")
	ErrDuplikasiData            = errors.New("Data sudah ada (duplikat)")
	ErrForeignKeyViolation      = errors.New("Pelanggaran foreign key (data terkait tidak ditemukan)")
	ErrNotNullViolation         = errors.New("Kolom tidak boleh kosong")
	ErrCheckConstraintViolation = errors.New("Pelanggaran check constraint")
	ErrInvalidTransaction       = errors.New("Transaksi tidak valid")
	ErrUnknownError             = errors.New("Terjadi kesalahan server")
)

func TranslateGormError(err error) error {
	if err == nil {
		return nil
	}

	switch {
	case errors.Is(err, gorm.ErrRecordNotFound):
		return ErrDataTidakDitemukan
	case errors.Is(err, gorm.ErrInvalidTransaction):
		return ErrInvalidTransaction
	case strings.Contains(err.Error(), "SQLSTATE 23505"):
		return ErrDuplikasiData // unique_violation
	case strings.Contains(err.Error(), "SQLSTATE 23503"):
		return ErrForeignKeyViolation // foreign_key_violation
	case strings.Contains(err.Error(), "SQLSTATE 23502"):
		return ErrNotNullViolation // not_null_violation
	case strings.Contains(err.Error(), "SQLSTATE 23514"):
		return ErrCheckConstraintViolation // check_violation
	default:
		return err
	}
}
