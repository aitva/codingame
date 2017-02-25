package main

import (
	"fmt"
	"os"
)

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

type factory struct {
	ID      int
	Faction int
	Cyborg  int
	Prod    int
}

func (f *factory) findClosest(faction int) int {
	row := game.Board[f.ID]
	closest := -1
	for i := range row {
		if i == f.ID {
			continue
		}
		f := game.Factories[i]
		if f.Faction != faction {
			continue
		}
		fmt.Fprintln(os.Stderr, "closest:", closest)
		if !(closest == -1 || row[i] < row[closest]) {
			continue
		}
		closest = i
	}
	return closest
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
	Troops    map[int]*troop
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

	game.Troops = make(map[int]*troop)
	game.Factories = make(map[int]*factory)
	for {
		// entityCount: the number of entities (e.g. factories and troops)
		var entityCount int
		fmt.Scan(&entityCount)

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

		fmt.Fprintln(os.Stderr, game)

		action := "WAIT"
		for _, f := range game.Factories {
			if f.Faction != playerFaction {
				continue
			}
			id := f.findClosest(neutralFaction)
			if id == -1 {
				continue
			}
			dst := game.Factories[id]
			if dst.Cyborg > f.Cyborg {
				continue
			}
			action = fmt.Sprintf("MOVE %d %d %d", f.ID, dst.ID, dst.Cyborg+1)
		}

		// Any valid action, such as "WAIT" or "MOVE source destination cyborgs"
		fmt.Println(action)
	}
}
