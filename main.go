package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

type model struct {
	size                 int
	selectedX, selectedY int
	grid                 [][]rune
	answergrid           [][]rune
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
		default:
			if msg.Type == tea.KeyRunes && len(msg.Runes) == 1 {
				m.grid[m.selectedX][m.selectedY] = msg.Runes[0]
			}
		}
	}
	return m, nil
}

func (m model) isInBounds(x, y int) bool {
	if x >= 0 && x < m.size && y >= 0 && y < m.size {
		return true
	}
	return false
}

func (m model) View() string {
	return renderGrid(m.size, m.selectedX, m.selectedY, m.grid) + "\n(press esc or ctrl+c to quit)\n"
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
			if row == sely && col == selx {
				b.WriteString(fmt.Sprintf("│\x1b[7m %c \x1b[0m", cellText(row, col, grid)))
			} else {
				b.WriteString(fmt.Sprintf("│ %c ", cellText(row, col, grid)))
			}
		}
		b.WriteString("|\n")
		if row < n-1 {
			b.WriteString(line("├", "┼", "┤", "───"))
		}
	}
	b.WriteString(line("└", "┴", "┘", "───"))
	return b.String()
}

func cellText(row, col int, grid [][]rune) rune {
	if grid[col][row] != 0 {
		return grid[col][row]
	}
	return ' '
}

func main() {
	// TODO: add a newModel func
	size := flag.Int("size", 5, "width and height of the grid")
	flag.Parse()
	if *size < 1 {
		fmt.Println("size must be at least 1")
		os.Exit(1)
	}
	grid := make([][]rune, *size)
	for i := range grid {
		grid[i] = make([]rune, *size)
	}

	if _, err := tea.NewProgram(model{size: *size, grid: grid}).Run(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
