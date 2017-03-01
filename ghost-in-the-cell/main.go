package main

import (
	"fmt"
	"os"
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
	Coef    int
}

func (f *factory) String() string {
	return fmt.Sprintf("{ID: %d, Cyborg: %d}", f.ID, f.Cyborg)
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
	Bomb  struct {
		Count int
		Timer int
	}
	FactoryCount int
	Factories    map[int]*factory
	TroopMaxID   int
	Troops       map[int]*troop
	Turn         int
	Path         []path
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

		// fmt.Fprintln(os.Stderr, "min:", min)
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

// searchTroopDst return the index of a faction shooting on id.
func searchTroopDst(id, faction int) int {
	for _, t := range game.Troops {
		if t.Faction == faction && t.Dst == id {
			return t.ID
		}
	}
	return -1
}

func searchTroop(src, dst int) int {
	s := game.Factories[src]
	d := game.Factories[dst]
	for _, t := range game.Troops {
		if t.Src == s.ID && t.Dst == d.ID {
			return t.ID
		}
	}
	return -1
}

// upateCyborgs compute number of cyborgs in all factories
// once all the troops reach destination.
func upateCyborgs() {
	for _, t := range game.Troops {
		f := game.Factories[t.Dst]
		if f.Faction == t.Faction {
			f.Cyborg += t.Cyborg
		} else if f.Faction != t.Faction {
			f.Cyborg -= t.Cyborg
		}
	}
}

func searchBestShots(src *factory) []*factory {
	// Get target factories.
	targets := make([]*factory, 0, game.FactoryCount)
	for _, f := range game.Factories {
		if f.Prod < 1 {
			continue
		}
		if f.Faction == opponentFaction && f.Cyborg < 0 {
			continue
		}
		if f.Faction == playerFaction && f.Cyborg >= 0 {
			continue
		}
		if f.Faction != playerFaction && f.Cyborg < 0 {
			continue
		}
		targets = append(targets, f)
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

func computeTroopSize(path []int, src *factory) int {
	cyborg := 0
	for _, id := range path {
		dst := game.Factories[id]
		// We don't add unit to take our own faction.
		if dst.Faction == playerFaction {
			continue
		}
		// Minimum amount for neutralFaction.
		tmp := dst.Cyborg + 1
		if tmp < 0 {
			// Might be negative if ennemy troops are moving.
			tmp = (tmp - 2) * -1
		}
		if dst.Faction == opponentFaction {
			// TODO: use dijkstra instead
			turns := game.getDistance(src.ID, dst.ID)
			tmp += turns * dst.Prod
		}
		if src.Cyborg-cyborg-tmp <= 0 {
			cyborg = src.Cyborg - 1
			break
		}
		cyborg += tmp
	}
	return cyborg
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
		// p := game.Path[i]
		// fmt.Fprintln(os.Stderr, "id:", i)
		// fmt.Fprintln(os.Stderr, "p.Dist:", dist)
		// fmt.Fprintln(os.Stderr, "p.Prev:", prev)
		// fmt.Fprintln(os.Stderr, "p.Closest:", p.Closest)
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

		action := ""
		if game.Bomb.Timer <= 0 && game.Bomb.Count > 0 {
			target := -1
			for id, f := range game.Factories {
				found := target == -1
				found = found || f.Prod > game.Factories[target].Prod
				found = found && f.Faction == opponentFaction
				if found {
					target = id
				}
			}
			row := game.Board[target]
			factory := -1
			for i := range row {
				found := factory == -1
				found = found || row[i] < row[factory]
				found = found && game.Factories[i].Faction == playerFaction
				if found {
					factory = i
				}
			}
			game.Bomb.Timer = bombTime + row[factory]
			action = fmt.Sprintf("BOMB %d %d; ", factory, target)
			game.Bomb.Count--
		}
		for _, f := range game.Factories {
			if f.Faction != playerFaction {
				continue
			}
			upateCyborgs() // To improve shot.
			if f.Cyborg <= 0 {
				continue
			}
			if f.Cyborg > 15 && f.Prod >= 1 && f.Prod < 3 {
				action += fmt.Sprintf("INC %d; ", f.ID)
				continue
			}

			// Choose an action.
			targets := searchBestShots(f)
			fmt.Fprintln(os.Stderr, "f:", f, "targets:", targets)
			for _, t := range targets {
				if t.ID == f.ID {
					continue
				}
				// fmt.Fprintln(os.Stderr, "prev:", game.Path[f.ID].Prev, "t.ID:", t.ID)
				path := pathToDst(game.Path[f.ID].Prev, t.ID)
				fmt.Fprintln(os.Stderr, "path:", path)
				// cyborg := computeTroopSize(path, f)
				cyborg := f.Cyborg
				fmt.Fprintf(os.Stderr, "f: %v; t: %v; cyborg: %d\n", f, t, cyborg)
				if cyborg == 0 {
					continue
				}
				action += fmt.Sprintf("MOVE %d %d %d; ", f.ID, path[0], cyborg)
				game.TroopMaxID++
				game.Troops[game.TroopMaxID] = &troop{
					ID:      game.TroopMaxID,
					Faction: playerFaction,
					Cyborg:  cyborg,
					Src:     f.ID,
					Dst:     t.ID,
					Turns:   game.Path[f.ID].Dist[path[0]],
				}
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
	}
}
