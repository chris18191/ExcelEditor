package main

import (

	"github.com/charmbracelet/bubbles/key"
)

// keyMap defines a set of keybindings. To work for help it must satisfy
// key.Map. It could also very easily be a map[string]key.Binding.
type keyMap struct {
	PrevDay    key.Binding
	NextDay  key.Binding
	Quit  key.Binding
	Help  key.Binding
  Up key.Binding
  Down key.Binding
  Left key.Binding
  Right key.Binding

  Edit key.Binding
  CancelEdit key.Binding
  Save key.Binding
  FocusPrev key.Binding
  FocusNext key.Binding

  Add key.Binding
  Delete key.Binding
}

// ShortHelp returns keybindings to be shown in the mini help view. It's part
// of the key.Map interface.
func (k keyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Help, k.Quit, k.Add, k.Delete}
}

// FullHelp returns keybindings for the expanded help view. It's part of the
// key.Map interface.
func (k keyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.PrevDay, k.Left,  k.Down, k.FocusPrev, k.Edit, k.Add, k.CancelEdit, k.Help},     // first column
		{k.NextDay, k.Right, k.Up,   k.FocusNext, k.Save, k.Delete, k.Quit},           // second column
	}
}

var keys = keyMap{
	Add: key.NewBinding(
		key.WithKeys("a"),
		key.WithHelp("a", "Add"),
	),
	Delete: key.NewBinding(
		key.WithKeys("d"),
		key.WithHelp("d", "Delete"),
	),
	Up: key.NewBinding(
		key.WithKeys("k"),
		key.WithHelp("k", "Up"),
	),
	Down: key.NewBinding(
		key.WithKeys("j"),
		key.WithHelp("j", "Down"),
	),
	Left: key.NewBinding(
		key.WithKeys("h"),
		key.WithHelp("h", "Left"),
	),
	Right: key.NewBinding(
		key.WithKeys("l"),
		key.WithHelp("l", "Right"),
	),
	Edit: key.NewBinding(
		key.WithKeys("enter"),
		key.WithHelp("enter", "Edit"),
	),
	CancelEdit: key.NewBinding(
		key.WithKeys("esc"),
		key.WithHelp("esc", "Cancel"),
	),


	FocusNext: key.NewBinding(
		key.WithKeys("tab"),
		key.WithHelp("tab", "Next input"),
	),
	FocusPrev: key.NewBinding(
		key.WithKeys("shift+tab"),
		key.WithHelp("shift+tab", "Prev input"),
	),
	Save: key.NewBinding(
		key.WithKeys("ctrl+s"),
		key.WithHelp("ctrl+s", "Save"),
	),
	PrevDay: key.NewBinding(
		key.WithKeys("q", "ctrl+h"),
		key.WithHelp("q", "Previous Day"),
	),
	NextDay: key.NewBinding(
		key.WithKeys("e", "ctrl+l"),
		key.WithHelp("e", "Next Day"),
	),
	Quit: key.NewBinding(
		key.WithKeys("ctrl+c"),
		key.WithHelp("crtl+c", "quit"),
	),
	Help: key.NewBinding(
		key.WithKeys("?"),
		key.WithHelp("?", "toggle help"),
	),
}
