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
	6, 6, 6, 6, 6, 6, 6, 6}

type StateConfig struct {
	minDur, maxDur time.Duration
	curDur         time.Duration

	nextState State
}

type TrafficMachine struct {
	curState State

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
		stateCfg:     buff,
		initialDur:   initDur,
		pedTime:      pedTime,
		mode:         m,
		intersection: inter,
	}
}
func (t *TrafficMachine) Run(ctx context.Context) {
	fmt.Println("🚦 Traffic light raised. Mode:", t.mode)
	// to break init state properly
	select {
	case <-ctx.Done():
		return
	case <-time.After(t.initialDur):
		t.curState = 0
	}
	for {
		if t.mode == ADAPTIVE {
			if t.intersection.GetAllSensorsCnt() == 0 {
				fmt.Printf("\n[%s] 🚶 No cars detected. All directions RED. Pedestrians GREEN.\n", time.Now().Format("15:04:05"))
				t.intersection.GetPedStatus()
			IdleLoop:
				for {
					select {
					case <-ctx.Done():
						fmt.Println("🛑 Traffic light shut down")
						return
					case <-time.After(1 * time.Second):
						if t.intersection.GetAllSensorsCnt() > 0 {
							fmt.Printf("\n[%s] 🚗 Car arrived! Waking up traffic cycle...\n", time.Now().Format("15:04:05"))

							t.curState = 0
							break IdleLoop
						}
						t.intersection.GetPedStatus()
					}
				}
			}

			skippedCount := 0
			for t.curState%2 == 0 && t.getStateAwaitingCarCnt(t.curState) == 0 {
				if skippedCount >= int(totalStates)/2 {
					break
				}
				fmt.Printf("\n⏭️  Skipping empty phase: %v\n", t.curState)
				t.curState = (t.curState + 2) % totalStates
				skippedCount++

				if t.curState == 0 {
					t.recalculateTimings()
				}
			}
		}

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
				fmt.Println("🛑 Traffic light shut down")
				return

			case <-tick5.C:
				rem := time.Until(endTime).Round(time.Second)
				if rem > 0 {
					fmt.Printf("\t Phase time left: %v\n", rem)
				}

			case <-phaseTimer.C:
				tick5.Stop()
				// Immitation
				if t.curState%2 == 0 {
					carsCanPass := int(cfg.curDur.Seconds()) / 2
					sensorIDs := t.getStateSensors(t.curState)
					passed := t.intersection.(*models.IntersectionUno).DequeueCars(sensorIDs, carsCanPass)
					if passed > 0 {
						fmt.Printf("🚗 %d cars passed during green light.\n", passed)
					}
				}
				if t.mode == ADAPTIVE {
					if t.intersection.GetPedStatus() {
						t.handlePedestrian(ctx, cfg.nextState)
					}
				}

				t.curState = cfg.nextState

				if t.curState == 0 && t.mode == ADAPTIVE {
					t.recalculateTimings()
				}

				break PhaseLoop
			}
		}
	}
}

func (t *TrafficMachine) handlePedestrian(ctx context.Context, state State) {
	if t.getStateAwaitingCarCnt(state) < minCarAwait || state == totalStates-1 {
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

func (t *TrafficMachine) getStateSensors(state State) []int {
	buff := state % 4
	switch buff {
	case 0:
		// Прямо и направо
		return []int{int(state), int(state) + 2}
	case 2:
		// Налево
		return []int{
			int(state) - 1,
			int(state) + 1,
			(int(state) + 2) % int(totalStates),
			(int(state) + 4) % int(totalStates),
		}
	default:
		return []int{}
	}
}

func (t *TrafficMachine) getStateAwaitingCarCnt(state State) int {
	sensors := t.getStateSensors(state)
	total := 0
	for _, id := range sensors {
		total += t.intersection.GetSensorCnt(id)
	}
	return total
}

func (t *TrafficMachine) recalculateTimings() {
	fmt.Println("\n🔄 [CYCLE END] Recalculating timings...")

	totalWaiting := t.intersection.(*models.IntersectionUno).GetAllSensorsCnt()

	if totalWaiting == 0 {
		fmt.Println("No cars waiting. Returning to minimum durations.")
		for i := 0; i < len(t.stateCfg); i++ {
			t.stateCfg[i].curDur = t.stateCfg[i].minDur
		}
		return
	}

	for i := State(0); i < totalStates; i += 2 {
		waitingForPhase := t.getStateAwaitingCarCnt(i)
		cfg := &t.stateCfg[i]

		if waitingForPhase == 0 {
			cfg.curDur = cfg.minDur
			continue
		}

		newSeconds := int(cfg.minDur.Seconds()) + (waitingForPhase * 2)
		newDur := time.Duration(newSeconds) * time.Second

		if newDur > cfg.maxDur {
			newDur = cfg.maxDur
		}

		if cfg.curDur != newDur {
			fmt.Printf("⏱️ Phase %d duration changed: %v -> %v (Cars waiting: %d)\n", i, cfg.curDur, newDur, waitingForPhase)
			cfg.curDur = newDur
		}
	}
}
