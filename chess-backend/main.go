package main

import (
	"flag"
	"fmt"
	"log"
	"math"
	"net/http"
	"strings"
	"time"
	"yrk06/chess-backend/moveset"

	"github.com/gorilla/websocket"
)

const BOT_MINIMAX_DEPTH = 40

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

func pgnToByte(pgn string) uint8 {
	l := Location{}
	l.frompgn(pgn)
	return l.toByte()
}

type PossibleMove struct {
	piece   int
	end_pos Location
	c       Chessboard
}

type Piece uint8

type PlayerPieces [24]Piece

/*
	Map from playerPiece index to char representation of pieces
*/
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

/*
	Map from char representation to possible index
*/
var pieceIndexMap = map[byte][]int{
	'r': {0, 7, 16, 17, 18, 19, 20, 21, 22, 23},
	'n': {1, 6, 16, 17, 18, 19, 20, 21, 22, 23},
	'b': {2, 5, 16, 17, 18, 19, 20, 21, 22, 23},
	'q': {3, 16, 17, 18, 19, 20, 21, 22, 23},
	'k': {4},
	'p': {8, 9, 10, 11, 12, 13, 14, 15},
}

/*
	Chessboard object
*/
type Chessboard struct {
	white         PlayerPieces
	whitePieceMap map[int]string
	black         PlayerPieces
	blackPieceMap map[int]string
	toMove        bool
	wK            bool
	wQ            bool
	bK            bool
	bQ            bool
	enpassant     uint8
	mc            int
	rounds        int
}

/*
	Init checkboard with starting position
*/
func (c *Chessboard) Init() {

	// Set castling to true
	c.wK = true
	c.wQ = true
	c.bK = true
	c.bQ = true

	// White moves first
	c.toMove = true

	// Create the aditional piece data
	c.whitePieceMap = make(map[int]string)
	c.blackPieceMap = make(map[int]string)

	// Init white pieces
	for y := 0; y < 2; y++ {
		for x := 0; x < 8; x++ {
			l := Location{x: x, y: y}
			c.white[y*8+x] = Piece(l.toByte())
		}
	}

	// Init white pieces
	for y := 0; y < 2; y++ {
		for x := 0; x < 8; x++ {
			l := Location{x: x, y: 7 - y}
			c.black[y*8+x] = Piece(l.toByte())
		}
	}

}

//rnbqkbnr/pppppppp/11111111/11111111/11111111/11111111/PPPPPPPP/RNBQKBNR w KQkq - 0 1
/*
	Convert chessboard object to FEN string
*/
func (c *Chessboard) fen() string {

	// If there is a current en passant going on, add it to the FEN
	enpassant := "-"
	if (c.enpassant&(1<<7))>>7 == 1 {
		move := Location{}
		move.fromByte(c.enpassant)
		enpassant = move.pgn()
	}

	// Create FEN string and insert castling rights
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

	// Loop through all white pieces and Add them to the FEN
	for idx, wp := range c.white {

		//Check if piece is valid
		if (wp>>7)&1 != 1 {
			continue
		}

		//Create Location
		l := Location{}
		l.fromByte(uint8(wp))

		// Get piece char
		if idx > 15 {
			fen[l.toFen()] = strings.ToUpper(c.whitePieceMap[idx])[0]
		} else {
			fen[l.toFen()] = strings.ToUpper(indexPieceMap[idx])[0]
		}

	}

	// Loop through all black pieces and Add them to the FEN
	for idx, wp := range c.black {

		//Check if piece is valid
		if (wp>>7)&1 != 1 {
			continue
		}

		//Create Location
		l := Location{}
		l.fromByte(uint8(wp))

		// Get piece char
		if idx > 15 {
			fen[l.toFen()] = (c.blackPieceMap[idx])[0]
		} else {
			fen[l.toFen()] = (indexPieceMap[idx])[0]
		}
	}
	return string(fen)
}

/*
	Check if there are any pieces in the posisiton. Takes into account en passant (ghost piece).
	bit 5 is 1 if piece is white 0 otherwise
*/
func (c *Chessboard) hasPieceInPosition(position uint8, enpassant bool) (bool, int) {

	// Start with calculating En passant
	if enpassant {
		if position == c.enpassant {
			return true, int(((position >> 3) & 0b111) + 8)
		}
	}

	// Check all the white pieces
	for idx, p := range c.white {
		if position == uint8(p) {
			return true, 1<<5 | idx
		}
	}

	// Check all black pieces
	for idx, p := range c.black {
		if position == uint8(p) {
			return true, idx
		}
	}

	return false, 0
}

/*
	Check if PIECE from TEAM attacks END_POS
*/
func (c *Chessboard) pieceAttacks(piece int, team bool, end_pos Location) bool {

	// Get piece movement table offset
	start_pos := Location{}
	piecei := 0

	if team {
		start_pos.fromByte(uint8(c.white[piece]))

		if piece > 15 {
			piecei = pieceMap[c.whitePieceMap[int(piece)][0]]
		} else {
			piecei = pieceMap[indexPieceMap[int(piece)][0]]
		}

	} else {
		start_pos.fromByte(uint8(c.black[piece]))

		if piece > 15 {
			p := c.blackPieceMap[int(piece)][0]
			if p == 'p' {
				p = 'P'
			}
			piecei = pieceMap[p]
		} else {
			p := indexPieceMap[int(piece)][0]
			if p == 'p' {
				p = 'P'
			}
			piecei = pieceMap[p]
		}

	}

	// Loop through movement table
	ep := end_pos.toByte()
	for moveLines := 0; moveLines < 8; moveLines++ {
		for move := 0; move < 7; move++ {
			value := moveset.Mset[piecei+56*start_pos.y+56*8*start_pos.x+moveLines*7+move]

			// Value 0 is invalid movement
			if value == 0 {
				break
			}

			// If position is in the table
			if (ep & 0b111111) == (value & 0b111111) {

				// Special Pawn rules
				if piece > 7 && piece < 16 {
					if moveLines > 0 {
						// Pawns only attack on >1 movelines
						return true
					} else {
						return false
					}
				} else {
					return true
				}

			} else {

				// If movement blocked by piece
				if c, _ := c.hasPieceInPosition((value&0b111111)|1<<7, false); c {
					break
				}
			}

		}
	}
	return false
}

/*
	Check if square is attacked by TEAM_ATTACKING
*/
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

/*
	Check if state is valid (king is not in check)
*/
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
}

/*
	Attemps to move a piece and returns TRUE if move is valid
*/
func (c *Chessboard) MakeMove(piece uint8, team bool, end_pos Location, promote_to byte) bool {

	// Setup Vars
	start_pos := Location{}
	piecei := 0

	if team {
		start_pos.fromByte(uint8(c.white[piece]))

		if piece > 15 {
			piecei = pieceMap[c.whitePieceMap[int(piece)][0]]
		} else {
			piecei = pieceMap[indexPieceMap[int(piece)][0]]
		}

	} else {
		start_pos.fromByte(uint8(c.black[piece]))

		if piece > 15 {
			p := c.blackPieceMap[int(piece)][0]
			if p == 'p' {
				p = 'P'
			}
			piecei = pieceMap[p]
		} else {
			p := indexPieceMap[int(piece)][0]
			if p == 'p' {
				p = 'P'
			}
			piecei = pieceMap[p]
		}

	}

	// Control variables
	valid := false
	enpassant := false
	enpassant_active := false
	canAttack := true
	promotion := false
	promotion_idx := 0

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

	// Check if piece can move there
	ep := end_pos.toByte()
	for moveLines := 0; moveLines < 8; moveLines++ {
		for move := 0; move < 7; move++ {
			value := moveset.Mset[piecei+56*start_pos.y+56*8*start_pos.x+moveLines*7+move]

			if value == 0 {
				break
			}

			// if position is valid
			if (ep & 0b111111) == (value & 0b111111) {

				// Special Pawn rules
				if piece > 7 && piece < 16 {

					if moveLines == 0 {

						// check if doing En passant
						if move == 1 {
							l := Location{}
							l.fromByte(moveset.Mset[piecei+56*start_pos.y+56*8*start_pos.x+moveLines*7])
							c.enpassant = l.toByte()
							enpassant_active = true
						}
						canAttack = false
						valid = true
					} else {

						// Take en passant into account
						if co, _ := c.hasPieceInPosition((value&0b111111)|1<<7, true); co {
							valid = true
							if c.enpassant != 0 {
								enpassant = true
							}

						}
					}

					if end_pos.y == 7 && team {
						promotion = true
						log.Println("Pawn Promotion")

					}

					if end_pos.y == 0 && !team {
						promotion = true
						log.Println("Pawn Promotion")

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
		pieceTeam := target_p >> 5
		// Cannot capture your own pieces
		if pieceTeam == 1 && team {
			return false
		}
		if pieceTeam == 0 && !team {
			return false
		}
		if !canAttack {
			return false
		}

	}

	// Make the Move and check if king is in check
	if team {
		c.white[piece] = Piece(ep)
	} else {
		c.black[piece] = Piece(ep)
	}
	oldbQ := c.bQ
	oldbK := c.bK

	oldwQ := c.wQ
	oldwK := c.wK

	// Capture piece
	if target {
		if team {
			if target_p&0x1F == 0 {
				c.bQ = false
			}
			if target_p&0x1F == 7 {
				c.bK = false
			}
			c.black[target_p&0x1F] = 0
		} else {
			if target_p&0x1F == 0 {
				c.wQ = false
			}
			if target_p&0x1F == 7 {
				c.wK = false
			}
			c.white[target_p&0x1F] = 0
		}
	}

	if !enpassant_active {
		c.enpassant = 0
	}

	if promotion {
		if team {
			for i := 16; i < 24; i++ {
				if c.white[i] == 0 {
					c.whitePieceMap[i] = string(promote_to)
					c.white[i] = c.white[piece]
					c.white[piece] = 0
					promotion_idx = i
					break
				}
			}
		} else {
			for i := 16; i < 24; i++ {
				if c.black[i] == 0 {
					c.blackPieceMap[i] = string(promote_to)
					c.black[i] = c.black[piece]
					c.black[piece] = 0
					promotion_idx = i
					break
				}
			}
		}

	}

	// Check if move was valid by check rules and rollback if invalid
	if !c.verifyState(team) {
		if team {
			c.white[piece] = Piece(start_pos.toByte())
		} else {
			c.black[piece] = Piece(start_pos.toByte())
		}
		if target {
			if team {
				if target_p&0xF == 0 {
					c.bQ = oldbQ
				}
				if target_p&0xF == 7 {
					c.bK = oldbK
				}
				c.black[target_p&0xF] = Piece(ep)
			} else {
				if target_p&0xF == 0 {
					c.wQ = oldwQ
				}
				if target_p&0xF == 7 {
					c.wK = oldwK
				}
				c.white[target_p&0xF] = Piece(ep)
			}
		}

		if promotion {
			if team {
				c.white[promotion_idx] = 0
			} else {
				c.black[promotion_idx] = 0
			}

		}
		return false
	}

	// Remove checking rights
	if piece == 4 {
		if team {
			c.wK = false
			c.wQ = false
		}
	}

	c.toMove = !c.toMove
	return true
}

// Duplicate chessboard
func (c *Chessboard) Duplicate() Chessboard {

	board := Chessboard{
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

	board.blackPieceMap = make(map[int]string)
	for k, v := range c.blackPieceMap {
		board.blackPieceMap[k] = v
	}

	board.whitePieceMap = make(map[int]string)
	for k, v := range c.whitePieceMap {
		board.whitePieceMap[k] = v
	}

	return board
}

// Value per piece
var pieceValue = map[int]int{
	0: 500,
	1: 300,
	2: 300,
	3: 900,
	4: 42000,
	5: 300,
	6: 300,
	7: 500,

	8:  100,
	9:  100,
	10: 100,
	11: 100,
	12: 100,
	13: 100,
	14: 100,
	15: 100,
}

// Calculate all possible moves
func (c *Chessboard) possibleMoves(team bool) []PossibleMove {
	moves := make([]PossibleMove, 0)

	// White {} black pieces
	if team {
		for idx, pos := range c.white {
			if pos == 0 {
				continue
			}

			//Get piece memory offset
			start_pos := Location{}
			real_idx := idx
			if idx > 15 {
				real_idx = pieceIndexMap[c.whitePieceMap[idx][0]][0]
			}
			piecei := pieceMap[indexPieceMap[real_idx][0]]
			start_pos.fromByte(uint8(pos))

			// Check all moves
			for moveLines := 0; moveLines < 8; moveLines++ {
				for move := 0; move < 7; move++ {
					value := moveset.Mset[piecei+56*start_pos.y+56*8*start_pos.x+moveLines*7+move]

					if value == 0 {
						break
					}

					end_pos := Location{}
					end_pos.fromByte(value)
					newBoard := c.Duplicate()
					if !newBoard.MakeMove(uint8(idx), team, end_pos, 'q') {
						continue
					}

					// If move valid, append to list
					moves = append(moves, PossibleMove{piece: idx, end_pos: end_pos, c: newBoard})

				}
			}
		}
	} else {
		for idx, pos := range c.black {
			if pos == 0 {
				continue
			}
			piecei := 0
			if idx > 15 {
				piecei = pieceMap[c.blackPieceMap[int(idx)][0]]
			} else {
				piecei = pieceMap[indexPieceMap[int(idx)][0]]
			}
			start_pos := Location{}

			if idx > 7 && idx < 16 {
				piecei = pieceMap['P']
			}

			start_pos.fromByte(uint8(pos))

			for moveLines := 0; moveLines < 8; moveLines++ {
				for move := 0; move < 7; move++ {
					value := moveset.Mset[piecei+56*start_pos.y+56*8*start_pos.x+moveLines*7+move]

					if value == 0 {
						break
					}

					end_pos := Location{}
					end_pos.fromByte(value)
					newBoard := c.Duplicate()
					if !newBoard.MakeMove(uint8(idx), team, end_pos, 'q') {
						continue
					}

					moves = append(moves, PossibleMove{piece: idx, end_pos: end_pos, c: newBoard})

				}
			}
		}
	}

	return moves
}

var evaluationPieceOffset = map[int]int{
	0:  64 * 3,
	1:  64 * 1,
	2:  64 * 2,
	3:  64 * 4,
	5:  64 * 2,
	4:  64 * 5,
	6:  64 * 1,
	7:  64 * 3,
	8:  0,
	9:  0,
	10: 0,
	11: 0,
	12: 0,
	13: 0,
	14: 0,
	15: 0,
}

/*
	Evaluate board value
*/
func (c *Chessboard) evaluate() float64 {
	total := 0.0
	wpiece := 0
	bpiece := 0
	for idx, loc := range c.white {
		if loc == 0 {
			continue
		}
		wpiece += 1
		total += float64(pieceValue[idx])
		if idx == 4 {
			continue
		}
		if idx > 15 {
			rune := c.whitePieceMap[idx][0]
			realidx := pieceIndexMap[rune][0]
			total += moveset.Pst[evaluationPieceOffset[realidx]+int((loc>>3)&0b111)+8*int((loc&0b111))]
		} else {
			total += moveset.Pst[evaluationPieceOffset[idx]+int((loc>>3)&0b111)+8*int((loc&0b111))]
		}

	}
	for idx, loc := range c.black {
		if loc == 0 {
			continue
		}
		bpiece += 1
		total -= float64(pieceValue[idx])
		if idx == 4 {
			continue
		}
		if idx > 15 {
			rune := c.blackPieceMap[idx][0]
			realidx := pieceIndexMap[rune][0]
			total += moveset.Pst[evaluationPieceOffset[realidx]+int((loc>>3)&0b111)+56-8*int((loc&0b111))]
		} else {
			total += moveset.Pst[evaluationPieceOffset[idx]+int((loc>>3)&0b111)+56-8*int((loc&0b111))]
		}
	}
	if bpiece <= 8 || wpiece <= 8 {
		if total < 0 {
			loc := c.white[4]
			total += moveset.Pst[evaluationPieceOffset[4]+int((loc>>3)&0b111)+8*int((loc&0b111))]

			locb := c.black[4]
			total -= math.Abs(float64((loc>>3)&0b111-(locb>>3)&0b111+(loc)&0b111-(locb)&0b111)) * 500

		} else {
			loc := c.black[4]
			total -= moveset.Pst[evaluationPieceOffset[4]+int((loc>>3)&0b111)+56-8*int((loc&0b111))]

			locb := c.white[4]
			total += math.Abs(float64((loc>>3)&0b111-(locb>>3)&0b111+(loc)&0b111-(locb)&0b111)) * 500
		}
	}
	return total
}

func (c *Chessboard) minimax(depth int, alfa float64, beta float64, team bool) (float64, PossibleMove) {
	if depth == 0 {
		return c.evaluate(), PossibleMove{}
	}

	// Check for stalemate
	pcw := 0
	pcb := 0

	for _, p := range c.white {
		if p != 0 {
			pcw++
		}
	}

	for _, p := range c.black {
		if p != 0 {
			pcb++
		}
	}
	if pcw == pcb && pcw == 1 {
		return 0, PossibleMove{}
	}

	if team {
		maxEval := math.Inf(-1)
		var maxEvalState PossibleMove
		pm := c.possibleMoves(team)
		if len(pm) == 0 {

			if !c.verifyState(team) {
				// White is checkmated
				return float64(-100000 * (depth + 1)), PossibleMove{}
			} else {
				// White has no legal moves
				return 0, PossibleMove{}
			}
		}

		for _, state := range pm {
			score, _ := state.c.minimax(depth-1, alfa, beta, !team)
			if score > maxEval {
				maxEval = score
				maxEvalState = state
			}
			alfa = math.Max(alfa, score)
			if beta <= alfa {
				break
			}
		}

		return maxEval, maxEvalState
	} else {
		minEval := math.Inf(+1)
		var minEvalState PossibleMove
		pm := c.possibleMoves(team)
		if len(pm) == 0 {

			if !c.verifyState(team) {
				// black is checkmated
				return float64(100000 * (depth + 1)), PossibleMove{}
			} else {
				// black has no legal moves
				return 0, PossibleMove{}
			}
		}

		for _, state := range pm {
			score, _ := state.c.minimax(depth-1, alfa, beta, !team)
			if score < minEval {
				minEval = score
				minEvalState = state
			}
			beta = math.Min(beta, score)
			if beta <= alfa {
				break
			}
		}

		return minEval, minEvalState
	}
}

var upgrader = websocket.Upgrader{} // use default options

// Game http handler
func echo(w http.ResponseWriter, r *http.Request) {

	// Upgrade conection to websocket
	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Print("upgrade:", err)
		return
	}
	defer c.Close()

	//Create chessboard
	board := Chessboard{}
	board.Init()

	/*for i := 0; i < 16; i++ {
		board.white[i] = 0
		board.black[i] = 0
	}

	board.bK = false
	board.wK = false
	board.bQ = false
	board.wQ = false

	board.white[3] = Piece(pgnToByte("c7"))
	//board.white[7] = Piece(pgnToByte("c8"))
	board.white[4] = Piece(pgnToByte("d7"))

	board.black[4] = Piece(pgnToByte("e4"))*/

	// Number of moves
	m := 0

	//Which team they are
	self := false
	player := true
	var total_time time.Duration
	for {

		mt, message, err := c.ReadMessage()
		if err != nil {
			log.Println("read:", err)
			break
		}
		botvalid := false
		if m != 0 {
			// Player
			move := strings.Split(string(message), "-")
			team := true
			pieceRune := strings.ToLower(string(move[0][1]))[0]
			piece := -1
			found_piece := false
			if move[0][0] == 'b' {
				team = false
			}
			if team == player {
				// player
				if player {
					for _, v := range pieceIndexMap[pieceRune] {
						if board.white[v] == Piece(pgnToByte(move[1])) {
							found_piece = true
							piece = v
						}
					}
				} else {
					for _, v := range pieceIndexMap[pieceRune] {
						if board.black[v] == Piece(pgnToByte(move[1])) {
							found_piece = true
							piece = v
						}
					}
				}
				valid := false
				if found_piece {
					log.Printf("Move piece %d %t to %s", piece, team, move[2])
					l := Location{}
					l.frompgn(move[2])
					valid = board.MakeMove(uint8(piece), team, l, 'q')
				}

				// Bot
				if valid {
					start := time.Now()
					score, botmove := board.minimax(4, math.Inf(-1), math.Inf(+1), self)
					log.Printf("Best Move with Score %f\n", score)
					botvalid = board.MakeMove(uint8(botmove.piece), self, botmove.end_pos, 'q')
					/*pm := board.possibleMoves(self)
					if len(pm) != 0 {
						pick := pm[rand.Intn(len(pm))]
						botvalid = board.MakeMove(uint8(pick.piece), self, pick.end_pos, 'q')
					}*/
					d := time.Since(start)
					log.Printf("Time: %s", d)
					total_time += d
				}

			}
		} else {
			// First message, set teams
			if string(message) == "black" {
				player = false
				self = true
			}

			// If bot is white, make the first move
			if self {
				start := time.Now()
				score, botmove := board.minimax(4, math.Inf(-1), math.Inf(+1), self)
				log.Printf("Best Move with Score %f\n", score)
				botvalid = board.MakeMove(uint8(botmove.piece), self, botmove.end_pos, 'q')
				/*pm := board.possibleMoves(self)
				if len(pm) != 0 {
					pick := pm[rand.Intn(len(pm))]
					botvalid = board.MakeMove(uint8(pick.piece), self, pick.end_pos, 'q')
				}*/
				d := time.Since(start)
				log.Printf("Time: %s", d)
				total_time += d

			}
		}
		m += 1
		board.rounds = m
		log.Printf("recv: %s", message)

		// Test for checkmate or draw
		if botvalid {
			if len(board.possibleMoves(player)) == 0 {

				if !board.verifyState(player) {
					err = c.WriteMessage(mt, []byte("Checkmate"))
				} else {
					err = c.WriteMessage(mt, []byte("Draw"))
				}
				log.Printf("Total Bot Time: %d", total_time)
			}
		} else {
			if len(board.possibleMoves(self)) == 0 {
				if !board.verifyState(self) {
					err = c.WriteMessage(mt, []byte("Checkmate"))
				} else {
					err = c.WriteMessage(mt, []byte("Draw"))
				}
				log.Printf("Total Bot Time: %d", total_time)

			}
		}
		pcw := 0
		pcb := 0

		for _, p := range board.white {
			if p != 0 {
				pcw++
			}
		}

		for _, p := range board.black {
			if p != 0 {
				pcb++
			}
		}

		// Test for stalemate
		if pcw == pcb && pcw == 1 {
			err = c.WriteMessage(mt, []byte("stalemate"))
		}

		err = c.WriteMessage(mt, []byte(board.fen()))
		if err != nil {
			log.Println("write:", err)
			break
		}
	}
}

var addr = flag.String("addr", "localhost:8080", "http service address")

func main() {

	upgrader.CheckOrigin = func(r *http.Request) bool {
		return true
	}
	flag.Parse()
	log.SetFlags(0)
	http.HandleFunc("/echo", echo)
	http.ListenAndServe(*addr, nil)
}
