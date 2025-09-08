package repository

import (
	"github.com/dragoneena12/measure-review-time/domain/entity"
)

type Printer interface {
	Print(owner, repo string, metrics []*entity.ReviewMetrics) error
}