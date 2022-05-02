package moveset

var Pst = [...]float64{

	// Pawn
	0 / 100.0, 0 / 100.0, 0 / 100.0, 0 / 100.0, 0 / 100.0, 0 / 100.0, 0 / 100.0, 0 / 100.0,
	5 / 100.0, 10 / 100.0, 10 / 100.0, -20 / 100.0, -20 / 100.0, 10 / 100.0, 10 / 100.0, 5 / 100.0,
	5 / 100.0, -5 / 100.0, -10 / 100.0, 0 / 100.0, 0 / 100.0, -10 / 100.0, -5 / 100.0, 5 / 100.0,
	0 / 100.0, 0 / 100.0, 0 / 100.0, 20 / 100.0, 20 / 100.0, 0 / 100.0, 0 / 100.0, 0 / 100.0,
	5 / 100.0, 5 / 100.0, 10 / 100.0, 25 / 100.0, 25 / 100.0, 10 / 100.0, 5 / 100.0, 5 / 100.0,
	10 / 100.0, 10 / 100.0, 20 / 100.0, 30 / 100.0, 30 / 100.0, 20 / 100.0, 10 / 100.0, 10 / 100.0,
	50 / 100.0, 50 / 100.0, 50 / 100.0, 50 / 100.0, 50 / 100.0, 50 / 100.0, 50 / 100.0, 50 / 100.0,
	0 / 100.0, 0 / 100.0, 0 / 100.0, 0 / 100.0, 0 / 100.0, 0 / 100.0, 0 / 100.0, 0 / 100.0,

	//Knights
	-50 / 100.0, -40 / 100.0, -30 / 100.0, -30 / 100.0, -30 / 100.0, -30 / 100.0, -40 / 100.0, -50 / 100.0,
	-40 / 100.0, -20 / 100.0, 0 / 100.0, 0 / 100.0, 0 / 100.0, 0 / 100.0, -20 / 100.0, -40 / 100.0,
	-30 / 100.0, 0 / 100.0, 10 / 100.0, 15 / 100.0, 15 / 100.0, 10 / 100.0, 0 / 100.0, -30 / 100.0,
	-30 / 100.0, 5 / 100.0, 15 / 100.0, 20 / 100.0, 20 / 100.0, 15 / 100.0, 5 / 100.0, -30 / 100.0,
	-30 / 100.0, 0 / 100.0, 15 / 100.0, 20 / 100.0, 20 / 100.0, 15 / 100.0, 0 / 100.0, -30 / 100.0,
	-30 / 100.0, 5 / 100.0, 10 / 100.0, 15 / 100.0, 15 / 100.0, 10 / 100.0, 5 / 100.0, -30 / 100.0,
	-40 / 100.0, -20 / 100.0, 0 / 100.0, 5 / 100.0, 5 / 100.0, 0 / 100.0, -20 / 100.0, -40 / 100.0,
	-50 / 100.0, -40 / 100.0, -30 / 100.0, -30 / 100.0, -30 / 100.0, -30 / 100.0, -40 / 100.0, -50 / 100.0,

	// Bishops
	-50 / 100.0, -40 / 100.0, -30 / 100.0, -30 / 100.0, -30 / 100.0, -30 / 100.0, -40 / 100.0, -50 / 100.0,
	-40 / 100.0, -20 / 100.0, 0 / 100.0, 5 / 100.0, 5 / 100.0, 0 / 100.0, -20 / 100.0, -40 / 100.0,
	-30 / 100.0, 0 / 100.0, 10 / 100.0, 15 / 100.0, 15 / 100.0, 10 / 100.0, 0 / 100.0, -30 / 100.0,
	-30 / 100.0, 5 / 100.0, 15 / 100.0, 20 / 100.0, 20 / 100.0, 15 / 100.0, 5 / 100.0, -30 / 100.0,
	-30 / 100.0, 0 / 100.0, 15 / 100.0, 20 / 100.0, 20 / 100.0, 15 / 100.0, 0 / 100.0, -30 / 100.0,
	-30 / 100.0, 5 / 100.0, 10 / 100.0, 15 / 100.0, 15 / 100.0, 10 / 100.0, 5 / 100.0, -30 / 100.0,
	-40 / 100.0, -20 / 100.0, 0 / 100.0, 0 / 100.0, 0 / 100.0, 0 / 100.0, -20 / 100.0, -40 / 100.0,
	-50 / 100.0, -40 / 100.0, -30 / 100.0, -30 / 100.0, -30 / 100.0, -30 / 100.0, -40 / 100.0, -50 / 100.0,

	// Rook
	0 / 100.0, 0 / 100.0, 0 / 100.0, 5 / 100.0, 5 / 100.0, 0 / 100.0, 0 / 100.0, 0 / 100.0,
	-5 / 100.0, 0 / 100.0, 0 / 100.0, 0 / 100.0, 0 / 100.0, 0 / 100.0, 0 / 100.0, -5 / 100.0,
	-5 / 100.0, 0 / 100.0, 0 / 100.0, 0 / 100.0, 0 / 100.0, 0 / 100.0, 0 / 100.0, -5 / 100.0,
	-5 / 100.0, 0 / 100.0, 0 / 100.0, 0 / 100.0, 0 / 100.0, 0 / 100.0, 0 / 100.0, -5 / 100.0,
	-5 / 100.0, 0 / 100.0, 0 / 100.0, 0 / 100.0, 0 / 100.0, 0 / 100.0, 0 / 100.0, -5 / 100.0,
	-5 / 100.0, 10 / 100.0, 10 / 100.0, 10 / 100.0, 10 / 100.0, 10 / 100.0, 10 / 100.0, 5 / 100.0,
	-5 / 100.0, 0 / 100.0, 0 / 100.0, 0 / 100.0, 0 / 100.0, 0 / 100.0, 0 / 100.0, -5 / 100.0,

	// Queen
	-20 / 100.0, -10 / 100.0, -10 / 100.0, -5 / 100.0, -5 / 100.0, -10 / 100.0, -10 / 100.0, -20 / 100.0,
	-10 / 100.0, 0 / 100.0, 0 / 100.0, 0 / 100.0, 0 / 100.0, 0 / 100.0, 0 / 100.0, -10 / 100.0,
	-10 / 100.0, 0 / 100.0, 5 / 100.0, 5 / 100.0, 5 / 100.0, 5 / 100.0, 0 / 100.0, -10 / 100.0,
	-5 / 100.0, 0 / 100.0, 5 / 100.0, 5 / 100.0, 5 / 100.0, 5 / 100.0, 0 / 100.0, -5 / 100.0,
	-5 / 100.0, 0 / 100.0, 5 / 100.0, 5 / 100.0, 5 / 100.0, 5 / 100.0, 0 / 100.0, -5 / 100.0,
	-10 / 100.0, 5 / 100.0, 5 / 100.0, 5 / 100.0, 5 / 100.0, 5 / 100.0, 0 / 100.0, -10 / 100.0,
	-10 / 100.0, 0 / 100.0, 5 / 100.0, 0 / 100.0, 0 / 100.0, 0 / 100.0, 0 / 100.0, -10 / 100.0,
	-20 / 100.0, -10 / 100.0, -10 / 100.0, -5 / 100.0, -5 / 100.0, -10 / 100.0, -10 / 100.0, -20 / 100.0,
}
