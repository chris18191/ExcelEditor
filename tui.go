package main

import (
	"fmt"
	"log/slog"
	"maps"
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

var WEEKDAYS = []string{"So", "Mo", "Di", "Mi", "Do", "Fr", "Sa"}

var debugConfig = Configuration{
	ExcelFileName:       "res/test.xlsx",
	COL_ID_DATE:         0,
	COL_ID_HOURS_START:  2,
	COL_ID_HOURS_END:    3,
	COL_ID_HOURS_PAUSE:  4,
	ROW_ID_ENTRY_START:  6, // sixth row contains first entries
	OutputFile:          "./res/result.xlsx",
	ProjectNumbersSheet: "Projektnummern",
}

type EntryList struct {
	Entries [][][]RowEntry
}

func NewEntryList(config Configuration) EntryList {
	return EntryList{
		Entries: ReturnAll(config),
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
	config           Configuration
	numColumns       int // equals the number of columns that are printed for each row
	spinner          spinner.Model
	datepicker       DatePicker
	entryList        EntryList
	debugMessage     string
	projectNames     map[string]Project
	projectNumbers   map[string]Project
	projectCustomers map[string]Project

	keys               keyMap
	help               help.Model
	currentSelectedRow int

	editActive   bool
	textInputs   []textinput.Model
	focusedIndex int

	styles map[string]lipgloss.Style
	height int
	width  int

	projectNumberIndex        int
	projectNumberVisible      int
	potentialProjects         []Project
	lastProjectNumberSearched string
}

func initialModel(config Configuration) Model {

	help := help.New()
	// help.Styles.FullDesc = help.Styles.FullDesc.Background(tint.Bg())
	// help.Styles.ShortDesc = help.Styles.ShortDesc.Background(tint.Bg())
	// help.Styles.FullKey = help.Styles.FullKey.Background(tint.Bg())
	// help.Styles.ShortKey = help.Styles.ShortKey.Background(tint.Bg())
	// help.Styles.FullSeparator = help.Styles.FullSeparator.Background(tint.Bg())
	// help.Styles.ShortSeparator = help.Styles.ShortSeparator.Background(tint.Bg())
	// help.Styles.Ellipsis = help.Styles.Ellipsis.Background(tint.Bg())

	nr, name, custom := GetProjectNumbers(config)

	return Model{
		datepicker: NewDatePicker(),
		entryList:  NewEntryList(config),
		config:     config,

		projectNumbers:   nr,
		projectNames:     name,
		projectCustomers: custom,

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

		projectNumberIndex:   0,
		projectNumberVisible: 10,

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
			"dailySum": lipgloss.NewStyle().
				Bold(true).
				Foreground(tint.BrightCyan()),
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
	return (a + b) % b
}

func readTextInputWithDefault(i *textinput.Model) string {
	if i.Value() != "" {
		return i.Value()
	}
	return i.Placeholder
}

func trySettingCurrentSelectedProjectNr(m *Model) {
	if m.focusedIndex == 4 {
		if m.projectNumberIndex < len(m.potentialProjects) {
			m.textInputs[4].SetValue(m.potentialProjects[m.projectNumberIndex].ID)
			m.potentialProjects = []Project{m.potentialProjects[m.projectNumberIndex]}
		} else {
			m.potentialProjects = []Project{}
		}
	}
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, keys.PrevDay) && !m.editActive:
			if m.datepicker.currentDay.Month() == time.January && m.datepicker.currentDay.Day() == 1 {
				m.debugMessage = "Reached first day of the year!"
				break
			}
			m.datepicker.currentDay = m.datepicker.currentDay.AddDate(0, 0, -1)
			if m.currentSelectedRow >= len(*m.getCurrentDayEntries()) {
				m.currentSelectedRow -= 1
			}
			if m.currentSelectedRow < 0 {
				m.currentSelectedRow = 0
			}
			// m.currentSelectedRow = int(math.Min(float64(m.currentSelectedRow), float64(len(*m.getCurrentDayEntries())-1)))
			// m.entryList.UpdateCurrentIndex(m.datepicker.currentDay)
		case key.Matches(msg, keys.NextDay) && !m.editActive:
			if m.datepicker.currentDay.Month() == time.December && m.datepicker.currentDay.Day() == 31 {
				m.debugMessage = "Reached last day of the year!"
				break
			}
			m.datepicker.currentDay = m.datepicker.currentDay.AddDate(0, 0, 1)
			if m.currentSelectedRow >= len(*m.getCurrentDayEntries()) {
				m.currentSelectedRow -= 1
			}
			if m.currentSelectedRow < 0 {
				m.currentSelectedRow = 0
			}
			// m.currentSelectedRow = int(math.Min(float64(m.currentSelectedRow), float64(len(*m.getCurrentDayEntries())-1)))
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
			var sheets = make(map[string][][]RowEntry)
			for _, month := range m.entryList.Entries {
				if len(month) > 0 {
					if len(month[0]) > 0 {
						sheets[month[0][0].SheetName] = month
					} else if len(month[2]) > 0 {
						sheets[month[2][0].SheetName] = month
					}
				}
			}
			WriteRowEntries(sheets, m.config)
		case key.Matches(msg, keys.FocusPrev):
			if !m.editActive {
				break
			}
			trySettingCurrentSelectedProjectNr(&m)
			m.focusedIndex = helperMod(m.focusedIndex-1, len(m.textInputs))
			m.debugMessage = fmt.Sprintf("Focused index: %d", m.focusedIndex)
		case key.Matches(msg, keys.FocusNext):
			if !m.editActive {
				break
			}
			trySettingCurrentSelectedProjectNr(&m)
			m.focusedIndex = helperMod(m.focusedIndex+1, len(m.textInputs))
			m.debugMessage = fmt.Sprintf("Focused index: %d", m.focusedIndex)

		case key.Matches(msg, keys.Delete) && !m.editActive && len(m.entryList.Entries) > 0:
			m.debugMessage = "Trying to delete..."
			todaysEntries := m.getCurrentDayEntries()
			if len(*todaysEntries) == 0 {
				m.debugMessage += " No entry!"
				break
			}
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
			var newEntry RowEntry
			if len(*todaysEntries) > 0 {
				newEntry = RowEntry{
					Date:      (*todaysEntries)[len(*todaysEntries)-1].Date,
					Day:       (*todaysEntries)[len(*todaysEntries)-1].Day,
					SheetName: (*todaysEntries)[len(*todaysEntries)-1].SheetName,
					Start:     (*todaysEntries)[len(*todaysEntries)-1].End,
				}
			}
			*todaysEntries = append(*todaysEntries, newEntry)
			m.currentSelectedRow = len(*todaysEntries) - 1
			fallthrough // automatically edit new entry
		case key.Matches(msg, keys.Edit):
			m.debugMessage = "Pressed edit..."
			todaysEntries := *m.getCurrentDayEntries()
			if len(todaysEntries) == 0 {
				m.debugMessage += "No entry!"
				break
			}
			m.editActive = !m.editActive
			entry := (todaysEntries)[m.currentSelectedRow]
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
						t.Placeholder = "Description"
						t.SetValue(entry.Description)
						// if t.Placeholder = entry.Description; t.Placeholder == "" {
						// }
						t.Width = 40
					case 4:
						if t.Placeholder = entry.ProjectNr; t.Placeholder == "" {
							t.Placeholder = "Project-Nr."
						}
						t.CharLimit = 9
						t.Width = 9
					default:
						t.Placeholder = "UNDEFINED FIELD"
					}

					inputs[i] = t
				}
				m.textInputs = inputs
				m.focusedIndex = 0
			} else {
				m.debugMessage = "Saved entry starting at " + m.textInputs[0].Value()
				entry.Date = m.datepicker.currentDay
				entry.Day = WEEKDAYS[int(entry.Date.Weekday())]
				entry.Start, _ = time.Parse("15:04", readTextInputWithDefault(&m.textInputs[0]))
				entry.End, _ = time.Parse("15:04", readTextInputWithDefault(&m.textInputs[1]))
				entry.Pause, _ = time.ParseDuration(readTextInputWithDefault(&m.textInputs[2]))
				entry.Description = readTextInputWithDefault(&m.textInputs[3])
				trySettingCurrentSelectedProjectNr(&m)
				entry.ProjectNr = readTextInputWithDefault(&m.textInputs[4])
				entry.Project = m.projectNumbers[entry.ProjectNr].Name
				slog.Info("Trying to set project information...", "entry", entry)
			}
			m.entryList.Entries[m.datepicker.currentDay.Month()-1][m.datepicker.currentDay.Day()-1][m.currentSelectedRow] = entry

		case key.Matches(msg, keys.ArrowUp) && m.editActive && m.focusedIndex == 4:
			m.projectNumberIndex = helperMod(m.projectNumberIndex-1, len(m.potentialProjects))
			m.debugMessage = fmt.Sprintf("Arrow up. Project index %d/%d", m.projectNumberIndex, len(m.potentialProjects))
		case key.Matches(msg, keys.ArrowDown) && m.editActive && m.focusedIndex == 4:
			m.projectNumberIndex = helperMod(m.projectNumberIndex+1, len(m.potentialProjects))
			m.debugMessage = fmt.Sprintf("Arrow down. Project index %d/%d", m.projectNumberIndex, len(m.potentialProjects))
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
		if m.focusedIndex == 4 {

			if m.lastProjectNumberSearched != m.textInputs[4].Value() {
				m.lastProjectNumberSearched = m.textInputs[4].Value()
				var potentialProjects []Project
				for p := range maps.Values(m.projectNumbers) {
					if len(potentialProjects) >= m.projectNumberVisible {
						break
					}
					if strings.HasPrefix(p.ID, m.textInputs[4].Value()) {
						potentialProjects = append(potentialProjects, p)
					}
					if strings.Contains(p.Name, m.textInputs[4].Value()) {
						potentialProjects = append(potentialProjects, p)
					}
					if strings.Contains(p.Customer, m.textInputs[4].Value()) {
						potentialProjects = append(potentialProjects, p)
					}
				}
				m.potentialProjects = potentialProjects
			}
		}
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
	return fmt.Sprintf("%10s  %8s → %-8s [%9s] %20.20s :  %20.20s",
		r.Date.Format("Mon 02.01."),
		r.Start.Format("15:04"),
		r.End.Format("15:04"),
		r.Pause.String(),
		r.Project,
		r.Description,
	)
}

func (m *Model) ViewAsEdit() string {
	s := ""
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
	}

	return m.styles["inputField"].Render(s)
}

func (m Model) View() string {
	s := ""
	s += m.styles["header"].Render("Work Hour Editor")
	s += "\n"
	s += fmt.Sprintf("Current Date: [%12s] \n", m.datepicker.currentDay.Format("Mon 02.01.06"))
	s += fmt.Sprintf("\n")

	s += m.styles["tableHeader"].Render(
		fmt.Sprintf(" %-10s  %-8s   %-8s %-9s   %-20s    %-20s", "Date", "Start", "End", "Pause", "Project", "Description"),
	)
	s += "\n"

	indent := "  "
	todaysEntries := *m.getCurrentDayEntries()
	var totalWorkDay time.Duration = time.Duration(0)
	if len(todaysEntries) <= 0 {

	} else {
		for i := 0; i < m.currentSelectedRow; i++ {
			s += indent
			s += m.styles["unselectedEntry"].Render(todaysEntries[i].View()) + "\n"
			totalWorkDay += todaysEntries[i].End.Sub(todaysEntries[i].Start)
		}
		if m.editActive {
			s += indent + strings.Repeat(" ", 12)
			inputRow := m.ViewAsEdit()
			s += inputRow + "\n"

			if m.focusedIndex == 4 && m.textInputs[4].Value() != "" {

				indent := strings.Repeat(" ", 10)
				s += "\n"

				for i := 0; i < len(m.potentialProjects); i++ {
					if i == m.projectNumberIndex {
						s += m.styles["selectedEntry"].Render(fmt.Sprintf("%s%s %s: %-50.50s [%20.20s]", indent, "◉", m.potentialProjects[i].ID, m.potentialProjects[i].Name, m.potentialProjects[i].Customer)) + "\n"
					} else {
						s += m.styles["selectedEntry"].Render(fmt.Sprintf("%s%s %s: %50.50s [%20.20s]", indent, "○", m.potentialProjects[i].ID, m.potentialProjects[i].Name, m.potentialProjects[i].Customer)) + "\n"
					}
				}

			}

		} else {
			if m.currentSelectedRow >= len(todaysEntries) {
				m.debugMessage = "Current row > entries length"
			} else {
				s += indent
				s += m.styles["selectedEntry"].Render(todaysEntries[m.currentSelectedRow].View()) + "\n"
				totalWorkDay += todaysEntries[m.currentSelectedRow].End.Sub(todaysEntries[m.currentSelectedRow].Start)
			}
		}
		for i := m.currentSelectedRow + 1; i < len(todaysEntries); i++ {
			s += indent
			s += m.styles["unselectedEntry"].Render(todaysEntries[i].View()) + "\n"
			totalWorkDay += todaysEntries[i].End.Sub(todaysEntries[i].Start)
		}
	}

	s += "\n"

	totalWorkDay = totalWorkDay.Round(time.Duration(1) * time.Minute)

	s += m.styles["dailySum"].Render(fmt.Sprintf("Total hours: %02.0f:%02d", totalWorkDay.Hours(), int(totalWorkDay.Minutes())%60))

	s += "\n\n"
	s += "\n\n#######\nDebug: " + m.debugMessage + "\n#######\n\n"

	s += m.help.View(m.keys)
	return s
}

func Start(config Configuration) {
	// Set up color scheme
	tint.NewDefaultRegistry()
	tint.SetTint(tint.TintAfterglow)

	var resultChan = make(chan tea.Model)
	l := NewLoadingScreen(resultChan)
	go func() {
		resultChan <- initialModel(config)
	}()

	p := tea.NewProgram(l, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Println("An error occurred: ", err)
		os.Exit(1)
	}
}
