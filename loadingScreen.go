package main

import (

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	tint "github.com/lrstanley/bubbletint"
)

type LoadingScreen struct {
  spinner spinner.Model
  res [][]RowEntry
  status string
  loadResult chan tea.Model
}

func NewLoadingScreen(resultChan chan tea.Model) LoadingScreen {
  s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(tint.BrightCyan())
  
  return LoadingScreen{
    spinner: s,
    status: "Loading sheet...",
    loadResult: resultChan,
  }
}

func (l LoadingScreen) Init() tea.Cmd {
  return l.spinner.Tick
}

func (l LoadingScreen) Update(msg tea.Msg) (tea.Model, tea.Cmd){

  select {
  case model, ok := <- l.loadResult:
    if ok {
      return model, nil
    }
    l.status = "Failed to load data!"
  default:
    // pass
  }

  switch msg := msg.(type) {
  case tea.KeyMsg:
    switch msg.String() {
      case "esc", "ctrl+c":
        return l, tea.Quit
    }

  default:
    var cmd tea.Cmd
    l.spinner, cmd = l.spinner.Update(msg)
    return l, cmd
  }

  return l, nil
}

func (l LoadingScreen) View() string {
  s := "Loading... " + l.spinner.View() + "\n"
  return s
}
