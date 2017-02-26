package main

import "fmt"
import "os"

//import "os"

/***
 * Auto-generated code below aims at helping you parse
 * the standard input according to the problem statement.
***/

const (
	opponentFaction = -1
	neutralFaction  = 0
	playerFaction   = 1
	maxDistance     = 20
)

var game Game

type factoriesByDist struct {
	Dist []int
	IDs  []int
}

func newFactoriesByDist(dist []int) factoriesByDist {
	IDs := make([]int, len(dist))
	for i := range IDs {
		IDs[i] = i
	}
	return factoriesByDist{
		Dist: dist,
		IDs:  IDs,
	}
}
func (a factoriesByDist) Len() int { return len(a.Dist) }
func (a factoriesByDist) Swap(i, j int) {
	a.IDs[i], a.IDs[j] = a.IDs[j], a.IDs[i]
	a.Dist[i], a.Dist[j] = a.Dist[j], a.Dist[i]
}
func (a factoriesByDist) Less(i, j int) bool { return a.Dist[i] < a.Dist[j] }

type factoryList []*factory

func (f factoryList) filterFaction(faction int) factoryList {
	l := make(factoryList, 0, len(f))
	for i := range f {
		if f[i].Faction == faction {
			l = append(l, f[i])
		}
	}
	return l
}

func (f factoryList) filterTroops(troops troopMap) factoryList {
	l := make(factoryList, 0, len(f))
main_loop:
	for i := range f {
		for _, t := range troops {
			if t.Dst == f[i].ID {
				continue main_loop
			}
		}
		l = append(l, f[i])
	}
	return l
}

type troopMap map[int]*troop

func (troops troopMap) filterFaction(faction int) troopMap {
	m := make(troopMap)
	for _, t := range troops {
		if t.Faction == faction {
			m[t.ID] = t
		}
	}
	return m
}

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
	Board     [][]int
	Factories map[int]*factory
	Troops    troopMap
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

func (g *Game) getDistance(src, dst int) int {
	return g.Board[src][dst]
}

// orderFactories get all factories ordered by distance.
func (g *Game) orderFactories(id int) factoryList {
	row := g.Board[id]
	sorted := newFactoriesByDist(row)

	f := make([]*factory, 0, len(sorted.IDs))
	for _, i := range sorted.IDs {
		if i != id {
			f = append(f, g.Factories[i])
		}
	}

	return f
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
		meTroops := game.Troops.filterFaction(playerFaction)
		for _, f := range game.Factories {
			if f.Faction != playerFaction {
				continue
			}
			list := game.orderFactories(f.ID)
			list = list[1:]
			opp := list.filterFaction(opponentFaction)
			neutral := list.filterFaction(neutralFaction)
			all := append(neutral, opp...)
			//fmt.Fprintln(os.Stderr, "list:", all)
			all = all.filterTroops(meTroops)
			fmt.Fprintln(os.Stderr, "list:", all)
			if len(all) == 0 {
				continue
			}
			closest := all[0]
			// cyborg := game.getDistance(f.ID, closest.ID)*closest.Prod + closest.Cyborg
			// fmt.Fprintln(os.Stderr, "cyborg:", cyborg)
			// if cyborg < 20 && f.Cyborg <= cyborg {
			// 	continue
			// }
			action = fmt.Sprintf("MOVE %d %d %d", f.ID, closest.ID, f.Cyborg/2)
		}

		// Any valid action, such as "WAIT" or "MOVE source destination cyborgs"
		fmt.Println(action)
	}
}
