package main

import "fmt"

const (
	opponentFaction = -1
	neutralFaction  = 0
	playerFaction   = 1
)

var game Game

type factory struct {
	ID      int
	Faction int
	Cyborg  int
	Prod    int
	Action  string
	Damage  int
}

func (f *factory) String() string {
	return fmt.Sprintf("{ID: %d, Fac: %d}", f.ID, f.Faction)
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

// searchTopFive look for top 5 factories.
// If opponent or neutral: filter on positive prod & no bro shooting.
// If bro filter on ennemy shooting.
func searchTopFive(idx []int) []*factory {
	const max = 5
	closest := make([]*factory, 0, max)
	for i := 0; i < len(idx) && len(closest) < 5; i++ {
		id := idx[i]
		tmp := game.Factories[id]
		faction := playerFaction
		if tmp.Prod < 1 {
			continue
		}
		if tmp.Faction == playerFaction {
			faction = opponentFaction
		}
		t := searchTroopDst(tmp.ID, faction)
		if t != -1 {
			continue
		}
		closest = append(closest, tmp)
	}
	return closest
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

		action := ""
		for _, f := range game.Factories {
			if f.Faction != playerFaction {
				continue
			}
			dist := game.getDistances(f.ID)
			// fmt.Fprintf(os.Stderr, "dist: %#v\n", dist)
			idx := sortDistIdx(dist)
			// fmt.Fprintf(os.Stderr, "idx: %#v\n", idx)
			idx = idx[1:]

			// fmt.Fprintf(os.Stderr, "player: %v\n", f)
			top := searchFiveOpponent(idx)
			// fmt.Fprintf(os.Stderr, "opponents: %v\n", top)
			if top == nil {
				continue
			}
			for _, o := range top {
				turns := game.getDistance(f.ID, o.ID)
				if o.Faction == neutralFaction {
					turns = 0
				}
				cyborg := turns*o.Prod + o.Cyborg + 1
				actualCyborg := computeActualCyborg(f.ID)
				// fmt.Fprintf(os.Stderr, "cyborg(%d) <= f.Cyborg(%d)\n", cyborg, f.Cyborg)
				if cyborg <= actualCyborg {
					action += fmt.Sprint("MOVE ", f.ID, o.ID, cyborg, ";")
					f.Cyborg -= cyborg
					// Avoid duplicate troops.
					game.TroopMaxID++
					game.Troops[game.TroopMaxID] = &troop{
						ID:      game.TroopMaxID,
						Faction: playerFaction,
						Src:     f.ID,
						Dst:     o.ID,
						Cyborg:  cyborg,
						Turns:   turns,
					}
				}
			}
		}
		if action != "" {
			action = action[:len(action)-1]
		} else {
			action = "WAIT"
		}

		// Any valid action, such as "WAIT" or "MOVE source destination cyborgs"
		fmt.Println(action)
	}
}
