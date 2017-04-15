package main

import (
	"fmt"
	"math/rand"
	"os"
	"time"
)

const (
	EnnemyFaction Faction = iota
	PlayerFaction
)

type Object interface {
	ID() int
}

type Faction int

func (f Faction) String() string {
	str := "ennemy"
	if f == PlayerFaction {
		str = "player"
	}
	return str
}

type Vec3 struct {
	x, y, z int
}

func (a Vec3) Dist(b Vec3) int {
	return (abs(a.x-b.x) + abs(a.y-b.y) + abs(a.z-b.z)) / 2
}

type GameObject struct {
	id   int
	x, y int
}

func (o GameObject) String() string {
	return fmt.Sprintf("{id: %d}", o.id)
}

func (a GameObject) ID() int { return a.id }

func (a GameObject) CubeCoord() Vec3 {
	x := a.x - (a.y-(a.y&1))/2
	z := a.y
	return Vec3{
		x: x,
		z: z,
		y: -x - z,
	}
}

func (a GameObject) Dist(b GameObject) int {
	return a.CubeCoord().Dist(b.CubeCoord())
}

type Barrel struct {
	GameObject
	Rhum int
}

type Ship struct {
	GameObject
	F        Faction
	Rhum     int
	Speed    int
	Rotation int
}

type Game struct {
	Objects map[int]Object
	Round   int
	Ships   []*Ship
	Barrels []*Barrel
}

func main() {
	random := rand.New(rand.NewSource(time.Now().UnixNano()))
	game := &Game{}
	for {
		// myShipCount: the number of remaining ships
		var myShipCount int
		fmt.Scan(&myShipCount)
		game.Ships = make([]*Ship, 0, myShipCount)

		// entityCount: the number of entities (e.g. ships, mines or cannonballs)
		var entityCount int
		fmt.Scan(&entityCount)
		game.Objects = make(map[int]Object)
		game.Barrels = nil

		for i := 0; i < entityCount; i++ {
			var entityId int
			var entityType string
			var x, y, arg1, arg2, arg3, arg4 int
			fmt.Scan(&entityId, &entityType, &x, &y, &arg1, &arg2, &arg3, &arg4)
			o := GameObject{
				id: entityId,
				x:  x,
				y:  y,
			}
			switch entityType {
			case "SHIP":
				s := &Ship{
					GameObject: o,
					Rotation:   arg1,
					Speed:      arg2,
					Rhum:       arg3,
					F:          Faction(arg4),
				}
				game.Ships = append(game.Ships, s)
				game.Objects[s.id] = s
			case "BARREL":
				b := &Barrel{
					GameObject: o,
					Rhum:       arg1,
				}
				game.Barrels = append(game.Barrels, b)
				game.Objects[b.id] = b
			default:
				fmt.Fprintln(os.Stderr, "unknown entity:", entityType)
				continue
			}
		}
		for _, s := range game.Ships {
			if s.F == EnnemyFaction {
				continue
			}

			// Find closest barrel.
			barrels := barrelToGOSlice(game.Barrels)
			sortGOSlice(barrels, func(i, j int) bool { return barrels[i].Dist(s.GameObject) < barrels[j].Dist(s.GameObject) })

			action := "WAIT"
			if len(barrels) > 0 {
				action = fmt.Sprintf("MOVE %d %d\n", barrels[0].x, barrels[0].y)
			} else if game.Round%5 == 0 {
				action = fmt.Sprintf("MOVE %d %d\n", random.Intn(22), random.Intn(23))
			}
			os.Stdout.Write([]byte(action))
		}
		game.Round++
	}
}

func abs(x int) int {
	// TODO: once golang.org/issue/13095 is fixed, change this to:
	// return Float64frombits(Float64bits(x) &^ (1 << 63))
	// But for now, this generates better code and can also be inlined:
	if x < 0 {
		return -x
	}
	if x == 0 {
		return 0 // return correctly abs(-0)
	}
	return x
}

func sortGOSlice(objects []*GameObject, less func(i, j int) bool) {
	for i := range objects {
		for j := range objects[i:] {
			if less(i, j) {
				objects[i], objects[j] = objects[j], objects[i]
			}
		}
	}
}

func barrelToGOSlice(barrels []*Barrel) []*GameObject {
	objects := make([]*GameObject, len(barrels))
	for i, b := range barrels {
		objects[i] = &b.GameObject
	}
	return objects
}
