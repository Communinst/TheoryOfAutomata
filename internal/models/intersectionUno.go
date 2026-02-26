package models

import (
	"context"
	"sync"
)

type Direction = int

const (
	North Direction = iota
	South
	West
	East
	totalDirections
)

type Lane = int

const (
	Straight Lane = iota
	Left
	lanePerDir
)

type IntersectionUno struct {
	sensors [totalDirections * lanePerDir]Sensor // Nleft, Sleft ..., Nstraight ... yadayada

	isPedestrianWaiting bool

	PedestrianBtn chan struct{}
	CarArrival    chan CarEvent
	mu            sync.RWMutex
}

func NewIntersectionUno() *IntersectionUno {
	return &IntersectionUno{
		CarArrival:    make(chan CarEvent, 100),
		PedestrianBtn: make(chan struct{}, 10),
	}
}

func (i *IntersectionUno) Run(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case event := <-i.CarArrival:
			i.mu.Lock()
			i.sensors[event.Dir*lanePerDir+event.Lane].carCnt++
			i.mu.Unlock()

		case <-i.PedestrianBtn:
			i.mu.Lock()
			i.isPedestrianWaiting = true
			i.mu.Unlock()
		}
	}
}


func (i *IntersectionUno) GetPedStatus() bool {
	i.mu.Lock()
	defer i.mu.Unlock()

	if i.isPedestrianWaiting {
		i.isPedestrianWaiting = false
		return true
	}
	return false
}

func (i *IntersectionUno) GetSensorCnt(id int) int {
	i.mu.Lock()
	defer i.mu.Unlock()

	return i.sensors[id].carCnt
}

func (i *IntersectionUno) DequeueCars(sensorIDs []int, maxCars int) int {
	i.mu.Lock()
	defer i.mu.Unlock()

	passedCount := 0
	for _, id := range sensorIDs {
		if i.sensors[id].carCnt > 0 {
			if i.sensors[id].carCnt >= maxCars {
				i.sensors[id].carCnt -= maxCars
				passedCount += maxCars
			} else {
				passedCount += i.sensors[id].carCnt
				i.sensors[id].carCnt = 0
			}
		}
	}
	return passedCount
}

func (i *IntersectionUno) GetAllSensorsCnt() int {
	i.mu.Lock()
	defer i.mu.Unlock()
	total := 0
	for _, s := range i.sensors {
		total += s.carCnt
	}
	return total
}
