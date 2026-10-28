package main

import (
	"flag"
	"fmt"
	"os"
	"strings"
	"unicode"

	tea "github.com/charmbracelet/bubbletea"
)

const (
	blocked  = '!'
	selectBg = "\x1b[44m" // blue background
	selectFg = "\x1b[34m" // blue foreground
	reset    = "\x1b[0m"
)

const helpText = "tab/space to switch between across and down clue\n" +
	"(press esc or ctrl+c to quit)\n"

type model struct {
	size                 int
	selectedX, selectedY int
	grid                 [][]rune // indexed as grid[x][y]
	answerGrid           [][]rune
	horizontalFill       bool // is the grid being filled out horizontally, if false then its vertical
}

func newModel(size int) model {
	// TODO: load real answers instead of filling every open cell with 'a'
	return model{
		size:           size,
		grid:           newGrid(size, 0),
		answerGrid:     newGrid(size, 'a'),
		horizontalFill: true,
	}
}

// newGrid returns a size x size grid with the corners blocked and every other cell set to fill.
func newGrid(size int, fill rune) [][]rune {
	grid := make([][]rune, size)
	for x := range grid {
		grid[x] = make([]rune, size)
		for y := range grid[x] {
			if (x == 0 || x == size-1) && (y == 0 || y == size-1) {
				grid[x][y] = blocked
			} else {
				grid[x][y] = fill
			}
		}
	}
	return grid
}

func (m model) Init() tea.Cmd { return nil }

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	key, ok := msg.(tea.KeyMsg)
	if !ok {
		return m, nil
	}

	switch key.String() {
	case "ctrl+c", "esc":
		return m, tea.Quit
	case "up":
		m.move(0, -1)
	case "down":
		m.move(0, 1)
	case "left":
		m.move(-1, 0)
	case "right":
		m.move(1, 0)
	case "tab", " ": // swap between horizontal and vertical fill
		m.horizontalFill = !m.horizontalFill
	case "backspace": // clear the current letter, or if there isn't one step back and clear that cell
		if r := m.selectedCell(); r == 0 || r == blocked {
			m.step(-1)
		}
		if m.selectedCell() != blocked {
			m.setSelectedCell(0)
		}
	default:
		if key.Type != tea.KeyRunes || len(key.Runes) != 1 {
			break
		}
		if r := key.Runes[0]; unicode.IsLetter(r) && m.selectedCell() != blocked {
			m.setSelectedCell(unicode.ToLower(r))
			m.step(1)
		}
	}
	return m, nil
}

// move shifts the selection by (dx, dy) if the target cell is in bounds.
func (m *model) move(dx, dy int) {
	if m.isInBounds(m.selectedX+dx, m.selectedY+dy) {
		m.selectedX += dx
		m.selectedY += dy
	}
}

// step moves the selection by delta along the current fill direction.
func (m *model) step(delta int) {
	if m.horizontalFill {
		m.move(delta, 0)
	} else {
		m.move(0, delta)
	}
}

func (m model) selectedCell() rune {
	return m.grid[m.selectedX][m.selectedY]
}

func (m model) setSelectedCell(r rune) {
	m.grid[m.selectedX][m.selectedY] = r
}

// todo: also check if its blocked
func (m model) isInBounds(x, y int) bool {
	return x >= 0 && x < m.size && y >= 0 && y < m.size
}

// isSolved reports whether every cell in the grid matches the answer grid.
func (m model) isSolved() bool {
	for x := range m.grid {
		for y := range m.grid[x] {
			if m.grid[x][y] != m.answerGrid[x][y] {
				return false
			}
		}
	}
	return true
}

func (m model) View() string {
	view := m.renderGrid() + "\n"
	if m.isSolved() {
		view += "you got the answer!\n"
	}
	return view + helpText
}

// renderGrid draws the grid using box-drawing characters, highlighting the selected cell.
// TODO: swap this out for a snazzy lip gloss version !!!
func (m model) renderGrid() string {
	border := func(left, mid, right string) string {
		return left + strings.Repeat("───"+mid, m.size-1) + "───" + right + "\n"
	}

	var b strings.Builder
	b.WriteString(border("┌", "┬", "┐"))
	for y := 0; y < m.size; y++ {
		for x := 0; x < m.size; x++ {
			selected := x == m.selectedX && y == m.selectedY
			b.WriteString("│" + cellText(m.grid[x][y], selected))
		}
		b.WriteString("│\n")
		if y < m.size-1 {
			b.WriteString(border("├", "┼", "┤"))
		}
	}
	b.WriteString(border("└", "┴", "┘"))
	return b.String()
}

// cellText returns the 3-character-wide contents of a cell.
func cellText(r rune, selected bool) string {
	if r == blocked {
		if selected {
			return selectFg + "███" + reset
		}
		return "███"
	}
	if r == 0 {
		r = ' '
	}
	if selected {
		return fmt.Sprintf("%s %c %s", selectBg, r, reset)
	}
	return fmt.Sprintf(" %c ", r)
}

func main() {
	size := flag.Int("size", 5, "width and height of the grid")
	flag.Parse()
	if *size < 1 {
		fmt.Fprintln(os.Stderr, "size must be at least 1")
		os.Exit(1)
	}
	if _, err := tea.NewProgram(newModel(*size)).Run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
