package main

import "fmt"
import "sort"
import "math"

const (
	MaxSnaffles     = 7
	WizardPerPlayer = 2
)

var goals = []Point{
	{x: 0, y: 3750},
	{x: 16000, y: 3750},
}

type Object interface {
	ID() int
	Update(x, y, vx, vy, state int)
	Pos() Point
}

type ByX []Object

func (a ByX) Len() int           { return len(a) }
func (a ByX) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ByX) Less(i, j int) bool { return a[i].Pos().x < a[j].Pos().x }

type Point struct {
	x int
	y int
}

type Vector struct {
	x int
	y int
}

type GameObject struct {
	id    int
	pos   Point
	v     Vector
	state int
}

func NewGameObject(id, x, y, vx, vy, state int) *GameObject {
	return &GameObject{
		id: id,
		pos: Point{
			x: x,
			y: y,
		},
		v: Vector{
			x: vx,
			y: vy,
		},
		state: state,
	}
}

func (s *GameObject) ID() int {
	return s.id
}

func (s *GameObject) Update(x, y, vx, vy, state int) {
	s.pos.x = x
	s.pos.y = y
	s.v.x = vx
	s.v.y = vy
	s.state = state
}

func (s *GameObject) Pos() Point {
	return s.pos
}

type Wizard struct {
	*GameObject
	hasSnaffle bool
}

func NewWizard(obj *GameObject) *Wizard {
	return &Wizard{
		GameObject: obj,
	}
}

func (w *Wizard) Update(x, y, vx, vy, state int) {
	w.GameObject.Update(x, y, vx, vy, state)
	w.hasSnaffle = false
	if state == 1 {
		w.hasSnaffle = true
	}
}

type Game struct {
	objects   map[int]Object
	snaffles  []Object
	players   []Object
	opponents []Object
	scoreLeft bool
	goals     struct {
		mine   Point
		theirs Point
	}
}

func NewGame(teamID int) *Game {
	g := &Game{
		scoreLeft: false,
		objects:   make(map[int]Object),
	}
	g.goals.mine = goals[0]
	g.goals.theirs = goals[1]
	if teamID == 1 {
		g.scoreLeft = true
		g.goals.theirs = goals[0]
		g.goals.mine = goals[1]
	}
	return g
}

func (g *Game) Reset() {
	g.snaffles = make([]Object, 0, MaxSnaffles)
	g.players = make([]Object, 0, WizardPerPlayer)
	g.opponents = make([]Object, 0, WizardPerPlayer)
}

func (g *Game) Update(id int, objType string, x, y, vx, vy, state int) {
	obj, ok := g.objects[id]
	if ok {
		obj.Update(x, y, vx, vy, state)
	}
	switch objType {
	case "WIZARD", "OPPONENT_WIZARD":
		w, ok := obj.(*Wizard)
		if !ok {
			o := NewGameObject(id, x, y, vx, vy, state)
			w = NewWizard(o)
			g.objects[id] = w
		}
		if objType == "WIZARD" {
			g.players = append(g.players, w)
		} else {
			g.opponents = append(g.opponents, w)
		}
	case "SNAFFLE":
		o, ok := obj.(*GameObject)
		if !ok {
			o = NewGameObject(id, x, y, vx, vy, state)
			g.objects[id] = o
		}
		g.snaffles = append(g.snaffles, o)
	}
}

func (g *Game) ComputeDistance(o Object, objs []Object) []float64 {
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

func (g *Game) MoveDefender() {
	power := 150
	action := "MOVE"
	w := g.players[0].(*Wizard)
	p := g.snaffles[1].Pos()

	if w.hasSnaffle {
		action = "THROW"
		p = g.goals.theirs
		power = 500
	}

	fmt.Printf("%s %d %d %d\n", action, p.x, p.y, power)
}

func (g *Game) MoveAttacker() {
	p := g.goals.theirs
	w := g.players[1].(*Wizard)

	if w.hasSnaffle {
		fmt.Printf("THROW %d %d 500\n", p.x, p.y)
		return
	}

	dists := g.ComputeDistance(w, g.snaffles)
	min := 0
	for i := 1; i < len(dists); i++ {
		if dists[i] < dists[min] {
			min = i
		}
	}
	p = g.snaffles[min].Pos()

	fmt.Printf("MOVE %d %d 150\n", p.x, p.y)
}

func main() {
	// myTeamId: if 0 you need to score on the right of the map, if 1 you need to score on the left
	var myTeamId int
	fmt.Scan(&myTeamId)

	game := NewGame(myTeamId)
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
		sort.Sort(ByX(game.snaffles))

		// Edit this line to indicate the action for each wizard (0 <= thrust <= 150, 0 <= power <= 500)
		// i.e.: "MOVE x y thrust" or "THROW x y power"
		game.MoveDefender()
		game.MoveAttacker()
	}
}
