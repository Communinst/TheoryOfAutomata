package main

import (
	"bufio"
	"context"
	"fmt"
	"math/rand"
	"os"
	"strings"
	"time"

	"github.com/Communinst/TheoryOfAutomata/internal/fsm"
	"github.com/Communinst/TheoryOfAutomata/internal/models"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	reader := bufio.NewReader(os.Stdin)

	fmt.Println("🚦 Traffic Light Simulation")
	fmt.Println("Time to select the mode.")
	fmt.Println("  0 - Cycle (FIXED)")
	fmt.Println("  1 - Adaptive Control (ADAPTIVE)")
	fmt.Print("Your choice (0 by default): ")

	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(input)

	var mode fsm.Mode
	if input == "1" {
		mode = fsm.ADAPTIVE
		fmt.Println("✅ ADAPTIVE mode selected.")
	} else {
		mode = fsm.FIXED
		fmt.Println("✅ FIXED mode selected.")
	}

	fmt.Println("--------------------------------------------------")

	intersection := models.NewIntersectionUno()

	go intersection.Run(ctx)
	go SpawnTraffic(ctx, 6, models.North, intersection.CarArrival, 15)
	go SpawnTraffic(ctx, 6, models.South, intersection.CarArrival, 15)
	go SpawnTraffic(ctx, 12, models.West, intersection.CarArrival, 5)
	go SpawnTraffic(ctx, 12, models.East, intersection.CarArrival, 5)

	time.Sleep(1 * time.Second)

	trafficLight := fsm.NewTrafficMachine(
		fsm.NSG_STRAIGHT_RIGHT,
		100,
		15*time.Second,
		mode,
		intersection,
	)

	go trafficLight.Run(ctx)
	go SpawnPedestrians(ctx, 10, intersection.PedestrianBtn)

	fmt.Println(">>> Press [Enter] to stop <<<")

	_, _ = reader.ReadString('\n')

	fmt.Println("\nShut down..")
	cancel()

	time.Sleep(100 * time.Millisecond)
	fmt.Println("Shut down succeded")
}

func SpawnPedestrians(ctx context.Context, timePer int, pedChan chan<- struct{}) {
	go func() {
		for {
			select {
			case <-ctx.Done():
				fmt.Println("Pedestrian generator stopped.")
				return
			case <-time.After(time.Duration(rand.Intn(timePer)+20) * time.Second):
				fmt.Println("Pedestrian approached.")
				pedChan <- struct{}{}
			}
		}
	}()
}

func SpawnTraffic(ctx context.Context, timePer int, dir models.Direction, carChan chan<- models.CarEvent, cnt int) {
	buff := 0
	go func() {
		for {
			select {
			case <-ctx.Done():
				fmt.Printf("Generator %d stopped.\n", dir)
				return

			case <-time.After(time.Duration(rand.Intn(timePer)+5) * time.Second):
				if buff == cnt {
					time.Sleep(60 * time.Second)
				}
				buff++
				var chosenLane models.Lane
				if rand.Float32() > 0.75 {
					chosenLane = models.Left
				} else {
					chosenLane = models.Straight
				}

				carChan <- models.CarEvent{
					Dir:  dir,
					Lane: chosenLane,
				}
			}
		}
	}()
}
