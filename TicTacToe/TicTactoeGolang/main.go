package main

import (
	"bufio"
	"errors"
	"fmt"
	"math"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"time"
)

// Marks
type Mark rune

const (
	Empty Mark = '.'
	X     Mark = 'X'
	O     Mark = 'O'
)

var winLines = [8][3]int{
	{0, 1, 2}, {3, 4, 5}, {6, 7, 8}, // rows
	{0, 3, 6}, {1, 4, 7}, {2, 5, 8}, // cols
	{0, 4, 8}, {2, 4, 6}, // diags
}

// Board encapsulates tic-tac-toe board (3x3)
type Board struct {
	cells [9]Mark
}

func NewBoard() *Board {
	b := &Board{}
	for i := range b.cells {
		b.cells[i] = Empty
	}
	return b
}

func (b *Board) Clone() *Board {
	nb := &Board{}
	copy(nb.cells[:], b.cells[:])
	return nb
}

func (b *Board) IsFull() bool {
	for _, c := range b.cells {
		if c == Empty {
			return false
		}
	}
	return true
}

func (b *Board) AvailableMoves() []int {
	var moves []int
	for i, c := range b.cells {
		if c == Empty {
			moves = append(moves, i)
		}
	}
	return moves
}

func (b *Board) MakeMove(idx int, m Mark) error {
	if idx < 0 || idx >= 9 {
		return errors.New("index out of bounds")
	}
	if b.cells[idx] != Empty {
		return errors.New("cell occupied")
	}
	b.cells[idx] = m
	return nil
}

func (b *Board) Winner() (Mark, bool) {
	for _, line := range winLines {
		a, b1, c := b.cells[line[0]], b.cells[line[1]], b.cells[line[2]]
		if a != Empty && a == b1 && a == c {
			return a, true
		}
	}
	return Empty, false
}

func (b *Board) String() string {
	var sb strings.Builder
	for r := 0; r < 3; r++ {
		for c := 0; c < 3; c++ {
			sb.WriteRune(rune(b.cells[r*3+c]))
			if c < 2 {
				sb.WriteString(" | ")
			}
		}
		if r < 2 {
			sb.WriteString("\n---------\n")
		}
	}
	return sb.String()
}

// Player interface: returns index 0..8 for move
type Player interface {
	Move(b *Board, mark Mark) (int, error)
	Name() string
}

// Human CLI player
type Human struct {
	reader *bufio.Reader
	name   string
}

func NewHuman(name string) *Human {
	return &Human{reader: bufio.NewReader(os.Stdin), name: name}
}

func (h *Human) Name() string { return h.name }

func (h *Human) Move(b *Board, mark Mark) (int, error) {
	fmt.Printf("%s (%c), enter move (0-8): ", h.name, mark)
	line, err := h.reader.ReadString('\n')
	if err != nil {
		return -1, err
	}
	line = strings.TrimSpace(line)
	i, err := strconv.Atoi(line)
	if err != nil {
		return -1, errors.New("invalid number")
	}
	if i < 0 || i > 8 {
		return -1, errors.New("index out of range")
	}
	if b.cells[i] != Empty {
		return -1, errors.New("cell occupied")
	}
	return i, nil
}

// Random player (for testing)
type RandomPlayer struct{ name string }

func NewRandom(name string) *RandomPlayer { return &RandomPlayer{name: name} }
func (r *RandomPlayer) Name() string      { return r.name }
func (r *RandomPlayer) Move(b *Board, mark Mark) (int, error) {
	moves := b.AvailableMoves()
	if len(moves) == 0 {
		return -1, errors.New("no moves")
	}
	return moves[rand.Intn(len(moves))], nil
}

// Minimax AI player
type MinimaxAI struct {
	name string
	me   Mark
}

func NewMinimax(name string) *MinimaxAI { return &MinimaxAI{name: name} }

func (ai *MinimaxAI) Name() string { return ai.name }

// evaluate returns a score for a terminal board
// +1 if AI wins, -1 if opponent wins, 0 if draw, NaN if not terminal
func (ai *MinimaxAI) evaluate(b *Board) float64 {
	if w, ok := b.Winner(); ok {
		if w == ai.me {
			return 1
		}
		return -1
	}
	if b.IsFull() {
		return 0
	}
	return math.NaN()
}

// Move picks best index using minimax
func (ai *MinimaxAI) Move(b *Board, mark Mark) (int, error) {
	ai.me = mark
	bestScore := math.Inf(-1)
	bestMove := -1
	for _, mv := range b.AvailableMoves() {
		nb := b.Clone()
		_ = nb.MakeMove(mv, mark)
		score := ai.minimax(nb, switchMark(mark), false)
		if score > bestScore {
			bestScore = score
			bestMove = mv
		}
	}
	if bestMove == -1 {
		return -1, errors.New("no moves available")
	}
	return bestMove, nil
}

// minimax with evaluation function
func (ai *MinimaxAI) minimax(b *Board, current Mark, maximizing bool) float64 {
	score := ai.evaluate(b)
	if !math.IsNaN(score) {
		return score
	}

	if maximizing {
		best := math.Inf(-1)
		for _, mv := range b.AvailableMoves() {
			nb := b.Clone()
			_ = nb.MakeMove(mv, current)
			score := ai.minimax(nb, switchMark(current), false)
			if score > best {
				best = score
			}
		}
		return best
	} else {
		best := math.Inf(1)
		for _, mv := range b.AvailableMoves() {
			nb := b.Clone()
			_ = nb.MakeMove(mv, current)
			score := ai.minimax(nb, switchMark(current), true)
			if score < best {
				best = score
			}
		}
		return best
	}
}

func switchMark(m Mark) Mark {
	if m == X {
		return O
	}
	return X
}

// Game orchestrator
type Game struct {
	board   *Board
	pX, pO  Player
	current Mark
}

func NewGame(px, po Player) *Game {
	return &Game{
		board:   NewBoard(),
		pX:      px,
		pO:      po,
		current: X,
	}
}

func (g *Game) Play() (Mark, error) {
	for {
		fmt.Println("\nBoard:")
		fmt.Println(g.board.String())
		if w, ok := g.board.Winner(); ok {
			fmt.Printf("Winner: %c\n", w)
			return w, nil
		}
		if g.board.IsFull() {
			fmt.Println("Draw")
			return Empty, nil
		}

		var p Player
		if g.current == X {
			p = g.pX
		} else {
			p = g.pO
		}
		move, err := p.Move(g.board, g.current)
		if err != nil {
			fmt.Printf("Player move error: %v\n", err)
			continue
		}
		if err := g.board.MakeMove(move, g.current); err != nil {
			fmt.Printf("Invalid move: %v\n", err)
			continue
		}
		g.current = switchMark(g.current)
	}
}

func main() {
	rand.Seed(time.Now().UnixNano())
	reader := bufio.NewReader(os.Stdin)

	fmt.Println("Tic-Tac-Toe - CLI demonstration")
	// Example: Human vs Minimax
	for {
		h := NewHuman("You")
		ai := NewMinimax("AI")
		game := NewGame(h, ai) // Human is X, AI is O

		winner, err := game.Play()
		if err != nil {
			fmt.Printf("Game ended with error: %v\n", err)
			return
		}
		if winner == Empty {
			fmt.Println("Game ended in a draw!")
		} else {
			fmt.Printf("Game over! Winner: %c\n", winner)
		}

		// Ask to play again
		fmt.Print("Do you want to play again? (y/n): ")
		answer, _ := reader.ReadString('\n')
		answer = strings.TrimSpace(strings.ToLower(answer))

		if answer != "y" {
			fmt.Println("Thanks for playing! Goodbye 👋")
			break
		}
	}
}
