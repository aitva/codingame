package main

import (
	"fmt"
	"math"
	"sort"
)

const (
	MaxSnaffles      = 7
	WizardsPerPlayer = 2
	RadiusWizard     = 400
	RadiusSnaffle    = 150
	RadiusBludger    = 200
)

var goals = struct {
	scoreLeft bool
	mine      Point
	theirs    Point
}{
	mine:   Point{x: 0, y: 3750},
	theirs: Point{x: 16000, y: 3750},
}

type Object interface {
	ID() int
	Update(x, y, vx, vy, state int)
	Pos() Point
	Radius() int
}

func ComputeDistance(o Object, objs []Object) []float64 {
	dist := make([]float64, len(objs))
	a := o.Pos()
	for i, oo := range objs {
		b := oo.Pos()
		x := float64((b.x - a.x) * (b.x - a.x))
		y := float64((b.y - a.y) * (b.y - a.y))
		dist[i] = math.Sqrt(x + y)
	}
	return dist
}

func RemoveFromSlice(o Object, src []Object) []Object {
	dst := make([]Object, 0, len(src))
	for _, oo := range src {
		if o.ID() == oo.ID() {
			continue
		}
		dst = append(dst, oo)
	}
	return dst
}

type ByX []Object

func (a ByX) Len() int           { return len(a) }
func (a ByX) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ByX) Less(i, j int) bool { return a[i].Pos().x < a[j].Pos().x }

type ByXDesc []Object

func (a ByXDesc) Len() int           { return len(a) }
func (a ByXDesc) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ByXDesc) Less(i, j int) bool { return a[i].Pos().x > a[j].Pos().x }

type Point struct {
	x int
	y int
}

type Vector struct {
	x int
	y int
}

type GameObject struct {
	id     int
	radius int
	pos    Point
	v      Vector
	state  int
}

func NewGameObject(id, radius int) *GameObject {
	return &GameObject{
		id:     id,
		radius: radius,
	}
}

func (obj *GameObject) ID() int {
	return obj.id
}

func (obj *GameObject) Update(x, y, vx, vy, state int) {
	obj.pos.x = x
	obj.pos.y = y
	obj.v.x = vx
	obj.v.y = vy
	obj.state = state
}

func (obj *GameObject) Pos() Point {
	return obj.pos
}

func (obj *GameObject) Radius() int {
	return obj.radius
}

type Wizard struct {
	*GameObject
	hasSnaffle bool
	target     Object
}

func NewWizard(id int) *Wizard {
	return &Wizard{
		GameObject: NewGameObject(id, RadiusWizard),
	}
}

func (w *Wizard) Update(x, y, vx, vy, state int) {
	w.GameObject.Update(x, y, vx, vy, state)
	w.hasSnaffle = false
	if state == 1 {
		w.hasSnaffle = true
	}
}

// Defend built a defensive action for the Wizard.
func (w *Wizard) Defend(snaffles []Object) string {
	w.target = nil
	if w.hasSnaffle {
		p := goals.theirs
		return fmt.Sprintf("THROW %d %d 500", p.x, p.y)
	}

	if goals.scoreLeft {
		sort.Sort(ByXDesc(snaffles))
	} else {
		sort.Sort(ByX(snaffles))
	}

	w.target = snaffles[0]
	p := w.target.Pos()
	return fmt.Sprintf("MOVE %d %d 150", p.x, p.y)
}

// Attack built a agressive action for the Wizard.
func (w *Wizard) Attack(snaffles []Object) string {
	w.target = nil
	p := goals.theirs
	if w.hasSnaffle {
		return fmt.Sprintf("THROW %d %d 500", p.x, p.y)
	}

	dists := ComputeDistance(w, snaffles)
	min := 0
	for i := 1; i < len(dists); i++ {
		if dists[i] < dists[min] {
			min = i
		}
	}
	w.target = snaffles[min]
	p = w.target.Pos()

	return fmt.Sprintf("MOVE %d %d 150", p.x, p.y)
}

func (w *Wizard) Avoid(bludgers []Object) (string, bool) {
	dist := ComputeDistance(w, bludgers)
	for i, d := range dist {
		if d < float64(w.radius+bludgers[i].Radius()) {
			// Compute cross product with rotation matrix.
			// Or find another way to avoid bludger.
			return "", true
		}
	}
	return "", false
}

type Game struct {
	objects   map[int]Object
	snaffles  []Object
	bludgers  []Object
	players   []Object
	opponents []Object
}

func NewGame() *Game {
	g := &Game{
		objects: make(map[int]Object),
	}
	return g
}

func (g *Game) Reset() {
	g.snaffles = make([]Object, 0, MaxSnaffles)
	g.players = make([]Object, 0, WizardsPerPlayer)
	g.opponents = make([]Object, 0, WizardsPerPlayer)
}

func (g *Game) Update(id int, objType string, x, y, vx, vy, state int) {
	obj, ok := g.objects[id]
	switch objType {
	case "WIZARD", "OPPONENT_WIZARD":
		if !ok {
			obj = NewWizard(id)
			g.objects[id] = obj
		}
		if objType == "WIZARD" {
			g.players = append(g.players, obj)
		} else {
			g.opponents = append(g.opponents, obj)
		}
	case "SNAFFLE":
		if !ok {
			obj = NewGameObject(id, RadiusSnaffle)
			g.objects[id] = obj
		}
		g.snaffles = append(g.snaffles, obj)

	case "BLUDGER":
		if !ok {
			obj = NewGameObject(id, RadiusSnaffle)
			g.objects[id] = obj
		}
		g.bludgers = append(g.bludgers, obj)
	}
	obj.Update(x, y, vx, vy, state)
}

func main() {
	// myTeamId: if 0 you need to score on the right of the map, if 1 you need to score on the left
	var myTeamId int
	fmt.Scan(&myTeamId)
	if myTeamId == 1 {
		goals.scoreLeft = true
		goals.theirs, goals.mine = goals.mine, goals.theirs
	}

	game := NewGame()
	for {
		// entities: number of entities still in game
		var entities int
		fmt.Scan(&entities)

		game.Reset()
		for i := 0; i < entities; i++ {
			// entityId: entity identifier
			// entityType: "WIZARD", "OPPONENT_WIZARD" or "SNAFFLE" (or "BLUDGER" after first league)
			// x: position
			// y: position
			// vx: velocity
			// vy: velocity
			// state: 1 if the wizard is holding a Snaffle, 0 otherwise
			var entityId int
			var entityType string
			var x, y, vx, vy, state int
			fmt.Scan(&entityId, &entityType, &x, &y, &vx, &vy, &state)
			game.Update(entityId, entityType, x, y, vx, vy, state)
		}

		//fmt.Fprintf(os.Stderr, "%#v\n", game)

		// Edit this line to indicate the action for each wizard (0 <= thrust <= 150, 0 <= power <= 500)
		// i.e.: "MOVE x y thrust" or "THROW x y power"

		snaffles := game.snaffles
		w := game.players[0].(*Wizard)
		action := w.Attack(snaffles)
		if w.target != nil && len(snaffles) > 2 {
			snaffles = RemoveFromSlice(w.target, snaffles)
		}
		fmt.Println(action)

		w = game.players[1].(*Wizard)
		action = w.Defend(snaffles)
		fmt.Println(action)
	}
}
