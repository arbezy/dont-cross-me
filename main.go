package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

type model struct {
	size int
}

func (m model) Init() tea.Cmd { return nil }

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if _, ok := msg.(tea.KeyMsg); ok {
		return m, tea.Quit
	}
	return m, nil
}

func (m model) View() string {
	return renderGrid(m.size) + "\n(press any key to quit)\n"
}

// renderGrid draws an n x n grid of empty squares using box-drawing characters.
func renderGrid(n int) string {
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
		b.WriteString(line("│", "│", "│", "   "))
		if row < n-1 {
			b.WriteString(line("├", "┼", "┤", "───"))
		}
	}
	b.WriteString(line("└", "┴", "┘", "───"))
	return b.String()
}

func main() {
	size := flag.Int("size", 5, "width and height of the grid")
	flag.Parse()
	if *size < 1 {
		fmt.Println("size must be at least 1")
		os.Exit(1)
	}

	if _, err := tea.NewProgram(model{size: *size}).Run(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
