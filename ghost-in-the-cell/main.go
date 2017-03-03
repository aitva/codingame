package main

import (
	"fmt"
	"os"
	"time"
)

const (
	opponentFaction  = -1
	neutralFaction   = 0
	playerFaction    = 1
	closestNeighboor = 5
	maxDistance      = 21
	invalidPath      = -1
	bombTime         = 5
)

var game Game

type factory struct {
	ID      int
	Faction int
	Cyborg  int
	Prod    int
	Troops  struct {
		Player   int
		Opponent int
	}
}

func (f *factory) String() string {
	return fmt.Sprintf("{ID: %d, Fa: %d, Cy: %d}", f.ID, f.Faction, f.Cyborg)
}

func (f *factory) EstimatedCyborg() int {
	cyborg := f.Cyborg - f.Troops.Player + f.Troops.Opponent
	if f.Faction == playerFaction {
		cyborg = f.Cyborg + f.Troops.Player - f.Troops.Opponent
	}
	return cyborg
}

type troop struct {
	ID      int
	Faction int
	Cyborg  int
	Src     int
	Dst     int
	Turns   int
}

type path struct {
	Dist    []int // distance to every factory
	Prev    []int // previous factory on the path
	Closest []int // factories ordered by distance
}

type Game struct {
	Board [][]int

	FactoryCount int
	Factories    map[int]*factory
	NeutralF     []*factory
	OpponentF    []*factory
	PlayerF      []*factory

	TroopMaxID int
	Troops     map[int]*troop

	Bomb struct {
		Count int
		Timer int
	}

	Turn int
	Path []path
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

func new2DSlice(n, m int) [][]int {
	tmp := make([]int, n*m)
	slice := make([][]int, n)
	for i := range slice {
		slice[i] = tmp[i*m : (i+1)*m]
	}
	return slice
}

func dijkstra(src int) (dist, prev []int) {
	unvisited := make(map[int]struct{})
	dist = make([]int, game.FactoryCount)
	prev = make([]int, game.FactoryCount)
	for i := 0; i < game.FactoryCount; i++ {
		dist[i] = maxDistance
		prev[i] = invalidPath
		unvisited[i] = struct{}{}
	}
	dist[src] = 0
	for len(unvisited) > 0 {
		min := -1
		for i := range unvisited {
			if min == -1 || dist[i] < dist[min] {
				min = i
			}
		}
		delete(unvisited, min)

		for v := range game.Board[min] {
			alt := dist[min] + game.Board[min][v]
			if alt < dist[v] {
				dist[v] = alt
				prev[v] = min
			}
		}
	}
	return
}

func pathToDst(prev []int, dst int) []int {
	path := make([]int, len(prev))
	i := dst
	j := len(prev) - 1
	for prev[i] != invalidPath {
		path[j] = i
		i = prev[i]
		j--
	}
	return path[j+1:]
}

func sortIndex(dist []int) []int {
	ids := make([]int, len(dist))
	for i := range ids {
		ids[i] = i
	}
	for i := range ids {
		for j := range dist {
			if dist[ids[j]] > dist[ids[i]] {
				ids[i], ids[j] = ids[j], ids[i]
			}
		}
	}
	return ids
}

// upateTroops compute number of cyborgs in all factories
// once all the troops reach destination.
func upateTroops() {
	for _, t := range game.Troops {
		f := game.Factories[t.Dst]
		if t.Faction == opponentFaction {
			f.Troops.Opponent += t.Cyborg
		} else if t.Faction == playerFaction {
			f.Troops.Player += t.Cyborg
		}
	}
}

func searchBestShots(src *factory) []*factory {
	// Get target factories.
	targets := make([]*factory, 0, game.FactoryCount)
	for _, f := range game.NeutralF {
		if f.Prod < 1 || f.Cyborg-f.Troops.Player < 0 {
			continue
		}
		targets = append(targets, f)
	}

	for _, f := range game.PlayerF {
		if f.Prod < 1 || f.EstimatedCyborg() >= 0 {
			continue
		}
		targets = append(targets, f)
	}

	if len(targets) == 0 {
		targets = append(targets, game.OpponentF...)
	}

	// Order by faction, prod, dist.
	dist := game.Path[src.ID].Dist
	for i := range targets {
		for j := range targets[i:] {
			swap := false
			swap = swap || dist[targets[i].ID] < dist[targets[j].ID]
			swap = swap || targets[i].Prod > targets[j].Prod
			//swap = swap || targets[i].Faction == opponentFaction
			if swap {
				targets[j], targets[i] = targets[i], targets[j]
			}
		}
	}

	return targets
}

func main() {
	// factoryCount: the number of factories
	var factoryCount int
	fmt.Scan(&factoryCount)
	game.FactoryCount = factoryCount
	game.Board = new2DSlice(factoryCount, factoryCount)
	game.Bomb.Count = 2

	// linkCount: the number of links between factories
	var linkCount int
	fmt.Scan(&linkCount)

	for i := 0; i < linkCount; i++ {
		var factory1, factory2, distance int
		fmt.Scan(&factory1, &factory2, &distance)
		game.Board[factory1][factory2] = distance
		game.Board[factory2][factory1] = distance
	}
	fmt.Fprintln(os.Stderr, &game)

	game.Path = make([]path, factoryCount)
	for i := range game.Path {
		dist, prev := dijkstra(i)
		game.Path[i].Dist = dist
		game.Path[i].Prev = prev
		game.Path[i].Closest = sortIndex(dist)
	}

	for {
		tstart := time.Now()
		// entityCount: the number of entities (e.g. factories and troops)
		var entityCount int
		fmt.Scan(&entityCount)

		game.Troops = make(map[int]*troop)
		game.Factories = make(map[int]*factory)
		game.NeutralF = make([]*factory, 0, game.FactoryCount)
		game.PlayerF = make([]*factory, 0, game.FactoryCount)
		game.OpponentF = make([]*factory, 0, game.FactoryCount)
		for i := 0; i < entityCount; i++ {
			var entityID int
			var entityType string
			var arg1, arg2, arg3, arg4, arg5 int
			fmt.Scan(&entityID, &entityType, &arg1, &arg2, &arg3, &arg4, &arg5)
			if entityType == "TROOP" {
				t := &troop{
					ID:      entityID,
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
					ID:      entityID,
					Faction: arg1,
					Cyborg:  arg2,
					Prod:    arg3,
				}
				game.Factories[f.ID] = f
				if f.Faction == neutralFaction && f.Prod > 0 {
					game.NeutralF = append(game.NeutralF, f)
				} else if f.Faction == playerFaction {
					game.PlayerF = append(game.PlayerF, f)
				} else if f.Faction == opponentFaction {
					game.OpponentF = append(game.OpponentF, f)
				}
			}
		}
		upateTroops()

		action := ""
		// Throw bomb one at a time.
		if game.Bomb.Timer <= 0 && game.Bomb.Count > 0 {
			var target *factory
			for _, f := range game.OpponentF {
				if target == nil || f.Prod > target.Prod {
					target = f
				}
			}
			row := game.Board[target.ID]
			var src *factory
			for id, f := range game.PlayerF {
				if src == nil || row[id] < row[src.ID] {
					src = f
				}
			}
			game.Bomb.Timer = bombTime + row[src.ID]
			action = fmt.Sprintf("BOMB %d %d; ", src.ID, target.ID)
			game.Bomb.Count--
		}
		for _, f := range game.Factories {
			if f.Faction != playerFaction {
				continue
			}
			if f.EstimatedCyborg() <= 0 {
				continue
			}
			if len(game.NeutralF) == 0 && f.Troops.Opponent == 0 && f.Cyborg > 15 && f.Prod >= 1 && f.Prod < 3 {
				action += fmt.Sprintf("INC %d; ", f.ID)
				continue
			}

			// Choose an action.
			targets := searchBestShots(f)
			fmt.Fprintln(os.Stderr, "f:", f, "targets:", targets)
			fmt.Fprintln(os.Stderr, "dist:", game.Path[f.ID].Dist)
			for _, t := range targets {
				if t.ID == f.ID {
					continue
				}
				// fmt.Fprintln(os.Stderr, "prev:", game.Path[f.ID].Prev, "t.ID:", t.ID)
				path := pathToDst(game.Path[f.ID].Prev, t.ID)
				// fmt.Fprintln(os.Stderr, "path:", path)
				// cyborg := computeTroopSize(path, f)
				cyborg := f.Cyborg
				fmt.Fprintf(os.Stderr, "- t: %v; cyborg: %d\n", t, cyborg)
				if cyborg == 0 {
					continue
				}
				if t.Faction == opponentFaction {
					action += fmt.Sprintf("MSG Attak!; ")
				}
				action += fmt.Sprintf("MOVE %d %d %d; ", f.ID, path[0], cyborg)
				// Improve shot.
				game.Factories[path[0]].Troops.Player += cyborg
				f.Cyborg -= cyborg
				break
			}
		}
		if action != "" {
			action = action[:len(action)-2]
		} else {
			action = "WAIT"
		}

		// Any valid action, such as "WAIT" or "MOVE source destination cyborgs"
		fmt.Println(action)
		game.Turn++
		game.Bomb.Timer--
		fmt.Fprintln(os.Stderr, "time:", time.Since(tstart))
	}
}
