package main

import (
	"flag"
	"fmt"
	"os"
	"strings"
	"unicode"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
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
	acrossClues          []clue
	downClues            []clue
}

type clue struct {
	number int
	text   string
}

func newModel(size int) model {
	// TODO: load real answers instead of filling every open cell with 'a'
	m := model{
		size:           size,
		grid:           newGrid(size, 0),
		answerGrid:     newGrid(size, 'a'),
		horizontalFill: true,
	}
	m.acrossClues, m.downClues = m.numberClues()
	return m
}

// numberClues numbers, in reading order, every cell that starts an across or down word
// of at least two letters, and returns a clue for each word.
// TODO: replace the filler text with real clues
func (m model) numberClues() (across, down []clue) {
	n := 0
	for y := 0; y < m.size; y++ {
		for x := 0; x < m.size; x++ {
			if !m.isOpen(x, y) {
				continue
			}
			startsAcross := !m.isOpen(x-1, y) && m.isOpen(x+1, y)
			startsDown := !m.isOpen(x, y-1) && m.isOpen(x, y+1)
			if !startsAcross && !startsDown {
				continue
			}
			n++
			if startsAcross {
				across = append(across, clue{n, fmt.Sprintf("test clue %d", n)})
			}
			if startsDown {
				down = append(down, clue{n, fmt.Sprintf("test clue %d", n)})
			}
		}
	}
	return across, down
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

// TODO: make it so that when we reach the end of the grid, i.e. complete a word, then we move on to the next one
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

// TODO: also check if its blocked
func (m model) isInBounds(x, y int) bool {
	return x >= 0 && x < m.size && y >= 0 && y < m.size
}

// isOpen reports whether (x, y) is an in-bounds cell that isn't blocked.
func (m model) isOpen(x, y int) bool {
	return m.isInBounds(x, y) && m.grid[x][y] != blocked
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
	const gap = "   "
	view := lipgloss.JoinHorizontal(lipgloss.Top,
		m.renderGrid(), gap,
		renderClues("Across", m.acrossClues), gap,
		renderClues("Down", m.downClues),
	) + "\n"
	if m.isSolved() {
		view += "you got the answer!\n"
	}
	return view + helpText
}

// renderClues draws a titled column of numbered clues.
func renderClues(title string, clues []clue) string {
	var b strings.Builder
	b.WriteString(title + "\n")
	for _, c := range clues {
		fmt.Fprintf(&b, "%d. %s\n", c.number, c.text)
	}
	return b.String()
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
