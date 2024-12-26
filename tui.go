package main

import (
	"fmt"
	"io"
	"log/slog"
	"math"
	"os"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	tint "github.com/lrstanley/bubbletint"
)

var defaultConfig = Configuration{
	EXCEL_FILE:         "res/example.xlsx",
	COL_ID_DATE:        0,
	COL_ID_HOURS_START: 2,
	COL_ID_HOURS_END:   3,
	COL_ID_HOURS_PAUSE: 4,
	ROW_ID_ENTRY_START: 5, // sixth row contains first entries
}


type EntryList struct {
	Entries [][][]RowEntry
}

func NewEntryList() EntryList {
	return EntryList{
		Entries: ReturnAll(defaultConfig),
	}
}

type DatePicker struct {
	selected   bool
	currentDay time.Time
}

func NewDatePicker() DatePicker {
	var res DatePicker
	now := time.Now()
	res.currentDay = time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	return res
}

type Model struct {
	numColumns   int // equals the number of columns that are printed for each row
	spinner      spinner.Model
	datepicker   DatePicker
	entryList    EntryList
	debugMessage string

	keys               keyMap
	help               help.Model
	currentSelectedRow int

	editActive   bool
	textInputs   []textinput.Model
	focusedIndex int

	styles map[string]lipgloss.Style
	height int
	width  int
}

func initialModel() Model {

	help := help.New()
	// help.Styles.FullDesc = help.Styles.FullDesc.Background(tint.Bg())
	// help.Styles.ShortDesc = help.Styles.ShortDesc.Background(tint.Bg())
	// help.Styles.FullKey = help.Styles.FullKey.Background(tint.Bg())
	// help.Styles.ShortKey = help.Styles.ShortKey.Background(tint.Bg())
	// help.Styles.FullSeparator = help.Styles.FullSeparator.Background(tint.Bg())
	// help.Styles.ShortSeparator = help.Styles.ShortSeparator.Background(tint.Bg())
	// help.Styles.Ellipsis = help.Styles.Ellipsis.Background(tint.Bg())

	return Model{
		datepicker:   NewDatePicker(),
		entryList:    NewEntryList(),
		debugMessage: "",

		keys:               keys,
		help:               help,
		currentSelectedRow: 0,

		editActive:   false,
		numColumns:   5,
		textInputs:   []textinput.Model{},
		focusedIndex: 0,

		height: 50,
		width:  100,

		styles: map[string]lipgloss.Style{
			"header": lipgloss.NewStyle().
				Bold(true).
				Border(lipgloss.RoundedBorder(), true, true).
				Foreground(tint.Cyan()),
			"tableHeader": lipgloss.NewStyle().
				Bold(true).
				Border(lipgloss.RoundedBorder(), false, true, true, true).
				PaddingBottom(0).
				MaxHeight(1).
				Foreground(tint.Cyan()),
			"unselectedEntry": lipgloss.NewStyle().
				Inline(true).
				Foreground(tint.Fg()),
			"selectedEntry": lipgloss.NewStyle().
				Inline(true).
				Bold(true).
				Foreground(tint.Cyan()),
			"inputField": lipgloss.NewStyle().
				Italic(true).
				Foreground(tint.Cyan()),
			"inputFieldErr": lipgloss.NewStyle().
				Italic(true).
				Foreground(tint.Red()),
		},
	}
}

func (m Model) getMonthEntries(i int) *[][]RowEntry {
	return &m.entryList.Entries[i]
}

func (m Model) getDayEntries(month int, day int) *[]RowEntry {
	return &m.entryList.Entries[month][day]
}

func (m Model) getCurrentMonthEntries() *[][]RowEntry {
	return &m.entryList.Entries[m.datepicker.currentDay.Month()-1]
}

func (m Model) getCurrentDayEntries() *[]RowEntry {
	return &m.entryList.Entries[m.datepicker.currentDay.Month()-1][m.datepicker.currentDay.Day()-1]
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(
		m.spinner.Tick,
		tea.WindowSize(),
	)
}


func validateTime(s string) error {
	_, err := time.Parse("15:04", string(s))
	return err
}

func validateDuration(s string) error {
	_, err := time.ParseDuration(string(s))
	return err
}

func helperMod(a, b int) int {
	return (a%b + b) % b
}

func readTextInputWithDefault(i *textinput.Model) string {
	if i.Value() != "" {
		return i.Value()
	}

	return i.Placeholder
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, keys.PrevDay) && !m.editActive:
			m.datepicker.currentDay = m.datepicker.currentDay.AddDate(0, 0, -1)
			m.currentSelectedRow = int(math.Min(float64(m.currentSelectedRow), float64(len(*m.getCurrentDayEntries())-1)))
			// m.entryList.UpdateCurrentIndex(m.datepicker.currentDay)
		case key.Matches(msg, keys.NextDay) && !m.editActive:
			m.datepicker.currentDay = m.datepicker.currentDay.AddDate(0, 0, 1)
			m.currentSelectedRow = int(math.Min(float64(m.currentSelectedRow), float64(len(*m.getCurrentDayEntries())-1)))
			// m.entryList.UpdateCurrentIndex(m.datepicker.currentDay)
		case key.Matches(msg, keys.Up) && !m.editActive:
			m.currentSelectedRow = helperMod(m.currentSelectedRow-1, len(*m.getCurrentDayEntries()))
			m.debugMessage = fmt.Sprintf("Pressed up (selected=%d/%d)", m.currentSelectedRow, len(*m.getCurrentDayEntries()))
		case key.Matches(msg, keys.Down) && !m.editActive:
			m.currentSelectedRow = helperMod(m.currentSelectedRow+1, len(*m.getCurrentDayEntries()))
			m.debugMessage = fmt.Sprintf("Pressed down (selected=%d/%d)", m.currentSelectedRow, len(*m.getCurrentDayEntries()))
		case key.Matches(msg, keys.Left) && !m.editActive:
			m.debugMessage = "Pressed left"
		case key.Matches(msg, keys.Right) && !m.editActive:
			m.debugMessage = "Pressed right"

		case key.Matches(msg, keys.Save) && !m.editActive:
			m.debugMessage = "Pressed save"
		case key.Matches(msg, keys.FocusPrev) && !m.editActive:
			if !m.editActive {
				break
			}
			m.focusedIndex = helperMod(m.focusedIndex-1, len(m.textInputs))
			m.debugMessage = fmt.Sprintf("Focused index: %d", m.focusedIndex)
		case key.Matches(msg, keys.FocusNext):
			if !m.editActive {
				break
			}
			m.focusedIndex = helperMod(m.focusedIndex+1, len(m.textInputs))
			m.debugMessage = fmt.Sprintf("Focused index: %d", m.focusedIndex)

		case key.Matches(msg, keys.Delete) && !m.editActive && len(m.entryList.Entries) > 0:
			m.debugMessage = "Trying to delete..."
			todaysEntries := m.getCurrentDayEntries()
			*todaysEntries = append((*todaysEntries)[0:m.currentSelectedRow], (*todaysEntries)[m.currentSelectedRow+1:]...)
			if len(*todaysEntries) == 0 {
				m.currentSelectedRow = 0
			} else {
				if m.currentSelectedRow > len(*todaysEntries)-1 {
					m.currentSelectedRow = len(*todaysEntries) - 1
				}
			}
		case key.Matches(msg, keys.Add) && !m.editActive:
			todaysEntries := m.getCurrentDayEntries()
			*todaysEntries = append(*todaysEntries,
				RowEntry{
					Date:      (*todaysEntries)[len(*todaysEntries)-1].Date,
					SheetName: (*todaysEntries)[len(*todaysEntries)-1].SheetName,
					Start:     (*todaysEntries)[len(*todaysEntries)-1].End,
				})
			m.currentSelectedRow = len(*todaysEntries) - 1
			fallthrough // automatically edit new entry
		case key.Matches(msg, keys.Edit):
			m.debugMessage = "Pressed edit"
			m.editActive = !m.editActive
			entry := (*m.getCurrentDayEntries())[m.currentSelectedRow]
			if m.editActive {
				// create inputs for all columns except date
				inputs := make([]textinput.Model, m.numColumns)
				for i := range inputs {
					t := textinput.New()
					switch i {
					case 0:
						// Start time
						t.Placeholder = entry.Start.Format("15:04")
						t.CharLimit = 5
						t.Width = 5
						t.Validate = validateTime
					case 1:
						// End time
						t.Placeholder = entry.End.Format("15:04")
						t.CharLimit = 5
						t.Width = 5
						t.Validate = validateTime
					case 2:
						// Pause
						t.Placeholder = entry.Pause.String()
						t.CharLimit = 9
						t.Width = 9
						t.Validate = validateDuration
					case 3:
						if t.Placeholder = entry.ProjectNr; t.Placeholder == "" {
							t.Placeholder = "Project-Nr."
						}
						t.CharLimit = 9
						t.Width = 9
					case 4:
						t.Placeholder = "Description"
						t.SetValue(entry.Description)
						// if t.Placeholder = entry.Description; t.Placeholder == "" {
						// }
						t.Width = 40
					default:
						t.Placeholder = "UNDEFINED FIELD"
					}

					inputs[i] = t
				}
				m.textInputs = inputs
				m.focusedIndex = 0
			} else {
				m.debugMessage = "Saved entry starting at " + m.textInputs[0].Value()
				entry.Start, _ = time.Parse("15:04", readTextInputWithDefault(&m.textInputs[0]))
				entry.End, _ = time.Parse("15:04", readTextInputWithDefault(&m.textInputs[1]))
				entry.Pause, _ = time.ParseDuration(readTextInputWithDefault(&m.textInputs[2]))
				entry.ProjectNr = readTextInputWithDefault(&m.textInputs[3])
				// TODO write customer from project number

				entry.Description = readTextInputWithDefault(&m.textInputs[4])
			}
			m.entryList.Entries[m.datepicker.currentDay.Month()-1][m.datepicker.currentDay.Day()-1][m.currentSelectedRow] = entry

		case key.Matches(msg, keys.CancelEdit):
			if m.editActive {
				m.editActive = false
				m.textInputs = []textinput.Model{}
			}

		case key.Matches(msg, keys.Help):
			m.help.ShowAll = !m.help.ShowAll
		case key.Matches(msg, keys.Quit):
			return m, tea.Quit
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.debugMessage = fmt.Sprintf("Resized to %dx%d", m.width, m.height)

	default:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd
	}

	if m.editActive {
		return m, m.updateInputs(msg)
	}

	return m, nil
}

func (m *Model) updateInputs(msg tea.Msg) tea.Cmd {
	cmds := make([]tea.Cmd, len(m.textInputs))

	// Only text inputs with Focus() set will respond, so it's safe to simply
	// update all of them here without any further logic.
	for i := range m.textInputs {
		m.textInputs[i], cmds[i] = m.textInputs[i].Update(msg)
	}

	return tea.Batch(cmds...)
}

func (r RowEntry) View() string {
	return fmt.Sprintf("%10s  %8s → %-8s [%9s] %20s",
		r.Date.Format("Mon 02.01."),
		r.Start.Format("15:04"),
		r.End.Format("15:04"),
		r.Pause.String(),
		r.Project,
	)
}

func (m *Model) ViewAsEdit() string {
	s := ""
	m.debugMessage = fmt.Sprintf("Creating edit view for %d inputs", len(m.textInputs))
	for i := range m.textInputs {
		res := ""
		if m.focusedIndex == i {
			m.textInputs[i].Focus()
		} else {
			m.textInputs[i].Blur()
		}
		if m.textInputs[i].Err != nil {
			res = m.styles["inputFieldErr"].Render(m.textInputs[i].View()) + "\t"
		} else {
			res = m.styles["inputField"].Render(m.textInputs[i].View()) + "\t"
		}
		s += res
		// m.debugMessage += fmt.Sprintf("Focused %d", m.focusedIndex)
	}
	return m.styles["inputField"].Render(s)
}

func (m Model) View() string {
	s := ""
	s += m.styles["header"].Render("Work Hour Editor")
	// s += strings.Repeat(" ", m.width+100) + "\n"
	s += "\n"
	s += fmt.Sprintf("Current Date: [%12s] \n", m.datepicker.currentDay.Format("Mon 02.01.06"))
	s += fmt.Sprintf("\n")

	s += m.styles["tableHeader"].Render(
		fmt.Sprintf(" %-10s  %-8s   %-8s %-9s   %-20s", "Date", "Start", "End", "Pause", "Description"),
	)
	s += "\n"

	indent := "  "
	todaysEntries := *m.getCurrentDayEntries()
	if len(todaysEntries) > 0 {
		for i := 0; i < m.currentSelectedRow; i++ {
			s += indent
			s += m.styles["unselectedEntry"].Render(todaysEntries[i].View()) + "\n"
		}
		if m.editActive {
			s += indent + strings.Repeat(" ", 12)
			s += m.ViewAsEdit() + "\n"
		} else {
			s += indent
			s += m.styles["selectedEntry"].Render(todaysEntries[m.currentSelectedRow].View()) + "\n"

		}
		for i := m.currentSelectedRow + 1; i < len(todaysEntries); i++ {
			s += indent
			s += m.styles["unselectedEntry"].Render(todaysEntries[i].View()) + "\n"
		}
	}

	s += "\n\n\n\n"
	s += "\n\n#######\nDebug: " + m.debugMessage + "\n#######\n\n"

	s += m.help.View(m.keys)
	// s += strings.Repeat("\n", m.height - strings.Count(s, "\n"))
	return s
	// return lipgloss.NewStyle().Background(tint.Bg()).Height(m.height).Width(m.width).Padding(2).Render(s)
}

func main() {
	tint.NewDefaultRegistry()
	tint.SetTint(tint.TintAfterglow)
	// tint.SetTint(tint.TintArthur)

	var debug_file io.Writer
	debug_file, err := os.OpenFile("./run.log", os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0600)
	if err != nil {
		fmt.Println("Could not create log file: ", err)
	}
	slog.SetDefault(slog.New(slog.NewTextHandler(debug_file, nil)))
	slog.SetLogLoggerLevel(slog.LevelInfo)
	slog.Info("Starting Excel-Editor...")

	var resultChan = make(chan tea.Model)
	l := NewLoadingScreen(resultChan)
	go func() {
		resultChan <- initialModel()
	}()

	p := tea.NewProgram(l, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Println("An error occurred: ", err)
		os.Exit(1)
	}
}
