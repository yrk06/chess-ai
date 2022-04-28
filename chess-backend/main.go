package main

import (
	"fmt"
	"log"
	"os"
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

type Piece uint8

type PlayerPieces [16]Piece

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

//rnbqkbnr/pppppppp/11111111/11111111/11111111/11111111/PPPPPPPP/RNBQKBNR w KQkq - 0 1
func (c *Chessboard) fen() string {
	fen := []byte("11111111/11111111/11111111/11111111/11111111/11111111/11111111/11111111 w KQkq - 0 1")
	King := Location{x: 1, y: 1}
	fen[King.toFen()] = 'K'
	return string(fen)
}

type Location struct {
	x int
	y int
}

func (l *Location) toFen() int {
	return (l.y)*9 + (l.x)
}

func (l *Location) pgn() string {
	return fmt.Sprintf("%s%d", string(l.x+97), l.y+1)
}

func (l *Location) frompgn(pos string) {
	l.x = int(pos[0]) - 97
	l.y = int(pos[1]-'0') - 1
}

func (l *Location) toByte() uint8 {
	return uint8(((l.x & 0b111) << 3) | (l.y & 0b111))
}

func (l *Location) fromByte(b uint8) {
	l.x = int((b >> 3) & 0b111)
	l.y = int((b) & 0b111)
}

func isMoveValid(piece uint8, start_pos Location, end_pos Location) bool {
	valid := false
	ep := end_pos.toByte()
	for moveLines := 0; moveLines < 8; moveLines++ {
		for move := 0; move < 7; move++ {
			value := moveset.Mset[pieceMap[piece]+56*start_pos.y+56*8*start_pos.x+moveLines*7+move]

			if value == 0 {
				break
			}
			if ep == (value & 0b111111) {
				valid = true
				break
			}
			if valid {
				break
			}
		}
	}
	return valid
}

func main() {
	start := os.Args[1]
	end := os.Args[2]

	piece := start[0]
	start_pos := Location{}
	start_pos.frompgn(start[1:])
	end_pos := Location{}
	end_pos.frompgn(end)

	log.Println(isMoveValid(piece, start_pos, end_pos))

	board := Chessboard{}

	log.Println(board.fen())
}
