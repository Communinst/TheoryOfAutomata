package fsm

import (
	"context"
	"fmt"
	"time"

	"github.com/Communinst/TheoryOfAutomata/internal/models"
)

type Mode = int

const (
	FIXED Mode = iota
	ADAPTIVE
)

const (
	minCarAwait int = 10
)

var cycle = [totalStates]State{
	NSG_STRAIGHT_RIGHT, NSY_STRAIGHT_RIGHT,
	NSG_LEFT, NSY_LEFT,
	EWG_STRAIGHT_RIGHT, EWY_STRAIGHT_RIGHT,
	EWG_LEFT, EWY_LEFT}

//	var timer = [totalStates]time.Duration{
//		30, 5, 20, 5, 30, 5, 20, 5}
var timer = [totalStates]time.Duration{
	5, 5, 5, 5, 5, 5, 5, 5}

type StateConfig struct {
	minDur, maxDur time.Duration
	curDur         time.Duration

	nextState State
}

type TrafficMachine struct {
	curState State
	started  time.Time

	stateCfg   [totalStates]StateConfig
	initialDur time.Duration

	pedTime time.Duration

	mode         Mode
	intersection models.IntersectionInterface
}

func NewTrafficMachine(init State, initDur, pedTime time.Duration, m Mode, inter models.IntersectionInterface) *TrafficMachine {
	var buff [totalStates]StateConfig

	for i := 0; i < len(cycle)-1; i++ {
		val := timer[i] * time.Second
		buff[i] = StateConfig{
			minDur:    val,
			maxDur:    val * 2,
			curDur:    val,
			nextState: cycle[i+1],
		}
	}
	val := timer[len(cycle)-1]
	buff[len(cycle)-1] = StateConfig{
		minDur:    val,
		maxDur:    val * 2,
		curDur:    val,
		nextState: cycle[0],
	}

	return &TrafficMachine{
		curState:     init,
		started:      time.Now(),
		stateCfg:     buff,
		initialDur:   initDur,
		pedTime:      pedTime,
		mode:         m,
		intersection: inter,
	}
}

func (t *TrafficMachine) Run(ctx context.Context) {
	fmt.Println("Traffic light raised.")
	select {
	case <-ctx.Done():
		return
	case <-time.After(t.initialDur):
		t.curState = 0
		t.started = time.Now()
	}
	for {
		fmt.Printf("\n[%s] Transition:\n%s", time.Now().Format("15:04:05"), t.curState)

		cfg := t.stateCfg[t.curState]
		endTime := time.Now().Add(cfg.curDur)
		phaseTimer := time.NewTimer(cfg.curDur)
		tick5 := time.NewTicker(5 * time.Second)

	PhaseLoop:
		for {
			select {
			case <-ctx.Done():
				phaseTimer.Stop()
				tick5.Stop()
				fmt.Println("Traffic light shut down")
				return

			case <-tick5.C:
				rem := time.Until(endTime).Round(time.Second)

				if rem > 0 {
					fmt.Printf("\t Phase time left: %v\n", rem)
				}

			case <-phaseTimer.C:
				tick5.Stop()

				if t.intersection.GetPedStatus() {
					t.handlePedestrian(ctx, cfg.nextState)
				}

				t.curState = cfg.nextState
				t.started = time.Now()

				break PhaseLoop
			}
		}
	}
}

func (t *TrafficMachine) handlePedestrian(ctx context.Context, state State) {
	if t.getStateAwaitingCarCnt(state) < minCarAwait {
		endTime := time.Now().Add(t.pedTime)
		phaseTimer := time.NewTimer(t.pedTime)
		tick5 := time.NewTicker(5 * time.Second)
		fmt.Println("Red in all directions. Pedestrian green")
	PhaseLoop:
		for {
			select {
			case <-ctx.Done():
				phaseTimer.Stop()
				tick5.Stop()
				fmt.Println("Traffic light shut down")
				return

			case <-tick5.C:
				rem := time.Until(endTime).Round(time.Second)

				if rem > 0 {
					fmt.Printf("\t Phase time left: %v\n", rem)
				}

			case <-phaseTimer.C:
				tick5.Stop()

				break PhaseLoop
			}
		}

	}
}

func (t *TrafficMachine) getStateAwaitingCarCnt(state State) int {
	buff := state % 4
	switch buff {
	case 0:
		return t.intersection.GetSensorCnt(int(state)) + t.intersection.GetSensorCnt(int(state)+2)
	case 2:
		return t.intersection.GetSensorCnt(int(state)-1) +
			t.intersection.GetSensorCnt(int(state)+1) +
			t.intersection.GetSensorCnt((int(state)+2)%int(totalStates)) +
			t.intersection.GetSensorCnt((int(state)+4)%int(totalStates))
	default:
		return 0
	}
}
