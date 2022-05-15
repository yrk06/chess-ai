package main

import (
	"bufio"
	"encoding/binary"
	"flag"
	"fmt"
	"log"
	"math"
	"math/rand"
	"net/http"
	"os"
	"runtime/pprof"
	"strings"
	"time"
	"yrk06/chess-backend/moveset"

	"github.com/gorilla/websocket"
)

// Minimax initial depth
var BOT_MINIMAX_DEPTH = 5

// Random movement bot control
const BOT_RANDOM_CHANCE = 200
const BOT_POINT_RANDOM_THRESHOLD = 10
const AI_GAME_DELAY = 1000000000 * 0.0

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

// piece type to movement memory offset map  Tabela de offset de memoria (movimentos) pra cada peça
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
	'K': 25088,
}

// Location Type
type Location struct {
	x int
	y int
}

// Converts a location object to FEN string index
func (l *Location) toFen() int {
	return (7-l.y)*9 + (l.x)
}

// Converts a location object to PGN
func (l *Location) pgn() string {
	return fmt.Sprintf("%s%d", string(l.x+97), l.y+1)
}

// Loads a position from a PGN
func (l *Location) frompgn(pos string) {
	l.x = int(pos[0]) - 97
	l.y = int(pos[1]-'0') - 1
}

// Converts a location to Byte
func (l *Location) toByte() uint8 {
	return uint8((1 << 7) | ((l.x & 0b111) << 3) | (l.y & 0b111))
}

// Loads a location from a byte
func (l *Location) fromByte(b uint8) {
	l.x = int((b >> 3) & 0b111)
	l.y = int((b) & 0b111)
}

// Transforms a PGN to a byte
func pgnToByte(pgn string) uint8 {
	l := Location{}
	l.frompgn(pgn)
	return l.toByte()
}

// Holds a possible move and changes to board state
type PossibleMove struct {
	score float64

	piece   int
	end_pos Location

	castle   bool
	spiece   int
	send_pos Location

	enpassant uint8

	target uint8

	wK bool
	wQ bool
	bK bool
	bQ bool

	promote    bool
	promote_to byte

	invalid bool
}

type Piece uint8

// Player piece array. Holds 16 regular pieces and 8 aditional pieces for promotion
type PlayerPieces [24]Piece

/*
	Piece index to char representation map
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
	Char representation to possible Piece index map
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

	toMove bool

	wK bool
	wQ bool
	bK bool
	bQ bool

	enpassant uint8

	mc     int
	rounds int

	plays map[int]int
}

/*
	Starts the chessboard at the start location
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
	c.whitePieceMap = make(map[int]string, 8)
	c.blackPieceMap = make(map[int]string, 8)

	c.plays = make(map[int]int)

	// Init white pieces
	for y := 0; y < 2; y++ {
		for x := 0; x < 8; x++ {
			l := Location{x: x, y: y}
			c.white[y*8+x] = Piece(l.toByte())
		}
	}

	// Init black pieces
	for y := 0; y < 2; y++ {
		for x := 0; x < 8; x++ {
			l := Location{x: x, y: 7 - y}
			c.black[y*8+x] = Piece(l.toByte())
		}
	}

}

// Zobrist hash table data
var zhtable [64][18]int64
var zh_black_to_move int64
var zh_depth [10]int64

// Init the Zobrist hash table used for hasing
func init_zhtable() {
	rand.Seed(190520024679183427)

	for idx := range zhtable {
		for idx2 := range zhtable[idx] {
			zhtable[idx][idx2] = rand.Int63()
		}
	}
	zh_black_to_move = rand.Int63()

	for idx := range zh_depth {
		zh_depth[idx] = rand.Int63()
	}
}

// Hashes a Board into an int64
func (c *Chessboard) zobristHash() int64 {
	var h int64
	if !c.toMove {
		h = h ^ zh_black_to_move
	}
	for idx, pos := range c.white {
		if pos == 0 {
			continue
		}
		index := 0
		if idx > 15 {
			index = pieceIndexMap[c.whitePieceMap[idx][0]][0]
		} else {
			index = pieceIndexMap[indexPieceMap[idx][0]][0]
		}

		h = h ^ zhtable[(pos>>3)&0b111+(pos&0b111*8)][index]

	}

	for idx, pos := range c.black {
		if pos == 0 {
			continue
		}
		index := 0
		if idx > 15 {
			index = pieceIndexMap[c.blackPieceMap[idx][0]][0]
		} else {
			index = pieceIndexMap[indexPieceMap[idx][0]][0]
		}

		h = h ^ zhtable[(pos>>3)&0b111+(pos&0b111*8)][index+8]

	}
	return h
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

// Load a board from a FEN string
func (c *Chessboard) fromFen(fen string) {
	c.plays = make(map[int]int)
	c.whitePieceMap = make(map[int]string, 8)
	c.blackPieceMap = make(map[int]string, 8)
	cchar := 0
	for i := 0; i < 64; i++ {
		char := fen[cchar]
		if char > 'a' && char < 'z' {
			for _, v := range pieceIndexMap[char] {
				if c.black[v] == 0 {
					if v >= 16 {
						c.blackPieceMap[v] = string(char)
					}
					l := Location{
						x: i % 8,
						y: 7 - i/8,
					}
					c.black[v] = Piece(l.toByte())
					break
				}

			}
		} else if char > 'A' && char < 'Z' {
			lcchar := strings.ToLower(string(char))[0]
			for _, v := range pieceIndexMap[lcchar] {
				if c.white[v] == 0 {
					if v >= 16 {
						c.whitePieceMap[v] = string(lcchar)
					}
					l := Location{
						x: i % 8,
						y: 7 - i/8,
					}
					c.white[v] = Piece(l.toByte())
					break
				}
			}
		} else if char == '/' {
			i--
		} else {
			jumps := int(char - '0')
			i += jumps - 1
		}
		cchar++
	}

	cchar += 2
	if fen[cchar] == 'w' {
		c.toMove = true
	}
	cchar += 2
	for ; ; cchar++ {
		char := fen[cchar]
		if char == '-' {
			cchar += 2
			break
		} else if char == ' ' {
			cchar += 1
			break
		}

		if char == 'q' {
			c.bQ = true
		}
		if char == 'k' {
			c.bK = true
		}
		if char == 'Q' {
			c.wQ = true
		}
		if char == 'Q' {
			c.wK = true
		}

	}
	if fen[cchar] != '-' {
		str := ""
		str += string(fen[cchar])
		cchar++
		str += string(fen[cchar])

		c.enpassant = pgnToByte(str)
	}
}

/*
	Check if there are any pieces in the posisiton. Takes into account en passant (ghost piece).
	bit 5 is 1 if piece is white 0 otherwise
*/
func (c *Chessboard) hasPieceInPosition(position uint8, enpassant bool) (bool, int) {

	// Start with calculating En passant
	if enpassant {
		if position == c.enpassant {
			hp, piece := c.hasPieceInPosition(position+1, false)
			if !hp {
				_, piece = c.hasPieceInPosition(position-1, false)
			}
			return true, piece
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
			piecei = pieceMap[p]
		} else {
			p := indexPieceMap[int(piece)][0]
			if p == 'p' {
				p = 'P'
			}
			if p == 'k' {
				p = 'K'
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
func (c *Chessboard) MakeMove(piece uint8, team bool, end_pos Location, promote_to byte) (bool, uint8) {

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
			piecei = pieceMap[p]
		} else {
			p := indexPieceMap[int(piece)][0]
			if p == 'p' {
				p = 'P'
			}
			if p == 'k' {
				p = 'K'
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
						return false, 0
					}
					sq1 := Location{x: 5, y: 0}
					if c, _ := c.hasPieceInPosition(sq1.toByte(), false); c {
						return false, 0
					}
					if c.isSquareAttacked(!team, sq1) {
						return false, 0
					}

					sq2 := Location{x: 6, y: 0}
					if c, _ := c.hasPieceInPosition(sq2.toByte(), false); c {
						return false, 0
					}
					if c.isSquareAttacked(!team, sq2) {
						return false, 0
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
						return false, 0
					}
					sq1 := Location{x: 1, y: 0}
					if c, _ := c.hasPieceInPosition(sq1.toByte(), false); c {
						return false, 0
					}

					sq2 := Location{x: 2, y: 0}
					if c, _ := c.hasPieceInPosition(sq2.toByte(), false); c {
						return false, 0
					}
					if c.isSquareAttacked(!team, sq2) {
						return false, 0
					}

					sq3 := Location{x: 3, y: 0}
					if c, _ := c.hasPieceInPosition(sq3.toByte(), false); c {
						return false, 0
					}
					if c.isSquareAttacked(!team, sq3) {
						return false, 0
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
						return false, 0
					}
					sq1 := Location{x: 5, y: 7}
					if c, _ := c.hasPieceInPosition(sq1.toByte(), false); c {
						return false, 0
					}
					if c.isSquareAttacked(!team, sq1) {
						return false, 0
					}

					sq2 := Location{x: 6, y: 7}
					if c, _ := c.hasPieceInPosition(sq2.toByte(), false); c {
						return false, 0
					}
					if c.isSquareAttacked(!team, sq2) {
						return false, 0
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
						return false, 0
					}
					sq1 := Location{x: 1, y: 7}
					if c, _ := c.hasPieceInPosition(sq1.toByte(), false); c {
						return false, 0
					}

					sq2 := Location{x: 2, y: 7}
					if c, _ := c.hasPieceInPosition(sq2.toByte(), false); c {
						return false, 0
					}
					if c.isSquareAttacked(!team, sq2) {
						return false, 0
					}

					sq3 := Location{x: 3, y: 7}
					if c, _ := c.hasPieceInPosition(sq3.toByte(), false); c {
						return false, 0
					}
					if c.isSquareAttacked(!team, sq3) {
						return false, 0
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
		return true, 0
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

					}

					if end_pos.y == 0 && !team {
						promotion = true

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
		return false, 0
	}

	// Check if piece can capture target (if there is a target)
	target, target_p := c.hasPieceInPosition(ep, enpassant)
	if target {
		pieceTeam := target_p >> 5
		// Cannot capture your own pieces
		if pieceTeam == 1 && team {
			return false, 0
		}
		if pieceTeam == 0 && !team {
			return false, 0
		}
		if !canAttack {
			return false, 0
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
		return false, 0
	}

	// Remove checking rights
	if piece == 4 {
		if team {
			c.wK = false
			c.wQ = false
		} else {
			c.bK = false
			c.bQ = false
		}
	}

	truetargetp := 0
	if target {
		truetargetp = target_p | 1<<5
	}

	c.toMove = !c.toMove
	return true, uint8(truetargetp)
}

/*
	Attemps to move PIECE and returns a possible move struct
*/
func (c *Chessboard) TestMove(piece uint8, team bool, end_pos Location, promote_to byte) (bool, PossibleMove) {

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
			piecei = pieceMap[p]
		} else {
			p := indexPieceMap[int(piece)][0]
			if p == 'p' {
				p = 'P'
			}
			if p == 'k' {
				p = 'K'
			}
			piecei = pieceMap[p]
		}

	}

	// Control variables
	predicted_score := 0.0
	valid := false
	enpassant := false
	enpassant_active := false
	canAttack := true
	promotion := false
	promotion_idx := 0

	// Check for castling
	castling := false
	cq := false
	cep := Location{}
	if piece == 4 {
		if team {

			if end_pos == (Location{x: 6, y: 0}) {
				if c.wK {
					if c.isSquareAttacked(!team, start_pos) {
						return false, PossibleMove{invalid: true}
					}
					sq1 := Location{x: 5, y: 0}
					if c, _ := c.hasPieceInPosition(sq1.toByte(), false); c {
						return false, PossibleMove{invalid: true}
					}
					if c.isSquareAttacked(!team, sq1) {
						return false, PossibleMove{invalid: true}
					}

					sq2 := Location{x: 6, y: 0}
					if c, _ := c.hasPieceInPosition(sq2.toByte(), false); c {
						return false, PossibleMove{invalid: true}
					}
					if c.isSquareAttacked(!team, sq2) {
						return false, PossibleMove{invalid: true}
					}

					c.white[4] = Piece(end_pos.toByte())
					c.white[7] = Piece(sq1.toByte())
					cep = sq1
					castling = true
					c.wK = false
					c.wQ = false

				} else {
					return false, PossibleMove{invalid: true}
				}

			}

			if end_pos == (Location{x: 2, y: 0}) {
				if c.wQ {
					if c.isSquareAttacked(!team, start_pos) {
						return false, PossibleMove{invalid: true}
					}
					sq1 := Location{x: 1, y: 0}
					if c, _ := c.hasPieceInPosition(sq1.toByte(), false); c {
						return false, PossibleMove{invalid: true}
					}

					sq2 := Location{x: 2, y: 0}
					if c, _ := c.hasPieceInPosition(sq2.toByte(), false); c {
						return false, PossibleMove{invalid: true}
					}
					if c.isSquareAttacked(!team, sq2) {
						return false, PossibleMove{invalid: true}
					}

					sq3 := Location{x: 3, y: 0}
					if c, _ := c.hasPieceInPosition(sq3.toByte(), false); c {
						return false, PossibleMove{invalid: true}
					}
					if c.isSquareAttacked(!team, sq3) {
						return false, PossibleMove{invalid: true}
					}

					c.white[4] = Piece(end_pos.toByte())
					c.white[0] = Piece(sq3.toByte())
					cep = sq3
					castling = true
					c.wK = false
					c.wQ = false
					cq = true

				} else {
					return false, PossibleMove{invalid: true}
				}
			}

		} else {

			if end_pos == (Location{x: 6, y: 7}) {
				if c.bK {
					if c.isSquareAttacked(!team, start_pos) {
						return false, PossibleMove{invalid: true}
					}
					sq1 := Location{x: 5, y: 7}
					if c, _ := c.hasPieceInPosition(sq1.toByte(), false); c {
						return false, PossibleMove{invalid: true}
					}
					if c.isSquareAttacked(!team, sq1) {
						return false, PossibleMove{invalid: true}
					}

					sq2 := Location{x: 6, y: 7}
					if c, _ := c.hasPieceInPosition(sq2.toByte(), false); c {
						return false, PossibleMove{invalid: true}
					}
					if c.isSquareAttacked(!team, sq2) {
						return false, PossibleMove{invalid: true}
					}

					c.black[4] = Piece(end_pos.toByte())
					c.black[7] = Piece(sq1.toByte())
					cep = sq1
					castling = true
					c.bK = false
					c.bQ = false
				}
			}

			if end_pos == (Location{x: 2, y: 7}) {
				if c.bQ {
					if c.isSquareAttacked(!team, start_pos) {
						return false, PossibleMove{invalid: true}
					}
					sq1 := Location{x: 1, y: 7}
					if c, _ := c.hasPieceInPosition(sq1.toByte(), false); c {
						return false, PossibleMove{invalid: true}
					}

					sq2 := Location{x: 2, y: 7}
					if c, _ := c.hasPieceInPosition(sq2.toByte(), false); c {
						return false, PossibleMove{invalid: true}
					}
					if c.isSquareAttacked(!team, sq2) {
						return false, PossibleMove{invalid: true}
					}

					sq3 := Location{x: 3, y: 7}
					if c, _ := c.hasPieceInPosition(sq3.toByte(), false); c {
						return false, PossibleMove{invalid: true}
					}
					if c.isSquareAttacked(!team, sq3) {
						return false, PossibleMove{invalid: true}
					}

					c.black[4] = Piece(end_pos.toByte())
					c.black[0] = Piece(sq3.toByte())
					cep = sq3
					castling = true
					c.bK = false
					c.bQ = false
					cq = true
				}
			}
		}
	}
	if castling {
		c.toMove = !c.toMove
		tower := 7
		if cq {
			tower = 0
		}
		predicted_score += 3
		return true, PossibleMove{castle: true, score: predicted_score, piece: int(piece), end_pos: end_pos, spiece: tower, send_pos: cep, wK: c.wK, wQ: c.wQ, bK: c.bK, bQ: c.bQ}
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

					}

					if end_pos.y == 0 && !team {
						promotion = true

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
		return false, PossibleMove{invalid: true}
	}

	// Check if piece can capture target (if there is a target)
	target, target_p := c.hasPieceInPosition(ep, enpassant)
	if target {
		pieceTeam := target_p >> 5
		// Cannot capture your own pieces
		if pieceTeam == 1 && team {
			return false, PossibleMove{invalid: true}
		}
		if pieceTeam == 0 && !team {
			return false, PossibleMove{invalid: true}
		}
		if !canAttack {
			return false, PossibleMove{invalid: true}
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
		predicted_score += 5
		if piece >= 8 && piece < 16 {
			predicted_score += 5
		}
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
		predicted_score += 10
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
		return false, PossibleMove{invalid: true}
	}

	// Remove castling rights
	if piece == 4 {
		predicted_score -= 3
		if team {
			c.wK = false
			c.wQ = false
		} else {
			c.bK = false
			c.bQ = false
		}
	}

	c.toMove = !c.toMove

	truetargetp := 0
	if target {
		truetargetp = target_p | 1<<5
	}

	return true, PossibleMove{score: predicted_score, piece: int(piece), end_pos: end_pos,
		wK: c.wK, wQ: c.wQ, bK: c.bK, bQ: c.bQ,
		enpassant: c.enpassant,
		target:    uint8(truetargetp),
		promote:   promotion, promote_to: promote_to}
}

/*
	Makes a move from a PossibleMove (no validity checks are made)
*/
func (c *Chessboard) MakeUnsafeMove(pm PossibleMove, team bool) {

	// Make the Move and check if king is in check
	if team {
		c.white[pm.piece] = Piece(pm.end_pos.toByte())
		if pm.castle {
			c.white[pm.spiece] = Piece(pm.send_pos.toByte())
		}
	} else {
		c.black[pm.piece] = Piece(pm.end_pos.toByte())
		if pm.castle {
			c.black[pm.spiece] = Piece(pm.send_pos.toByte())
		}
	}
	c.bQ = pm.bQ
	c.bK = pm.bK

	c.wQ = pm.wQ
	c.wK = pm.wK

	// Capture piece
	if pm.target != 0 {
		if team {
			c.black[pm.target&0x1F] = 0
		} else {
			c.white[pm.target&0x1F] = 0
		}
	}

	c.enpassant = pm.enpassant

	if pm.promote {
		if team {
			for i := 16; i < 24; i++ {
				if c.white[i] == 0 {
					c.whitePieceMap[i] = string(pm.promote_to)
					c.white[i] = c.white[pm.piece]
					c.white[pm.piece] = 0
					break
				}
			}
		} else {
			for i := 16; i < 24; i++ {
				if c.black[i] == 0 {
					c.blackPieceMap[i] = string(pm.promote_to)
					c.black[i] = c.black[pm.piece]
					c.black[pm.piece] = 0
					break
				}
			}
		}

	}
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

	board.plays = make(map[int]int)
	for k, v := range c.plays {
		board.plays[k] = v
	}

	return board
}

/*
	Saves a state (does not create a new object in memory)
*/
func (c *Chessboard) SaveState(target *Chessboard) {

	target.white = c.white
	target.black = c.black
	target.toMove = c.toMove
	target.wK = c.wK
	target.wQ = c.wQ
	target.bK = c.bK
	target.bQ = c.bQ
	target.enpassant = c.enpassant
	target.mc = c.mc
	target.rounds = c.rounds

	target.whitePieceMap = make(map[int]string, 8)
	target.blackPieceMap = make(map[int]string, 8)
	target.plays = make(map[int]int)

	for k, v := range c.blackPieceMap {
		target.blackPieceMap[k] = v
	}

	for k, v := range c.whitePieceMap {
		target.whitePieceMap[k] = v
	}

	for k, v := range c.plays {
		target.plays[k] = v
	}
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

// Calculate all possible moves for a team
func (c *Chessboard) possibleMoves(team bool) []PossibleMove {
	moves := make([]PossibleMove, 0)

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
		return moves
	}

	var board Chessboard
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
					//newBoard := c.Duplicate()
					c.SaveState(&board)

					var tg PossibleMove
					var co bool
					if co, tg = board.TestMove(uint8(idx), team, end_pos, 'q'); !co {
						continue
					}

					if tg.invalid {
						log.Panicf("tg invalid at %d %t %s", idx, team, end_pos.pgn())
					}

					moves = append(moves, tg)

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
					//newBoard := c.Duplicate()
					c.SaveState(&board)

					var tg PossibleMove
					var co bool
					if co, tg = board.TestMove(uint8(idx), team, end_pos, 'q'); !co {
						continue
					}

					if tg.invalid {
						log.Panicf("tg invalid at %d %t %s", idx, team, end_pos.pgn())
					}

					moves = append(moves, tg)

				}
			}
		}
	}

	/*sort.SliceStable(moves, func(i, j int) bool {
		return moves[i].score > moves[i].score
	})*/
	return moves
}

// Maps from Piece index to Evaluation memory offset
var evaluationPieceOffset = map[int]int{
	0:  64 * 3,
	1:  64 * 1,
	2:  64 * 2,
	3:  64 * 4,
	4:  64 * 5,
	5:  64 * 2,
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
			total -= moveset.Pst[evaluationPieceOffset[realidx]+int((loc>>3)&0b111)+56-8*int((loc&0b111))]
		} else {
			total -= moveset.Pst[evaluationPieceOffset[idx]+int((loc>>3)&0b111)+56-8*int((loc&0b111))]
		}
	}
	if bpiece <= 5 || wpiece <= 5 {
		if total < 0 {
			loc := c.white[4]
			total += moveset.Pst[evaluationPieceOffset[4]+int((loc>>3)&0b111)+8*int((loc&0b111))]

			locb := c.black[4]
			total -= math.Abs(float64((loc>>3)&0b111-(locb>>3)&0b111+(loc)&0b111-(locb)&0b111)) * math.Min(5-float64(wpiece), 0) * 85

		} else if total > 0 {
			loc := c.black[4]
			total -= moveset.Pst[evaluationPieceOffset[4]+int((loc>>3)&0b111)+56-8*int((loc&0b111))]

			locb := c.white[4]
			total += math.Abs(float64((loc>>3)&0b111-(locb>>3)&0b111+(loc)&0b111-(locb)&0b111)) * math.Min(5-float64(bpiece), 0) * 85
		}
	} else {
		loc := c.white[4]
		total += moveset.Pst[evaluationPieceOffset[4]+64+int((loc>>3)&0b111)+8*int((loc&0b111))]

		loc = c.black[4]
		total -= moveset.Pst[evaluationPieceOffset[4]+64+int((loc>>3)&0b111)+56-8*int((loc&0b111))]
	}
	return total
}

// Transposition Table
var transposition_table = make(map[int64]TranspositionEntry)

type TranspositionEntry struct {
	score float64
	depth int
}

//var TP_mutex sync.RWMutex

// Save transposition data to memory
func save_transposition_table(filename string) {
	/*TP_mutex.RLock()
	defer TP_mutex.RUnlock()*/
	f, err := os.Create(filename)
	if err != nil {
		log.Panic(err)
	}
	defer f.Close()
	for key, val := range transposition_table {

		keybuf := make([]byte, 8)
		binary.LittleEndian.PutUint64(keybuf, uint64(key))
		bits := math.Float64bits(val.score)
		valbuf := make([]byte, 8)
		binary.LittleEndian.PutUint64(valbuf, uint64(bits))

		valbuf2 := make([]byte, 2)
		binary.LittleEndian.PutUint16(valbuf2, uint16(val.depth))

		f.Write([]byte(fmt.Sprintf("%s/+++++++/%s/+++++++/%s\n", keybuf, valbuf, valbuf2)))
	}
}

func Readln(r *bufio.Reader) (string, error) {
	var (
		isPrefix bool  = true
		err      error = nil
		line, ln []byte
	)
	for isPrefix && err == nil {
		line, isPrefix, err = r.ReadLine()
		ln = append(ln, line...)
	}
	return string(ln), err
}

// Loads transposition table from file
func load_transposition_table(filename string) {
	/*TP_mutex.Lock()
	defer TP_mutex.Unlock()*/
	f, err := os.Open(filename)
	if err != nil {
		log.Println("Could not open file")
		return
	}
	defer f.Close()
	r := bufio.NewReader(f)
	s, e := r.ReadBytes('\n')
	for e == nil && len(s) > 0 {

		str := string(s)
		t := strings.Split(str, "/+++++++/")
		if len(t) < 3 {
			continue
		}
		log.Println(t[0])
		log.Println(t[1])
		log.Println(t[2])

		v := binary.LittleEndian.Uint64([]byte(t[0]))
		score := math.Float64frombits(binary.LittleEndian.Uint64([]byte(t[1])))
		depth := binary.LittleEndian.Uint16([]byte(t[2]))

		transposition_table[int64(v)] = TranspositionEntry{score: score, depth: int(depth)}

		s, e = r.ReadBytes('\n')
	}
}

// Calculate the best movement for a TEAM
func (c *Chessboard) minimax(depth int, alfa float64, beta float64, team bool, num_states *int) (float64, PossibleMove) {
	if depth == 0 {
		*num_states += 1
		return c.evaluate(), PossibleMove{invalid: true}
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
		return 0, PossibleMove{invalid: true}
	}

	zh := c.zobristHash()

	if *num_states != 0 {
		if c.plays[int(zh)] >= 3 {
			if team {
				return 0.01, PossibleMove{invalid: true}
			} else {
				return 0.01, PossibleMove{invalid: true}
			}

		}
		//TP_mutex.RLock()
		if val, ok := transposition_table[zh]; ok {
			if val.depth >= depth {
				//TP_mutex.RUnlock()
				*num_states += 1
				return val.score, PossibleMove{}
			}
		}
		//TP_mutex.RUnlock()

	}

	var board Chessboard
	if team {
		maxEval := math.Inf(-1)
		var maxEvalState PossibleMove
		pm := c.possibleMoves(team)
		if len(pm) == 0 {

			if !c.verifyState(team) {
				// White is checkmated
				return float64(-100000 * (depth + 1)), PossibleMove{invalid: true}
			} else {
				// White has no legal moves
				return 0, PossibleMove{invalid: true}
			}
		}

		for _, state := range pm {
			*num_states += 1
			c.SaveState(&board)
			board.MakeUnsafeMove(state, team)
			score, _ := board.minimax(depth-1, alfa, beta, !team, num_states)
			if score == math.Inf(-1) {
				maxEval = score
				maxEvalState = state
			} else if score > 0 && score < 0.1 {
				// Just ignore it
			} else if score > maxEval {
				maxEval = score
				maxEvalState = state
				// 10% of chance of getting a similar move (2 pawns of diference)
			} else if math.Abs(maxEval-score) < BOT_POINT_RANDOM_THRESHOLD && rand.Intn(BOT_RANDOM_CHANCE) == 1 {
				maxEval = score
				maxEvalState = state
			}
			alfa = math.Max(alfa, score)
			if beta <= alfa {
				break
			}
		}

		//TP_mutex.Lock()
		transposition_table[zh] = TranspositionEntry{score: maxEval, depth: depth}
		//TP_mutex.Unlock()
		return maxEval, maxEvalState
	} else {
		minEval := math.Inf(+1)
		var minEvalState PossibleMove
		pm := c.possibleMoves(team)
		if len(pm) == 0 {

			if !c.verifyState(team) {
				// black is checkmated
				return float64(100000 * (depth + 1)), PossibleMove{invalid: true}
			} else {
				// black has no legal moves
				return 0, PossibleMove{invalid: true}
			}
		}

		for _, state := range pm {
			*num_states += 1
			c.SaveState(&board)
			board.MakeUnsafeMove(state, team)
			score, _ := board.minimax(depth-1, alfa, beta, !team, num_states)

			if score == math.Inf(+1) {
				minEval = score
				minEvalState = state
			} else if score > 0 && score < 0.1 {
				// Just ignore it
			} else if score < minEval {
				minEval = score
				minEvalState = state
			} else if math.Abs(minEval-score) < BOT_POINT_RANDOM_THRESHOLD && rand.Intn(BOT_RANDOM_CHANCE) == 1 {
				minEval = score
				minEvalState = state
			}
			beta = math.Min(beta, score)
			if beta <= alfa {
				break
			}
		}

		//TP_mutex.Lock()
		transposition_table[zh] = TranspositionEntry{score: minEval, depth: depth}
		//TP_mutex.Unlock()
		return minEval, minEvalState
	}
}


func (c *Chessboard) calculateAllMovements(depth int, layer bool) int {
	pm1 := c.possibleMoves(layer)

	var Board Chessboard
	if depth > 1 {
		total := 0
		for _, p := range pm1 {

			c.SaveState(&Board)
			Board.MakeUnsafeMove(p, layer)
			total += Board.calculateAllMovements(depth-1, !layer)
		}
		return total
	} else {
		/*for _, p := range pm1 {
			c.SaveState(&Board)
			Board.MakeUnsafeMove(p, layer)
			log.Printf(Board.fen())
		}*/
		total := len(pm1)
		return total
	}

}

var upgrader = websocket.Upgrader{} // use default options

var startpos = flag.String("startpos", "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1", "FEN for starting position")

// Game http handler
func echo(w http.ResponseWriter, r *http.Request) {

	f, err := os.Create("Game.prof")
	if err != nil {
		log.Fatal(err)
	}
	pprof.StartCPUProfile(f)
	defer pprof.StopCPUProfile()
	// Upgrade conection to websocket
	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Print("upgrade:", err)
		return
	}
	defer c.Close()

	//Create chessboard
	board := Chessboard{}
	board.fromFen(*startpos)
	//board.Init()
	//board.fromFen("r1b111k1/1pp111p1/p1111p1p/P11n1111/11PpR1P1/1111111P/1P1K1P11/R1111111 b - c3 0 1")

	// Number of moves
	m := 0

	//Which team they are
	self := false
	player := true

	end_game := false

	bot_local_minimax_depth := BOT_MINIMAX_DEPTH

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
					l := Location{}
					l.frompgn(move[2])

					rank := string(((board.white[piece] >> 3) & 0b111) + 97)
					var capture uint8
					if !player {
						rank = string(((board.black[piece] >> 3) & 0b111) + 97)
					}
					valid, capture = board.MakeMove(uint8(piece), team, l, 'q')

					if valid {

						c.WriteMessage(mt, []byte(board.fen()))
						pgn_piece := string(indexPieceMap[piece])

						if piece > 15 {
							if player {
								pgn_piece = string(board.whitePieceMap[piece])
							} else {
								pgn_piece = string(board.blackPieceMap[piece])
							}
						}

						capturestr := ""
						if capture != 0 {
							capturestr = "x"
						}
						if pgn_piece == "p" {
							pgn_piece = ""
							if capture == 0 {
								rank = ""

							}
						}

						log.Printf("%s%s%s%s ", strings.ToUpper(pgn_piece), rank, capturestr, l.pgn())

						if !player {
							m += 1
							log.Printf("%d. ", m+1)
						}

					}

				}

				// Bot
				if valid {
					board.plays[int(board.zobristHash())] += 1
					start := time.Now()
					states_analized := 0
					score, botmove := board.minimax(bot_local_minimax_depth, math.Inf(-1), math.Inf(+1), self, &states_analized)
					if math.Abs(score) >= 100000 {
						depth_found := (math.Abs(score) / 100000) - 1
						log.Printf("Best Move M%d\n", bot_local_minimax_depth-int(depth_found))
					} else {
						log.Printf("Best Move with Score %f\n", score)
					}

					rank := string(((board.white[botmove.piece] >> 3) & 0b111) + 97)
					capture := botmove.target != 0
					if !self {
						rank = string(((board.black[botmove.piece] >> 3) & 0b111) + 97)
					}
					botvalid := !botmove.invalid
					if botvalid {
						board.MakeUnsafeMove(botmove, self)

						d := time.Since(start)

						log.Printf("States %d, %fs, Mean %f states/second", states_analized, d.Seconds(), float64(states_analized)/d.Seconds())
						total_time += d

						if end_game {
							if d.Seconds() < 15 && bot_local_minimax_depth < 8 {
								bot_local_minimax_depth += 1
								log.Printf("Deepening search to %d", bot_local_minimax_depth)
							} else if d.Seconds() < 10 && bot_local_minimax_depth < 11 {
								bot_local_minimax_depth += 1
								log.Printf("Deepening search to %d", bot_local_minimax_depth)
							} else if d.Seconds() > 210 {
								bot_local_minimax_depth -= 1
								log.Printf("Shallowing search to %d", bot_local_minimax_depth)
							}
						}

						pgn_piece := string(indexPieceMap[botmove.piece])

						if botmove.piece > 15 {
							if self {
								pgn_piece = string(board.whitePieceMap[botmove.piece])
							} else {
								pgn_piece = string(board.blackPieceMap[botmove.piece])
							}
						}

						if pgn_piece == "p" {
							pgn_piece = ""
							if !capture {
								rank = ""
							}
						}
						log.Printf("%s%s%s%s ", strings.ToUpper(pgn_piece), rank, map[bool]string{true: "x", false: ""}[capture], botmove.end_pos.pgn())

						if !self {
							m += 1
							log.Printf("%d. ", m)
						}
					}

				}

			}
		} else {
			// First message, set teams
			if string(message) == "black" {
				player = false
				self = true
			}
			m += 1
			// If bot is white, make the first move
			if self {
				start := time.Now()
				states := 0
				score, botmove := board.minimax(BOT_MINIMAX_DEPTH, math.Inf(-1), math.Inf(+1), self, &states)
				log.Printf("Best Move with Score %f\n", score)

				rank := string(((board.white[botmove.piece] >> 3) & 0b111) + 97)
				var capture uint8
				if !self {
					rank = string(((board.black[botmove.piece] >> 3) & 0b111) + 97)
				}
				botvalid, capture = board.MakeMove(uint8(botmove.piece), self, botmove.end_pos, 'q')

				d := time.Since(start)
				total_time += d

				if botvalid {
					pgn_piece := string(indexPieceMap[botmove.piece])

					if botmove.piece > 15 {
						if self {
							pgn_piece = string(board.whitePieceMap[botmove.piece])
						} else {
							pgn_piece = string(board.blackPieceMap[botmove.piece])
						}
					}

					capturestr := ""
					if capture != 0 {
						capturestr = "x"
					}
					if pgn_piece == "p" {
						pgn_piece = ""
						if capture == 0 {
							rank = ""
						}
					}
					log.Printf("%s%s%s%s ", strings.ToUpper(pgn_piece), rank, capturestr, botmove.end_pos.pgn())

					if !self {
						m += 1
						log.Printf("%d. ", m)
					}

				}

			}
		}

		board.rounds = m

		// Test for checkmate or draw
		if botvalid {

			board.plays[int(board.zobristHash())] += 1

		}

		if board.toMove == player {
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

		if pcw+pcb < 15 {
			end_game = true
		}

		err = c.WriteMessage(mt, []byte(board.fen()))
		c.WriteMessage(mt, []byte(fmt.Sprintf("eval %.5f", board.evaluate())))
		if err != nil {
			log.Println("write:", err)
			break
		}
		if m%2 == 0 && m != 0 {
			save_transposition_table("tpmemory3.tp")
		}
	}
	save_transposition_table("tpmemory3.tp")
}


func ai(w http.ResponseWriter, r *http.Request) {

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

	// Number of moves
	m := 0

	//Which team they are
	self := false
	player := true
	var total_time time.Duration
	var mt int
	for {

		botvalid := false
		if m == 0 {
			mt, _, _ = c.ReadMessage()
			if err != nil {
				log.Println("read:", err)
				break
			}
		} else {
			pm := board.possibleMoves(player)
			valid := false
			states := 0
			score, botmove := board.minimax(BOT_MINIMAX_DEPTH, math.Inf(-1), math.Inf(+1), player, &states)
			if math.Abs(score) < 0.1 && math.Abs(score) > 0 {
				botmove = pm[rand.Intn(len(pm))]
			}
			//log.Printf("Best Move with Score %f\n", score)
			rank := string(((board.white[botmove.piece] >> 3) & 0b111) + 97)
			var capture uint8
			valid, capture = board.MakeMove(uint8(botmove.piece), player, botmove.end_pos, 'q')

			if valid {
				pgn_piece := string(indexPieceMap[botmove.piece])

				if botmove.piece > 15 {
					if self {
						pgn_piece = string(board.whitePieceMap[botmove.piece])
					} else {
						pgn_piece = string(board.blackPieceMap[botmove.piece])
					}
				}

				capturestr := ""
				if capture != 0 {
					capturestr = "x"
				}
				if pgn_piece == "p" {
					pgn_piece = ""
					if capture == 0 {
						rank = ""
					}
				}
				log.Printf("%s%s%s%s ", strings.ToUpper(pgn_piece), rank, capturestr, botmove.end_pos.pgn())
			}
			c.WriteMessage(mt, []byte(board.fen()))
			time.Sleep(AI_GAME_DELAY)

			// Bot
			if valid {
				board.plays[int(board.zobristHash())] += 1
				start := time.Now()
				states := 0
				score, botmove := board.minimax(BOT_MINIMAX_DEPTH, math.Inf(-1), math.Inf(+1), self, &states)
				if math.Abs(score) < 0.1 && math.Abs(score) > 0 {
					botmove = pm[rand.Intn(len(pm))]
				}
				//log.Printf("Best Move with Score %f\n", score)
				rank := string(((board.black[botmove.piece] >> 3) & 0b111) + 97)
				var capture uint8
				botvalid, capture = board.MakeMove(uint8(botmove.piece), self, botmove.end_pos, 'q')
				if botvalid {

					pgn_piece := string(indexPieceMap[botmove.piece])

					if botmove.piece > 15 {
						if self {
							pgn_piece = string(board.whitePieceMap[botmove.piece])
						} else {
							pgn_piece = string(board.blackPieceMap[botmove.piece])
						}
					}

					capturestr := ""
					if capture != 0 {
						capturestr = "x"
					}
					if pgn_piece == "p" {
						pgn_piece = ""
						if capture == 0 {
							rank = ""
						}
					}
					log.Printf("%s%s%s%s ", strings.ToUpper(pgn_piece), rank, capturestr, botmove.end_pos.pgn())
				}
				d := time.Since(start)
				//log.Printf("Time: %s", d)
				total_time += d
			}
		}
		if m == 0 {
			m += 1
			log.Printf("%d. ", m)
		}
		board.rounds = m

		gamefinished := false
		// Test for checkmate or draw
		if botvalid {

			m += 1
			log.Printf("%d. ", m)
			board.plays[int(board.zobristHash())] += 1
			if len(board.possibleMoves(player)) == 0 {

				if !board.verifyState(player) {
					err = c.WriteMessage(mt, []byte("Checkmate"))
					log.Printf("0-1 ")
				} else {
					err = c.WriteMessage(mt, []byte("Draw"))
					log.Printf("1/2-1/2 ")
				}
				gamefinished = true
				log.Printf("Total Bot Time: %d", total_time)
			}
		} else {
			if len(board.possibleMoves(self)) == 0 {
				if !board.verifyState(self) {
					err = c.WriteMessage(mt, []byte("Checkmate"))
					log.Printf("1-0 ")
				} else {
					err = c.WriteMessage(mt, []byte("Draw"))
					log.Printf("1/2-1/2 ")
				}
				log.Printf("Total Bot Time: %d", total_time)
				gamefinished = true
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
			log.Printf("1/2-1/2 ")
			gamefinished = true
		}

		err = c.WriteMessage(mt, []byte(board.fen()))
		c.WriteMessage(mt, []byte(fmt.Sprintf("eval %.5f", board.evaluate())))
		if err != nil {
			log.Println("write:", err)
			break
		}
		if gamefinished {
			break
		}
		time.Sleep(AI_GAME_DELAY)
	}
}

var addr = flag.String("addr", "localhost:8080", "http service address")

func start_server() {
	log.Printf("Server starting at %s", *addr)
	http.HandleFunc("/echo", echo)
	http.HandleFunc("/ai", ai)
	http.Handle("/", http.FileServer(http.Dir("./static/")))
	http.ListenAndServe(*addr, nil)
}

func main() {

	//load_transposition_table("tpmemory3.tp")
	upgrader.CheckOrigin = func(r *http.Request) bool {
		return true
	}

	flag.Parse()
	//log.SetFlags(0)
	init_zhtable()

	/*err := exec.Command("rundll32", "url.dll,FileProtocolHandler", fmt.Sprintf("http://%s/", *addr)).Start()
	if err != nil {
		log.Fatal(err)
	}*/

	start_server()

	/*debug := true
	w := webview.New(debug)
	defer w.Destroy()
	w.SetTitle("Extreme Go Chess")
	w.SetSize(800, 800, webview.Hint(webview.HintNone))
	w.Navigate(fmt.Sprintf("http://%s/?s=black", *addr))
	w.Run()*/
}
