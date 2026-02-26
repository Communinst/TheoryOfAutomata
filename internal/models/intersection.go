package models

import (
	"context"
)

type Sensor struct {
	carCnt int
}

type IntersectionInterface interface {
	Run(ctx context.Context)
	GetPedStatus() bool
	GetSensorCnt(id int) int
	GetAllSensorsCnt() int
	DequeueCars(sensorIDs []int, maxCars int) int
}
