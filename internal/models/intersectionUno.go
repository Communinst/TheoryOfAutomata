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

func (i *IntersectionUno) CheckAndResetPed() bool {
	i.mu.Lock()
	defer i.mu.Unlock()

	if i.isPedestrianWaiting {
		i.isPedestrianWaiting = false
		return true
	}

	return false
}

func (i *IntersectionUno) GetPedStatus() bool {
	return i.isPedestrianWaiting
}

func (i *IntersectionUno) GetSensorCnt(id int) int {
	i.mu.Lock()
	defer i.mu.Unlock()

	return i.sensors[id].carCnt
}
