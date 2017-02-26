package main

import "fmt"

//import "os"

/***
 * Auto-generated code below aims at helping you parse
 * the standard input according to the problem statement.
***/

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

func searchTroopDst(id, faction int) int {
	for _, t := range game.Troops {
		if t.Faction == faction && t.Dst == id {
			return t.ID
		}
	}
	return -1
}

func searchClosestFive(idx []int) []*factory {
	closest := make([]*factory, 0, 5)
	for _, i := range idx {
		tmp := game.Factories[i]
		if tmp.Faction != playerFaction && tmp.Prod > 0 {
			t := searchTroopDst(tmp.ID, playerFaction)
			if t != -1 {
				continue
			}
			closest = append(closest, tmp)
			break
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

		action := "WAIT"
		for _, f := range game.Factories {
			if f.Faction != playerFaction {
				continue
			}
			dist := game.getDistances(f.ID)
			// fmt.Fprintf(os.Stderr, "dist: %#v\n", dist)
			idx := sortDistIdx(dist)
			// fmt.Fprintf(os.Stderr, "idx: %#v\n", idx)
			idx = idx[1:]

			closest := searchClosestFive(idx)
			if closest == nil {
				continue
			}
			for _, c := range closest {
				// cyborg := f.Prod + 1
				turns := game.getDistance(f.ID, c.ID)
				cyborg := turns*c.Prod + c.Cyborg + 5
				if cyborg <= f.Cyborg {
					action += fmt.Sprint(";MOVE ", f.ID, c.ID, cyborg)
					f.Cyborg -= cyborg
				}
			}
		}

		// Any valid action, such as "WAIT" or "MOVE source destination cyborgs"
		fmt.Println(action)
	}
}
