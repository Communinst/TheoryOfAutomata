package models

import (
	"context"
)

type Sensor struct {
	carCnt int
}

type IntersectionInterface interface {
	Run(ctx context.Context)
	CheckAndResetPed() bool
	GetPedStatus() bool
	GetSensorCnt(id int) int
}
