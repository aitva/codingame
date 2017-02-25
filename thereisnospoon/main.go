package main

import "fmt"
import "os"
import "bufio"

//import "strings"
//import "strconv"

/**
 * Don't let the machines win. You are humanity's last hope...
 **/

type Node struct {
	X, Y          int
	Right, Bottom *Node
}

type Tree struct {
	root *Node
	len  int
}

func (t *Tree) Add(n *Node) {
	if t.root == nil {
		t.root = n
		t.len++
		return
	}

	e := t.root
	for e != nil {
		if e.Y == n.Y || e.Bottom.Y > n.Y {
			break
		}
		e = e.Bottom
	}
	if e == nil || e.Y != n.Y {
		e.Bottom = n
		t.len++
		return
	}

	for e != nil {
		if e.X == n.X {
			break
		}
		e = e.Right
	}
}

func (t *Tree) Front() *Node {
	return t.root
}

func main() {
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Buffer(make([]byte, 1000000), 1000000)

	// width: the number of cells on the X axis
	var width int
	scanner.Scan()
	fmt.Sscan(scanner.Text(), &width)

	// height: the number of cells on the Y axis
	var height int
	scanner.Scan()
	fmt.Sscan(scanner.Text(), &height)

	for i := 0; i < height; i++ {
		scanner.Scan()
		line := scanner.Text() // width characters, each either 0 or .
		for j := 0; j < width; j++ {
			r := line[j]
			if r != '0' {
				continue
			}
			// nodes = append(nodes, Node{
			// 	X:      j,
			// 	Y:      i,
			// 	Right:  left,
			// 	Bottom: top,
			// })
		}
	}

	// fmt.Fprintln(os.Stderr, "Debug messages...")

	// Three coordinates: a node, its right neighbor, its bottom neighbor
	fmt.Println("0 0 1 0 0 1")
}
