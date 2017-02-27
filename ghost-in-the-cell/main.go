package main

import "fmt"
import "os"

const (
	opponentFaction  = -1
	neutralFaction   = 0
	playerFaction    = 1
	closestNeighboor = 4
)

var game Game

type factory struct {
	ID      int
	Faction int
	Cyborg  int
	Prod    int
	Action  string
	Coef    int
}

func (f *factory) String() string {
	return fmt.Sprintf("{ID: %d, Coef: %d}", f.ID, f.Coef)
}

type troop struct {
	ID      int
	Faction int
	Cyborg  int
	Src     int
	Dst     int
	Turns   int
}

type Game struct {
	Board      [][]int
	Factories  map[int]*factory
	TroopMaxID int
	Troops     map[int]*troop
}

func (g *Game) String() string {
	str := ""
	for _, row := range g.Board {
		for _, f := range row {
			str += fmt.Sprintf("%2d ", f)
		}
		str += "\n"
	}
	return str
}

func (g *Game) getDistance(srcID, dstID int) int {
	return g.Board[srcID][dstID]
}

func (g *Game) getDistances(ID int) []int {
	return g.Board[ID]
}

func sortDistIdx(dist []int) []int {
	tmp := make([]int, len(dist))
	copy(tmp, dist)
	idx := make([]int, len(tmp))
	for i := range idx {
		iMin := 0
		for j := range tmp {
			if tmp[iMin] == -1 {
				iMin = j
				continue
			}
			if tmp[j] >= 0 && tmp[j] < tmp[iMin] {
				iMin = j
			}
		}
		tmp[iMin] = -1
		idx[i] = iMin
	}
	return idx
}

// searchTroopDst return the index of a faction shooting on id.
func searchTroopDst(id, faction int) int {
	for _, t := range game.Troops {
		if t.Faction == faction && t.Dst == id {
			return t.ID
		}
	}
	return -1
}

func computeActualCyborg(id int) int {
	cyborg := game.Factories[id].Cyborg
	for _, t := range game.Troops {
		if t.Dst != id {
			continue
		}
		if t.Faction == opponentFaction {
			cyborg -= t.Cyborg
		} else if t.Faction == playerFaction {
			cyborg += t.Cyborg
		}
	}
	return cyborg
}

// searchFiveOpponent look for 5 factories (neutral or opponent).
// It is filtering on positive prod & no bro shooting.
func searchFiveOpponent(idx []int) []*factory {
	const max = 5
	closest := make([]*factory, 0, max)
	for i := 0; i < len(idx) && len(closest) < 5; i++ {
		id := idx[i]
		tmp := game.Factories[id]
		if tmp.Faction != playerFaction && tmp.Prod > 0 {
			t := searchTroopDst(tmp.ID, playerFaction)
			if t != -1 {
				continue
			}
			closest = append(closest, tmp)
		}
	}
	return closest
}

func computeTroopSize(src, dst *factory) int {
	turns := game.getDistance(src.ID, dst.ID)
	if dst.Faction == neutralFaction {
		turns = 0
	}
	return turns*dst.Prod + dst.Cyborg + 3
}

func main() {
	// factoryCount: the number of factories
	var factoryCount int
	fmt.Scan(&factoryCount)
	game.Board = make([][]int, factoryCount)
	for i := range game.Board {
		game.Board[i] = make([]int, factoryCount)
	}

	// linkCount: the number of links between factories
	var linkCount int
	fmt.Scan(&linkCount)

	for i := 0; i < linkCount; i++ {
		var factory1, factory2, distance int
		fmt.Scan(&factory1, &factory2, &distance)
		game.Board[factory1][factory2] = distance
		game.Board[factory2][factory1] = distance
	}

	for {
		// entityCount: the number of entities (e.g. factories and troops)
		var entityCount int
		fmt.Scan(&entityCount)

		game.Troops = make(map[int]*troop)
		game.Factories = make(map[int]*factory)
		for i := 0; i < entityCount; i++ {
			var entityId int
			var entityType string
			var arg1, arg2, arg3, arg4, arg5 int
			fmt.Scan(&entityId, &entityType, &arg1, &arg2, &arg3, &arg4, &arg5)
			if entityType == "TROOP" {
				t := &troop{
					ID:      entityId,
					Faction: arg1,
					Src:     arg2,
					Dst:     arg3,
					Cyborg:  arg4,
					Turns:   arg5,
				}
				game.Troops[t.ID] = t
				game.TroopMaxID = t.ID
			} else if entityType == "FACTORY" {
				f := &factory{
					ID:      entityId,
					Faction: arg1,
					Cyborg:  arg2,
					Prod:    arg3,
				}
				game.Factories[f.ID] = f
			}
		}
		//fmt.Fprintln(os.Stderr, &game)

		// Compute factories coeficients.
		for _, f := range game.Factories {
			if f.Faction == opponentFaction {
				f.Coef = -1
			} else if f.Faction == playerFaction {
				f.Coef = +1
			} else {
				continue
			}

			dist := game.getDistances(f.ID)
			idx := sortDistIdx(dist)
			for _, id := range idx[1 : closestNeighboor+1] {
				closest := game.Factories[id]
				closest.Coef += f.Coef
			}
		}

		action := ""
		for _, f := range game.Factories {
			if f.Faction != playerFaction {
				continue
			}
			dist := game.getDistances(f.ID)
			idx := sortDistIdx(dist)

			// Get target factories.
			targets := make([]*factory, closestNeighboor)
			for i, id := range idx[1 : len(targets)+1] {
				targets[i] = game.Factories[id]
			}

			// Order by coef.
			for i := range targets {
				for j := range targets {
					if targets[i].Coef < targets[j].Coef {
						targets[j], targets[i] = targets[i], targets[j]
					}
				}
			}
			fmt.Fprintln(os.Stderr, targets)

			// Choose an action.
			for _, t := range targets {
				if t.Faction == playerFaction {
					action += fmt.Sprintf("MOVE %d %d %d; ", f.ID, t.ID, f.Prod)
					f.Cyborg -= f.Prod
					break
				}
				cyborg := computeTroopSize(f, t)
				if f.Cyborg < cyborg {
					continue
				}
				action += fmt.Sprintf("MOVE %d %d %d; ", f.ID, t.ID, cyborg)
				f.Cyborg -= cyborg
			}
		}
		if action != "" {
			action = action[:len(action)-2]
		} else {
			action = "WAIT"
		}

		// Any valid action, such as "WAIT" or "MOVE source destination cyborgs"
		fmt.Println(action)
	}
}
