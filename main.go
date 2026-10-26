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

type model struct {
	size                 int
	selectedX, selectedY int
	grid                 [][]rune
	answergrid           [][]rune
	horizontalFill       bool // is the grid being filled out horizontally, if false then its vertical
}

func newModel(size int) model {
	grid := make([][]rune, size)
	for x := range grid {
		grid[x] = make([]rune, size)
		for y := range grid[x] {
			if (x == 0 || x == size-1) && (y == 0 || y == size-1) {
				grid[x][y] = blocked
			}
		}
	}
	return model{size: size, grid: grid, horizontalFill: true}
}

func (m model) Init() tea.Cmd { return nil }

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "esc":
			return m, tea.Quit
		case "up":
			if m.isInBounds(m.selectedX, m.selectedY-1) {
				m.selectedY -= 1
			}
		case "down":
			if m.isInBounds(m.selectedX, m.selectedY+1) {
				m.selectedY += 1
			}
		case "left":
			if m.isInBounds(m.selectedX-1, m.selectedY) {
				m.selectedX -= 1
			}
		case "right":
			if m.isInBounds(m.selectedX+1, m.selectedY) {
				m.selectedX += 1
			}
		case "tab", " ": // swap between horizontal and vertical fill
			m.horizontalFill = !m.horizontalFill
		case "backspace": // clear the current letter, or if there isn't one step back and clear that cell
			if r := m.grid[m.selectedX][m.selectedY]; r == 0 || r == blocked {
				if m.horizontalFill && m.isInBounds(m.selectedX-1, m.selectedY) {
					m.selectedX -= 1
				} else if !m.horizontalFill && m.isInBounds(m.selectedX, m.selectedY-1) {
					m.selectedY -= 1
				}
			}
			if m.grid[m.selectedX][m.selectedY] != blocked {
				m.grid[m.selectedX][m.selectedY] = 0
			}
		default:
			if msg.Type == tea.KeyRunes && len(msg.Runes) == 1 {
				r := msg.Runes[0]
				if unicode.IsLetter(r) && m.grid[m.selectedX][m.selectedY] != blocked {
					m.grid[m.selectedX][m.selectedY] = unicode.ToLower(r)
					if m.horizontalFill && m.isInBounds(m.selectedX+1, m.selectedY) {
						m.selectedX += 1
					} else if !m.horizontalFill && m.isInBounds(m.selectedX, m.selectedY+1) {
						m.selectedY += 1
					}
				}
			}
		}
	}
	return m, nil
}

// todo: also check if its blocked
func (m model) isInBounds(x, y int) bool {
	if x >= 0 && x < m.size && y >= 0 && y < m.size {
		return true
	}
	return false
}

func (m model) View() string {
	return renderGrid(m.size, m.selectedX, m.selectedY, m.grid) + "\ntab/space to switch between across and down clue\n(press esc or ctrl+c to quit)\n"
}

// TODO: swap this out for a snazzy lip gloss version !!!
// renderGrid draws an n x n grid of empty squares using box-drawing characters.
func renderGrid(n int, selx int, sely int, grid [][]rune) string {
	line := func(left, mid, right, fill string) string {
		cells := make([]string, n)
		for i := range cells {
			cells[i] = fill
		}
		return left + strings.Join(cells, mid) + right + "\n"
	}

	var b strings.Builder
	b.WriteString(line("┌", "┬", "┐", "───"))
	for row := 0; row < n; row++ {
		for col := 0; col < n; col++ {
			b.WriteString("│" + cellText(grid[col][row], row == sely && col == selx))
		}
		b.WriteString("│\n")
		if row < n-1 {
			b.WriteString(line("├", "┼", "┤", "───"))
		}
	}
	b.WriteString(line("└", "┴", "┘", "───"))
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
	// TODO: add answer grid, check current against answers
	size := flag.Int("size", 5, "width and height of the grid")
	flag.Parse()
	if *size < 1 {
		fmt.Println("size must be at least 1")
		os.Exit(1)
	}
	if _, err := tea.NewProgram(newModel(*size)).Run(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
