package main

import (
	"fmt"
	"os"
)

type Location struct {
	x int
	y int
}

func (l *Location) pgn() string {
	return (fmt.Sprintf("%s%d", string(l.x+97), l.y+1))
}

type Move struct {
	s      Location
	attack bool
}

type Piece int8

type MovesFromPos struct {
	s Location
	m [][]Move
}

var movelib map[Piece][]MovesFromPos

// Pawn

func exportPiece(movelib []MovesFromPos, out *os.File) {
	for _, mfp := range movelib {
		/*out.Write(mfp.s.pgn())
		out.Write([]byte(" ,"))*/
		for _, mset := range mfp.m {
			for _, move := range mset {
				attack := 0
				if move.attack {
					attack = 1
				}
				out.Write([]byte(fmt.Sprintf("0x%02x,",
					(1<<7)|((attack&0x1)<<6)|((move.s.x&0b111)<<3)|(move.s.y&0b111),
				)))
			}
			for i := len(mset); i < 7; i++ {
				out.Write([]byte("0x00,"))
			}
			//out.Write([]byte("/**/"))
		}
		for i := len(mfp.m); i < 8; i++ {
			for i := 0; i < 7; i++ {
				out.Write([]byte("0x00,"))
			}
			//out.Write([]byte("/**/"))
		}
		out.Write([]byte("\n\t"))
	}
}

func main() {

	movelib = make(map[Piece][]MovesFromPos)
	// WHITE PAWN
	{
		dir := 1
		movelib['p'] = make([]MovesFromPos, 0)

		for x := 0; x < 8; x++ {
			for y := 0; y < 8; y++ {
				movements := MovesFromPos{}
				movements.s.x = x
				movements.s.y = y

				if y != 0 && y != 7 {
					movements.m = append(movements.m, make([]Move, 0))
					movements.m[0] = append(movements.m[0], Move{s: Location{x: x, y: y + dir}, attack: false})
					if y == 1 {
						movements.m[0] = append(movements.m[0], Move{s: Location{x: x, y: y + 2*dir}, attack: false})
					}

					if x < 7 {
						movements.m = append(movements.m, make([]Move, 0))
						movements.m[1] = append(movements.m[1], Move{s: Location{x: x + 1, y: y + dir}, attack: true})
					}

					if x > 0 {
						movements.m = append(movements.m, make([]Move, 0))
						movements.m[len(movements.m)-1] = append(movements.m[len(movements.m)-1], Move{s: Location{x: x - 1, y: y + dir}, attack: true})
					}
				}
				if y == 6 {

				}

				movelib['p'] = append(movelib['p'], movements)
			}
		}
	}
	// BLACK PAWN
	{
		dir := -1
		movelib['P'] = make([]MovesFromPos, 0)

		for x := 0; x < 8; x++ {
			for y := 0; y < 8; y++ {
				movements := MovesFromPos{}
				movements.s.x = x
				movements.s.y = y

				if y != 0 && y != 7 {
					movements.m = append(movements.m, make([]Move, 0))
					movements.m[0] = append(movements.m[0], Move{s: Location{x: x, y: y + dir}, attack: false})
					if y == 6 {
						movements.m[0] = append(movements.m[0], Move{s: Location{x: x, y: y + 2*dir}, attack: false})
					}

					if x < 7 {
						movements.m = append(movements.m, make([]Move, 0))
						movements.m[1] = append(movements.m[1], Move{s: Location{x: x + 1, y: y + dir}, attack: true})
					}

					if x > 0 {
						movements.m = append(movements.m, make([]Move, 0))
						movements.m[len(movements.m)-1] = append(movements.m[len(movements.m)-1], Move{s: Location{x: x - 1, y: y + dir}, attack: true})
					}
				}
				movelib['P'] = append(movelib['P'], movements)
			}
		}
	}
	//ROOK
	{
		movelib['r'] = make([]MovesFromPos, 0)

		for x := 0; x < 8; x++ {
			for y := 0; y < 8; y++ {
				movements := MovesFromPos{}
				movements.s.x = x
				movements.s.y = y

				if x < 7 {
					movements.m = append(movements.m, make([]Move, 0))
				}
				for mx := x + 1; mx < 8; mx++ {

					movements.m[len(movements.m)-1] = append(movements.m[len(movements.m)-1], Move{s: Location{x: mx, y: y}, attack: true})
				}

				if x > 0 {
					movements.m = append(movements.m, make([]Move, 0))
				}
				for mx := x - 1; mx >= 0; mx -= 1 {
					movements.m[len(movements.m)-1] = append(movements.m[len(movements.m)-1], Move{s: Location{x: mx, y: y}, attack: true})
				}

				if y < 7 {
					movements.m = append(movements.m, make([]Move, 0))
				}
				for my := y + 1; my < 8; my++ {
					movements.m[len(movements.m)-1] = append(movements.m[len(movements.m)-1], Move{s: Location{x: x, y: my}, attack: true})
				}

				if y > 0 {
					movements.m = append(movements.m, make([]Move, 0))
				}
				for my := y - 1; my >= 0; my -= 1 {
					movements.m[len(movements.m)-1] = append(movements.m[len(movements.m)-1], Move{s: Location{x: x, y: my}, attack: true})
				}

				movelib['r'] = append(movelib['r'], movements)
			}
		}
	}

	//BISHOP
	{
		movelib['b'] = make([]MovesFromPos, 0)

		for x := 0; x < 8; x++ {
			for y := 0; y < 8; y++ {
				movements := MovesFromPos{}
				movements.s.x = x
				movements.s.y = y

				if x < 7 && y < 7 {
					movements.m = append(movements.m, make([]Move, 0))
				}
				for mx, my := x+1, y+1; mx < 8 && my < 8; {
					movements.m[len(movements.m)-1] = append(movements.m[len(movements.m)-1],
						Move{s: Location{x: mx, y: my}, attack: true})

					mx++
					my++
				}

				if x > 0 && y > 0 {
					movements.m = append(movements.m, make([]Move, 0))
				}
				for mx, my := x-1, y-1; mx >= 0 && my >= 0; {
					movements.m[len(movements.m)-1] = append(movements.m[len(movements.m)-1],
						Move{s: Location{x: mx, y: my}, attack: true})

					mx--
					my--
				}

				if x > 0 && y < 7 {
					movements.m = append(movements.m, make([]Move, 0))
				}
				for mx, my := x-1, y+1; mx >= 0 && my < 8; {
					movements.m[len(movements.m)-1] = append(movements.m[len(movements.m)-1],
						Move{s: Location{x: mx, y: my}, attack: true})

					mx--
					my++
				}

				if x < 7 && y > 0 {
					movements.m = append(movements.m, make([]Move, 0))
				}
				for mx, my := x+1, y-1; mx < 8 && my >= 0; {
					movements.m[len(movements.m)-1] = append(movements.m[len(movements.m)-1],
						Move{s: Location{x: mx, y: my}, attack: true})

					mx++
					my--
				}

				movelib['b'] = append(movelib['b'], movements)
			}
		}
	}

	//Knight
	{
		movelib['n'] = make([]MovesFromPos, 0)

		for x := 0; x < 8; x++ {
			for y := 0; y < 8; y++ {
				movements := MovesFromPos{}
				movements.s.x = x
				movements.s.y = y

				if x < 6 {
					if y > 0 {
						movements.m = append(movements.m, make([]Move, 0))
						movements.m[len(movements.m)-1] = append(movements.m[len(movements.m)-1],
							Move{s: Location{x: x + 2, y: y - 1}, attack: true})
					}
					if y < 7 {
						movements.m = append(movements.m, make([]Move, 0))
						movements.m[len(movements.m)-1] = append(movements.m[len(movements.m)-1],
							Move{s: Location{x: x + 2, y: y + 1}, attack: true})
					}

				}
				if x > 1 {
					if y > 0 {
						movements.m = append(movements.m, make([]Move, 0))
						movements.m[len(movements.m)-1] = append(movements.m[len(movements.m)-1],
							Move{s: Location{x: x - 2, y: y - 1}, attack: true})
					}
					if y < 7 {
						movements.m = append(movements.m, make([]Move, 0))
						movements.m[len(movements.m)-1] = append(movements.m[len(movements.m)-1],
							Move{s: Location{x: x - 2, y: y + 1}, attack: true})
					}

				}

				if y < 6 {
					if x > 0 {
						movements.m = append(movements.m, make([]Move, 0))
						movements.m[len(movements.m)-1] = append(movements.m[len(movements.m)-1],
							Move{s: Location{x: x - 1, y: y + 2}, attack: true})
					}
					if x < 7 {
						movements.m = append(movements.m, make([]Move, 0))
						movements.m[len(movements.m)-1] = append(movements.m[len(movements.m)-1],
							Move{s: Location{x: x + 1, y: y + 2}, attack: true})
					}

				}

				if y > 1 {
					if x > 0 {
						movements.m = append(movements.m, make([]Move, 0))
						movements.m[len(movements.m)-1] = append(movements.m[len(movements.m)-1],
							Move{s: Location{x: x - 1, y: y - 2}, attack: true})
					}
					if x < 7 {
						movements.m = append(movements.m, make([]Move, 0))
						movements.m[len(movements.m)-1] = append(movements.m[len(movements.m)-1],
							Move{s: Location{x: x + 1, y: y - 2}, attack: true})
					}

				}

				movelib['n'] = append(movelib['n'], movements)
			}
		}
	}

	//Queen
	{
		movelib['q'] = make([]MovesFromPos, 0)

		for x := 0; x < 8; x++ {
			for y := 0; y < 8; y++ {
				movements := MovesFromPos{}
				movements.s.x = x
				movements.s.y = y

				if x < 7 {
					movements.m = append(movements.m, make([]Move, 0))
				}
				for mx := x + 1; mx < 8; mx++ {

					movements.m[len(movements.m)-1] = append(movements.m[len(movements.m)-1], Move{s: Location{x: mx, y: y}, attack: true})
				}

				if x > 0 {
					movements.m = append(movements.m, make([]Move, 0))
				}
				for mx := x - 1; mx >= 0; mx -= 1 {
					movements.m[len(movements.m)-1] = append(movements.m[len(movements.m)-1], Move{s: Location{x: mx, y: y}, attack: true})
				}

				if y < 7 {
					movements.m = append(movements.m, make([]Move, 0))
				}
				for my := y + 1; my < 8; my++ {
					movements.m[len(movements.m)-1] = append(movements.m[len(movements.m)-1], Move{s: Location{x: x, y: my}, attack: true})
				}

				if y > 0 {
					movements.m = append(movements.m, make([]Move, 0))
				}
				for my := y - 1; my >= 0; my -= 1 {
					movements.m[len(movements.m)-1] = append(movements.m[len(movements.m)-1], Move{s: Location{x: x, y: my}, attack: true})
				}

				if x < 7 && y < 7 {
					movements.m = append(movements.m, make([]Move, 0))
				}
				for mx, my := x+1, y+1; mx < 8 && my < 8; {
					movements.m[len(movements.m)-1] = append(movements.m[len(movements.m)-1],
						Move{s: Location{x: mx, y: my}, attack: true})

					mx++
					my++
				}

				if x > 0 && y > 0 {
					movements.m = append(movements.m, make([]Move, 0))
				}
				for mx, my := x-1, y-1; mx >= 0 && my >= 0; {
					movements.m[len(movements.m)-1] = append(movements.m[len(movements.m)-1],
						Move{s: Location{x: mx, y: my}, attack: true})

					mx--
					my--
				}

				if x > 0 && y < 7 {
					movements.m = append(movements.m, make([]Move, 0))
				}
				for mx, my := x-1, y+1; mx >= 0 && my < 8; {
					movements.m[len(movements.m)-1] = append(movements.m[len(movements.m)-1],
						Move{s: Location{x: mx, y: my}, attack: true})

					mx--
					my++
				}

				if x < 7 && y > 0 {
					movements.m = append(movements.m, make([]Move, 0))
				}
				for mx, my := x+1, y-1; mx < 8 && my >= 0; {
					movements.m[len(movements.m)-1] = append(movements.m[len(movements.m)-1],
						Move{s: Location{x: mx, y: my}, attack: true})

					mx++
					my--
				}

				movelib['q'] = append(movelib['q'], movements)
			}
		}
	}

	//KING
	{
		movelib['k'] = make([]MovesFromPos, 0)

		for x := 0; x < 8; x++ {
			for y := 0; y < 8; y++ {
				movements := MovesFromPos{}
				movements.s.x = x
				movements.s.y = y

				if x < 7 {
					movements.m = append(movements.m, make([]Move, 0))
					movements.m[len(movements.m)-1] = append(movements.m[len(movements.m)-1],
						Move{s: Location{x: x + 1, y: y}, attack: true})
					if x == 4 && y == 0 {
						movements.m[len(movements.m)-1] = append(movements.m[len(movements.m)-1],
							Move{s: Location{x: x + 2, y: y}, attack: true})
					}
				}

				if x > 0 {
					movements.m = append(movements.m, make([]Move, 0))
					movements.m[len(movements.m)-1] = append(movements.m[len(movements.m)-1],
						Move{s: Location{x: x - 1, y: y}, attack: true})

					if x == 4 && y == 0 {
						movements.m[len(movements.m)-1] = append(movements.m[len(movements.m)-1],
							Move{s: Location{x: x - 2, y: y}, attack: true})
					}
				}

				if y < 7 {
					movements.m = append(movements.m, make([]Move, 0))
					movements.m[len(movements.m)-1] = append(movements.m[len(movements.m)-1],
						Move{s: Location{x: x, y: y + 1}, attack: true})
				}

				if y > 0 {
					movements.m = append(movements.m, make([]Move, 0))
					movements.m[len(movements.m)-1] = append(movements.m[len(movements.m)-1],
						Move{s: Location{x: x, y: y - 1}, attack: true})
				}

				if x < 7 && y < 7 {
					movements.m = append(movements.m, make([]Move, 0))
					movements.m[len(movements.m)-1] = append(movements.m[len(movements.m)-1],
						Move{s: Location{x: x + 1, y: y + 1}, attack: true})
				}

				if x > 0 && y > 0 {
					movements.m = append(movements.m, make([]Move, 0))
					movements.m[len(movements.m)-1] = append(movements.m[len(movements.m)-1],
						Move{s: Location{x: x - 1, y: y - 1}, attack: true})
				}

				if x > 0 && y < 7 {
					movements.m = append(movements.m, make([]Move, 0))
					movements.m[len(movements.m)-1] = append(movements.m[len(movements.m)-1],
						Move{s: Location{x: x - 1, y: y + 1}, attack: true})
				}

				if x < 7 && y > 0 {
					movements.m = append(movements.m, make([]Move, 0))
					movements.m[len(movements.m)-1] = append(movements.m[len(movements.m)-1],
						Move{s: Location{x: x + 1, y: y - 1}, attack: true})
				}

				movelib['k'] = append(movelib['k'], movements)
			}
		}
	}

	// BLACK KING
	{
		movelib['K'] = make([]MovesFromPos, 0)

		for x := 0; x < 8; x++ {
			for y := 0; y < 8; y++ {
				movements := MovesFromPos{}
				movements.s.x = x
				movements.s.y = y

				if x < 7 {
					movements.m = append(movements.m, make([]Move, 0))
					movements.m[len(movements.m)-1] = append(movements.m[len(movements.m)-1],
						Move{s: Location{x: x + 1, y: y}, attack: true})
					if x == 4 && y == 7 {
						movements.m[len(movements.m)-1] = append(movements.m[len(movements.m)-1],
							Move{s: Location{x: x + 2, y: y}, attack: true})
					}
				}

				if x > 0 {
					movements.m = append(movements.m, make([]Move, 0))
					movements.m[len(movements.m)-1] = append(movements.m[len(movements.m)-1],
						Move{s: Location{x: x - 1, y: y}, attack: true})

					if x == 4 && y == 7 {
						movements.m[len(movements.m)-1] = append(movements.m[len(movements.m)-1],
							Move{s: Location{x: x - 2, y: y}, attack: true})
					}
				}

				if y < 7 {
					movements.m = append(movements.m, make([]Move, 0))
					movements.m[len(movements.m)-1] = append(movements.m[len(movements.m)-1],
						Move{s: Location{x: x, y: y + 1}, attack: true})
				}

				if y > 0 {
					movements.m = append(movements.m, make([]Move, 0))
					movements.m[len(movements.m)-1] = append(movements.m[len(movements.m)-1],
						Move{s: Location{x: x, y: y - 1}, attack: true})
				}

				if x < 7 && y < 7 {
					movements.m = append(movements.m, make([]Move, 0))
					movements.m[len(movements.m)-1] = append(movements.m[len(movements.m)-1],
						Move{s: Location{x: x + 1, y: y + 1}, attack: true})
				}

				if x > 0 && y > 0 {
					movements.m = append(movements.m, make([]Move, 0))
					movements.m[len(movements.m)-1] = append(movements.m[len(movements.m)-1],
						Move{s: Location{x: x - 1, y: y - 1}, attack: true})
				}

				if x > 0 && y < 7 {
					movements.m = append(movements.m, make([]Move, 0))
					movements.m[len(movements.m)-1] = append(movements.m[len(movements.m)-1],
						Move{s: Location{x: x - 1, y: y + 1}, attack: true})
				}

				if x < 7 && y > 0 {
					movements.m = append(movements.m, make([]Move, 0))
					movements.m[len(movements.m)-1] = append(movements.m[len(movements.m)-1],
						Move{s: Location{x: x + 1, y: y - 1}, attack: true})
				}

				movelib['K'] = append(movelib['K'], movements)
			}
		}
	}

	/*
		Offset	Piece
		0		PAWN
		3584	BLACK PAWN
		7168	KNIGHT
		10752	ROOK
		14336	BISHOP
		17920	QUEEN
		21504	KING
		25088	BLACK KING

		MoveOffset:
		0	a1
		56	a2
		112	a3
		168	a4
		224	a5
		280	a6
		336	a7
		392	a8
		448	b1
		...

		offset = 56*y+56*8*x

		move = 1 byte (0 unused 0 attack_flag 000 X 000 Y)
		a move line is 7 at max so 7 bytes.
		the max move lines is 8
		startpos is 1 byte
		1 move from pos = 8 lines * 7 moves = 56 bytes
		64 movefrompos = 3584 bytes


	*/

	moveset, _ := os.Create("moveset/moveset.go")
	moveset.Write([]byte("package moveset\nvar Mset = [...]uint8{\n\t"))
	moveset.WriteString("//Pawn\n\t")
	exportPiece(movelib['p'], moveset)
	moveset.WriteString("//Black Pawn\n\t")
	exportPiece(movelib['P'], moveset)
	moveset.WriteString("//Knight\n\t")
	exportPiece(movelib['n'], moveset)
	moveset.WriteString("//Rook\n\t")
	exportPiece(movelib['r'], moveset)
	moveset.WriteString("//Bishop\n\t")
	exportPiece(movelib['b'], moveset)
	moveset.WriteString("//Queen\n\t")
	exportPiece(movelib['q'], moveset)
	moveset.WriteString("//King\n\t")
	exportPiece(movelib['k'], moveset)
	moveset.WriteString("//Black King\n\t")
	exportPiece(movelib['K'], moveset)
	moveset.Write([]byte("\n}"))
	/*for key, value := range movelib {
		moveset.Write([]byte("#"))
		moveset.Write([]byte(string(key)))
		moveset.Write([]byte("\n"))
		for _, mfp := range value {
			moveset.Write(mfp.s.pgn())
			moveset.Write([]byte(" "))
			for mset_idx, mset := range mfp.m {
				for _, m := range mset {
					moveset.Write(m.s.pgn())
					if m.attack {
						moveset.Write([]byte("a"))
					}
					if m != mset[len(mset)-1] {
						moveset.Write([]byte(","))
					}

				}
				if mset_idx != len(mfp.m)-1 {
					moveset.Write([]byte(";"))
				}
			}
			moveset.Write([]byte("\n"))
		}
	}*/
}
