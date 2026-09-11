package rcpostgres

import (
	"errors"

	"gorm.io/gorm"

	"github.com/KDarenskii/order-service/internal/app/entity"
)

func RowsAffected(res *gorm.DB) int64 {
	return res.RowsAffected
}

func NotFoundIfNoRows(res *gorm.DB) error {
	if res.Error == nil && RowsAffected(res) == 0 {
		return entity.ErrNotFound
	}
	if errors.Is(res.Error, gorm.ErrRecordNotFound) {
		return entity.ErrNotFound
	}
	return res.Error
}

func IsNotFoundError(res *gorm.DB) bool {
	return errors.Is(res.Error, gorm.ErrRecordNotFound)
}
