package main

import (
	"fmt"
	"log"
	"strings"
	"yrk06/chess-backend/moveset"
)

/*
Offset	Piece
0		PAWN
3584	BLACK PAWN
7168	KNIGHT
10752	ROOK
14336	BISHOP
17920	QUEEN
21504	KING


*/
var pieceMap = map[byte]int{
	'p': 0,
	'P': 3584,

	'n': 7168,
	'N': 7168,

	'r': 10752,
	'R': 10752,

	'b': 14336,
	'B': 14336,

	'q': 17920,
	'Q': 17920,

	'k': 21504,
	'K': 21504,
}

type Location struct {
	x int
	y int
}

func (l *Location) toFen() int {
	return (7-l.y)*9 + (l.x)
}

func (l *Location) pgn() string {
	return fmt.Sprintf("%s%d", string(l.x+97), l.y+1)
}

func (l *Location) frompgn(pos string) {
	l.x = int(pos[0]) - 97
	l.y = int(pos[1]-'0') - 1
}

func (l *Location) toByte() uint8 {
	return uint8((1 << 7) | ((l.x & 0b111) << 3) | (l.y & 0b111))
}

func (l *Location) fromByte(b uint8) {
	l.x = int((b >> 3) & 0b111)
	l.y = int((b) & 0b111)
}

type Piece uint8

type PlayerPieces [16]Piece

var indexPieceMap = map[int]string{
	0:  "r",
	1:  "n",
	2:  "b",
	3:  "q",
	4:  "k",
	5:  "b",
	6:  "n",
	7:  "r",
	8:  "p",
	9:  "p",
	10: "p",
	11: "p",
	12: "p",
	13: "p",
	14: "p",
	15: "p",
}

type Chessboard struct {
	white     PlayerPieces
	black     PlayerPieces
	toMove    bool
	wK        bool
	wQ        bool
	bK        bool
	bQ        bool
	enpassant uint8
	mc        int
	rounds    int
}

func (c *Chessboard) Init() {
	c.wK = true
	c.wQ = true
	c.bK = true
	c.bQ = true
	c.toMove = true

	for y := 0; y < 2; y++ {
		for x := 0; x < 8; x++ {
			l := Location{x: x, y: y}
			c.white[y*8+x] = Piece(l.toByte())
		}
	}
	for y := 0; y < 2; y++ {
		for x := 0; x < 8; x++ {
			l := Location{x: x, y: 7 - y}
			c.black[y*8+x] = Piece(l.toByte())
		}
	}

}

//rnbqkbnr/pppppppp/11111111/11111111/11111111/11111111/PPPPPPPP/RNBQKBNR w KQkq - 0 1
func (c *Chessboard) fen() string {

	enpassant := "-"
	if (c.enpassant&(1<<7))>>7 == 1 {
		move := Location{}
		move.fromByte(c.enpassant)
		enpassant = move.pgn()
	}

	fen := []byte(
		fmt.Sprintf("11111111/11111111/11111111/11111111/11111111/11111111/11111111/11111111 %s %s%s%s%s %s 0 1",
			map[bool]string{true: "w", false: "b"}[c.toMove],

			map[bool]string{true: "K", false: "-"}[c.wK],
			map[bool]string{true: "Q", false: "-"}[c.wQ],

			map[bool]string{true: "k", false: "-"}[c.bK],
			map[bool]string{true: "q", false: "-"}[c.bQ],

			enpassant,
		),
	)
	for idx, wp := range c.white {
		if (wp>>7)&1 != 1 {
			continue
		}
		l := Location{}
		l.fromByte(uint8(wp))
		fen[l.toFen()] = strings.ToUpper(indexPieceMap[idx])[0]
	}
	for idx, wp := range c.black {
		if (wp>>7)&1 != 1 {
			continue
		}
		l := Location{}
		l.fromByte(uint8(wp))
		fen[l.toFen()] = indexPieceMap[idx][0]
	}
	return string(fen)
}

func (c *Chessboard) hasPieceInPosition(position uint8, enpassant bool) (bool, int) {
	l := Location{}
	l.fromByte(position)
	if enpassant {
		if position == c.enpassant {
			return true, int(((position >> 3) & 0b111) + 8)
		}
	}
	for idx, p := range c.white {
		if position == uint8(p) {
			return true, 1<<4 | idx
		}
	}
	for idx, p := range c.black {
		if position == uint8(p) {
			return true, idx
		}
	}
	return false, 0
}

func (c *Chessboard) pieceAttacks(piece int, team bool, end_pos Location) bool {

	start_pos := Location{}
	piecei := pieceMap[indexPieceMap[int(piece)][0]]
	if team {
		start_pos.fromByte(uint8(c.white[piece]))
		piecei = pieceMap[indexPieceMap[int(piece)][0]]
	} else {
		start_pos.fromByte(uint8(c.black[piece]))
		if indexPieceMap[int(piece)][0] == 'p' {
			piecei = pieceMap['P']
		}
	}
	ep := end_pos.toByte()
	for moveLines := 0; moveLines < 8; moveLines++ {
		for move := 0; move < 7; move++ {
			value := moveset.Mset[piecei+56*start_pos.y+56*8*start_pos.x+moveLines*7+move]

			if value == 0 {
				break
			}
			if (ep & 0b111111) == (value & 0b111111) {

				return true

			} else {
				if c, _ := c.hasPieceInPosition((value&0b111111)|1<<7, false); c {
					break
				}
			}

		}
	}
	return false
}

func (c *Chessboard) isSquareAttacked(team_attacking bool, loc Location) bool {
	if team_attacking {
		for idx, pos := range c.white {
			if pos == 0 {
				continue
			}

			attack := c.pieceAttacks(idx, team_attacking, loc)
			if attack {
				return true
			}
		}
	} else {
		for idx, pos := range c.black {
			if pos == 0 {
				continue
			}

			attack := c.pieceAttacks(idx, team_attacking, loc)
			if attack {
				return true
			}
		}
	}
	return false
}

func (c *Chessboard) verifyState(team bool) bool {
	if team {
		kingloc := Location{}
		kingloc.fromByte(uint8(c.white[4]))
		return !c.isSquareAttacked(!team, kingloc)
	} else {
		kingloc := Location{}
		kingloc.fromByte(uint8(c.black[4]))
		return !c.isSquareAttacked(!team, kingloc)
	}
	return true
}

func (c *Chessboard) MakeMove(piece uint8, team bool, end_pos Location) bool {

	// Setup Vars
	start_pos := Location{}
	piecei := pieceMap[indexPieceMap[int(piece)][0]]
	if team {
		start_pos.fromByte(uint8(c.white[piece]))
		piecei = pieceMap[indexPieceMap[int(piece)][0]]
	} else {
		start_pos.fromByte(uint8(c.black[piece]))
		if indexPieceMap[int(piece)][0] == 'p' {
			piecei = pieceMap['P']
		}
	}

	// Check if piece can move there and check if the path is blocked
	valid := false
	enpassant := false
	enpassant_active := false

	// Check for castling
	castling := false
	if piece == 4 {
		if team {
			if c.wK {
				if end_pos == (Location{x: 6, y: 0}) {

					if c.isSquareAttacked(!team, start_pos) {
						return false
					}
					sq1 := Location{x: 5, y: 0}
					if c, _ := c.hasPieceInPosition(sq1.toByte(), false); c {
						return false
					}
					if c.isSquareAttacked(!team, sq1) {
						return false
					}

					sq2 := Location{x: 6, y: 0}
					if c, _ := c.hasPieceInPosition(sq2.toByte(), false); c {
						return false
					}
					if c.isSquareAttacked(!team, sq2) {
						return false
					}

					c.white[4] = Piece(end_pos.toByte())
					c.white[7] = Piece(sq1.toByte())
					castling = true
					c.wK = false
					c.wQ = false

				}
			}
			if c.wQ {
				if end_pos == (Location{x: 2, y: 0}) {

					if c.isSquareAttacked(!team, start_pos) {
						return false
					}
					sq1 := Location{x: 1, y: 0}
					if c, _ := c.hasPieceInPosition(sq1.toByte(), false); c {
						return false
					}

					sq2 := Location{x: 2, y: 0}
					if c, _ := c.hasPieceInPosition(sq2.toByte(), false); c {
						return false
					}
					if c.isSquareAttacked(!team, sq2) {
						return false
					}

					sq3 := Location{x: 3, y: 0}
					if c, _ := c.hasPieceInPosition(sq3.toByte(), false); c {
						return false
					}
					if c.isSquareAttacked(!team, sq3) {
						return false
					}

					c.white[4] = Piece(end_pos.toByte())
					c.white[0] = Piece(sq3.toByte())
					castling = true
					c.wK = false
					c.wQ = false

				}
			}

		} else {
			if c.bK {
				if end_pos == (Location{x: 6, y: 7}) {

					if c.isSquareAttacked(!team, start_pos) {
						return false
					}
					sq1 := Location{x: 5, y: 7}
					if c, _ := c.hasPieceInPosition(sq1.toByte(), false); c {
						return false
					}
					if c.isSquareAttacked(!team, sq1) {
						return false
					}

					sq2 := Location{x: 6, y: 7}
					if c, _ := c.hasPieceInPosition(sq2.toByte(), false); c {
						return false
					}
					if c.isSquareAttacked(!team, sq2) {
						return false
					}

					c.black[4] = Piece(end_pos.toByte())
					c.black[7] = Piece(sq1.toByte())
					castling = true
					c.bK = false
					c.bQ = false

				}
			}
			if c.bQ {
				if end_pos == (Location{x: 2, y: 7}) {

					if c.isSquareAttacked(!team, start_pos) {
						return false
					}
					sq1 := Location{x: 1, y: 7}
					if c, _ := c.hasPieceInPosition(sq1.toByte(), false); c {
						return false
					}

					sq2 := Location{x: 2, y: 7}
					if c, _ := c.hasPieceInPosition(sq2.toByte(), false); c {
						return false
					}
					if c.isSquareAttacked(!team, sq2) {
						return false
					}

					sq3 := Location{x: 3, y: 7}
					if c, _ := c.hasPieceInPosition(sq3.toByte(), false); c {
						return false
					}
					if c.isSquareAttacked(!team, sq3) {
						return false
					}

					c.black[4] = Piece(end_pos.toByte())
					c.black[0] = Piece(sq3.toByte())
					castling = true
					c.bK = false
					c.bQ = false

				}
			}
		}
	}
	if castling {
		c.toMove = !c.toMove
		return true
	}
	ep := end_pos.toByte()
	for moveLines := 0; moveLines < 8; moveLines++ {
		for move := 0; move < 7; move++ {
			value := moveset.Mset[piecei+56*start_pos.y+56*8*start_pos.x+moveLines*7+move]

			if value == 0 {
				break
			}
			if (ep & 0b111111) == (value & 0b111111) {

				if piece > 7 {
					if moveLines == 0 {
						// En passant
						if move == 1 {
							l := Location{}
							l.fromByte(moveset.Mset[piecei+56*start_pos.y+56*8*start_pos.x+moveLines*7])
							c.enpassant = l.toByte()
							enpassant_active = true
						}
						valid = true
					} else {
						// Take en passant into account
						if c, _ := c.hasPieceInPosition((value&0b111111)|1<<7, true); c {
							valid = true
							enpassant = true
						}
					}
				} else {
					valid = true
				}

			} else {
				if c, _ := c.hasPieceInPosition((value&0b111111)|1<<7, false); c {
					break
				}
			}

		}
		if valid {
			break
		}
	}

	if !valid {
		return false
	}

	// Check if piece can capture target (if there is a target)
	target, target_p := c.hasPieceInPosition(ep, enpassant)
	if target {
		if !enpassant {
			pieceTeam := target_p >> 4
			// Cannot capture your own pieces
			if pieceTeam == 1 && team {
				return false
			}
			if pieceTeam == 0 && !team {
				return false
			}
		}

	}

	// Make the Move and check if king is in check
	if team {
		c.white[piece] = Piece(ep)
	} else {
		c.black[piece] = Piece(ep)
	}
	if target {
		if team {
			c.black[target_p&0xF] = 0
		} else {
			c.white[target_p&0xF] = 0
		}
	}

	if !enpassant_active {
		c.enpassant = 0
	}
	if !c.verifyState(team) {
		if team {
			c.white[piece] = Piece(start_pos.toByte())
		} else {
			c.black[piece] = Piece(start_pos.toByte())
		}
		if target {
			if team {
				c.black[target_p&0xF] = Piece(ep)
			} else {
				c.white[target_p&0xF] = Piece(ep)
			}
		}
		return false
	}

	if piece == 4 {
		if team {
			c.wK = false
			c.wQ = false
		}

	}
	c.toMove = !c.toMove
	return true
}

func (c *Chessboard) Duplicate() Chessboard {
	return Chessboard{
		white:     c.white,
		black:     c.black,
		toMove:    c.toMove,
		wK:        c.wK,
		wQ:        c.wQ,
		bK:        c.bK,
		bQ:        c.bQ,
		enpassant: c.enpassant,
		mc:        c.mc,
		rounds:    c.rounds,
	}
}

func (c *Chessboard) evaluate() int {
	return 0
}

func main() {
	start := "e4"
	end := "d4"

	//piece := start[0]
	start_pos := Location{}
	start_pos.frompgn(start)
	end_pos := Location{}
	end_pos.frompgn(end)

	//log.Println(isMoveValid(piece, start_pos, end_pos))

	board := Chessboard{}
	board.Init()

	board.black[5] = 0
	board.black[6] = 0
	board.black[12] = 0
	//board.black[3] = Piece(start_pos.toByte())
	board.black[1] = 0
	board.black[2] = 0
	board.black[3] = 0

	l := Location{x: 6, y: 7}
	board.MakeMove(4, false, l)
	log.Println(board.fen())

}
