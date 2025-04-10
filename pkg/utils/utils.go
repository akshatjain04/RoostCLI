package utils

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"reflect"
	"runtime"
	"strconv"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/viper"
)

type PromptContent struct {
	ErrorMsg string
	Label    string
}

type RoostIoLoginResponse struct {
	// AsPlaintText=true WILL NOT REDACT the content.
	AsPlainText    bool                  `json:"-"`
	Username       string                `json:"username"`
	ID             string                `json:"id"`
	FirstName      string                `json:"firstName"`
	LastName       string                `json:"lastName"`
	Company        string                `json:"company"`
	Location       string                `json:"location"`
	PhotoUrl       string                `json:"photoUrl"`
	Email          string                `json:"email"`
	IsActive       bool                  `json:"isActive"`
	TotalServices  string                `json:"totalServices"`
	Certifications string                `json:"certifications"`
	ResidentSince  string                `json:"residentSince"`
	RoostSessions  string                `json:"roostSessions"`
	Bio            string                `json:"bio"`
	ExpiresIn      string                `json:"expiresIn"`
	AccessToken    string                `json:"accessToken"`
	ThirdPartyApps []ThirdPartyAppConfig `json:"thirdPartyApps"`
}

type ThirdPartyAppConfig struct {
	Thirdparty_app_id string `json:"thirdparty_app_id"`
	AppUserID         string `json:"app_user_id"`
	DisplayName       string `json:"display_name"`
}

func FileOrFolderExists(filename string) bool {
	_, err := os.Stat(filename)
	return !os.IsNotExist(err)
}

func HTTPClientRequest(operation, command, authKey string, params io.Reader) (int, []byte, error) {
	client := &http.Client{}
	url := "https://" + viper.Get("roost_ent_server").(string) + command
	req, err := http.NewRequest(operation, url, params)
	if err != nil {
		return http.StatusBadRequest, nil, errors.New("Failed to create HTTP request." + err.Error())
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	if authKey != "" {
		req.Header.Set("Authorization", authKey)
	}
	if strings.Contains(command, "git/events/add") {
		req.Header.Set("token-type", "on-demand")
	}
	resp, err := client.Do(req)
	if err != nil {
		return http.StatusBadRequest, nil, err
	}
	defer resp.Body.Close()

	body, ioErr := io.ReadAll(resp.Body)
	return resp.StatusCode, body, ioErr
}

/*
AcceptFromPrompt is an utility function which accepts default request data. Prompts user to get it modified if needed.
// to: must be pointer to struct with exported fields.
// Supported types are int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, ,float32, float64, string
*/
func AcceptFromPrompt(to any) error {
	var err error
	v := reflect.Indirect(reflect.ValueOf(to))
	t := reflect.TypeOf(v)
	if !v.CanSet() {
		return fmt.Errorf("can't update field from prompt. Pass reference of struct in function call to allow updation")
	}

	if t.Kind() == reflect.Struct {
		err = PromptTextInput(&v)
	}
	return err

}

const listHeight = 14

var (
	titleStyle        = lipgloss.NewStyle().MarginLeft(2)
	itemStyle         = lipgloss.NewStyle().PaddingLeft(4)
	selectedItemStyle = lipgloss.NewStyle().PaddingLeft(2).Foreground(lipgloss.Color("50"))
	paginationStyle   = list.DefaultStyles().PaginationStyle.PaddingLeft(4)
	helpStyle         = list.DefaultStyles().HelpStyle.PaddingLeft(4).PaddingBottom(1)
	quitTextStyle     = lipgloss.NewStyle().Margin(1, 0, 2, 4)
)

type item string

func (i item) FilterValue() string { return "" }

type itemDelegate struct{}

func (d itemDelegate) Height() int                               { return 1 }
func (d itemDelegate) Spacing() int                              { return 0 }
func (d itemDelegate) Update(msg tea.Msg, m *list.Model) tea.Cmd { return nil }
func (d itemDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	i, ok := listItem.(item)
	if !ok {
		return
	}

	str := fmt.Sprintf("%d. %s", index+1, i)

	fn := itemStyle.Render
	if index == m.Index() {
		fn = func(s string) string {
			return selectedItemStyle.Render("> " + s)
		}
	}

	fmt.Fprint(w, fn(str))
}

type modelselect struct {
	list         list.Model
	choice       string
	quitting     bool
	clusteralias string
}

func (m modelselect) Init() tea.Cmd {
	return nil
}

func (m modelselect) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.list.SetWidth(msg.Width)
		return m, nil

	case tea.KeyMsg:
		switch keypress := msg.String(); keypress {
		case "ctrl+c":
			m.quitting = true
			return m, tea.Quit

		case "enter":
			i, ok := m.list.SelectedItem().(item)
			if ok {
				m.choice = string(i)
			}
			return m, tea.Quit
		}
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m modelselect) View() string {
	if m.choice != "" {
		return quitTextStyle.Render(fmt.Sprintf("Selected option is %s", m.choice))
	}
	if m.quitting {
		return quitTextStyle.Render("No option selected")
	}
	return "\n" + m.list.View()
}

func PromptSelectInput(custtoken []string, msg string) string {

	const defaultWidth = 20
	items := []list.Item{}

	for _, clusteralias := range custtoken {
		items = append(items, item(clusteralias))
	}

	l := list.New(items, itemDelegate{}, defaultWidth, listHeight)
	l.Title = msg
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)
	l.Styles.Title = titleStyle
	l.Styles.PaginationStyle = paginationStyle
	l.Styles.HelpStyle = helpStyle

	m := modelselect{list: l}

	x, err := tea.NewProgram(m).Run()
	if err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
	}
	if m, ok := x.(modelselect); ok && m.choice != "" {
		return m.choice
	}

	return ""
}

var (
	focusedStyle        = lipgloss.NewStyle().Foreground(lipgloss.Color("50"))
	blurredStyle        = lipgloss.NewStyle().Foreground(lipgloss.Color("250"))
	cursorStyle         = focusedStyle.Copy()
	noStyle             = lipgloss.NewStyle()
	helpStylePrompt     = blurredStyle.Copy()
	cursorModeHelpStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("50"))

	focusedButton = focusedStyle.Copy().Render("[ Submit ]")
	blurredButton = fmt.Sprintf("[ %s ]", blurredStyle.Render("Submit"))
)

type promptmodel struct {
	focusIndex int
	inputs     []textinput.Model
	cursorMode textinput.CursorMode
	quitting   bool
}

func initialModel(promptfields *reflect.Value) promptmodel {
	m := promptmodel{
		inputs: make([]textinput.Model, promptfields.NumField()),
	}

	var t textinput.Model
	for i := range m.inputs {
		t = textinput.New()
		t.CursorStyle = cursorStyle
		t.CharLimit = 32

		fieldName := promptfields.Type().Field(i).Name
		t.Prompt = "> " + fieldName + ": "
		if i == 0 {
			t.Focus()
			t.PromptStyle = focusedStyle
			t.TextStyle = focusedStyle
		}
		var Placeholder string
		if promptfields.Field(i).Type().Name() == "int" {
			Placeholder = fmt.Sprintf("%d", promptfields.Field(i).Int())
		} else {
			Placeholder = promptfields.Field(i).String()
		}

		t.Placeholder = " " + Placeholder

		m.inputs[i] = t

	}

	return m
}

func (m promptmodel) Init() tea.Cmd {
	return textinput.Blink
}

func (m promptmodel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "esc":
			m.quitting = true
			return m, tea.Quit

		// Change cursor mode
		case "ctrl+r":
			m.cursorMode++
			if m.cursorMode > textinput.CursorHide {
				m.cursorMode = textinput.CursorBlink
			}
			cmds := make([]tea.Cmd, len(m.inputs))
			for i := range m.inputs {
				cmds[i] = m.inputs[i].SetCursorMode(m.cursorMode)
			}
			return m, tea.Batch(cmds...)

		// Set focus to next input
		case "tab", "shift+tab", "enter", "up", "down":
			s := msg.String()

			// Did the user press enter while the submit button was focused?
			// If so, exit.
			if s == "enter" && m.focusIndex == len(m.inputs) {
				return m, tea.Quit
			}

			// Cycle indexes
			if s == "up" || s == "shift+tab" {
				m.focusIndex--
			} else {
				m.focusIndex++
			}

			if m.focusIndex > len(m.inputs) {
				m.focusIndex = 0
			} else if m.focusIndex < 0 {
				m.focusIndex = len(m.inputs)
			}

			cmds := make([]tea.Cmd, len(m.inputs))
			for i := 0; i <= len(m.inputs)-1; i++ {
				if i == m.focusIndex {
					// Set focused state
					cmds[i] = m.inputs[i].Focus()
					m.inputs[i].PromptStyle = focusedStyle
					m.inputs[i].TextStyle = focusedStyle
					continue
				}
				// Remove focused state
				m.inputs[i].Blur()
				m.inputs[i].PromptStyle = noStyle
				m.inputs[i].TextStyle = noStyle
			}

			return m, tea.Batch(cmds...)
		}
	}

	// Handle character input and blinking
	cmd := m.updateInputs(msg)

	return m, cmd
}

func (m *promptmodel) updateInputs(msg tea.Msg) tea.Cmd {
	cmds := make([]tea.Cmd, len(m.inputs))

	// Only text inputs with Focus() set will respond, so it's safe to simply
	// update all of them here without any further logic.
	for i := range m.inputs {
		m.inputs[i], cmds[i] = m.inputs[i].Update(msg)
	}

	return tea.Batch(cmds...)
}

func (m promptmodel) View() string {
	var b strings.Builder

	for i := range m.inputs {
		b.WriteString(m.inputs[i].View())
		if i < len(m.inputs)-1 {
			b.WriteRune('\n')
		}
	}

	button := &blurredButton
	if m.focusIndex == len(m.inputs) {
		button = &focusedButton
	}
	fmt.Fprintf(&b, "\n\n%s\n\n", *button)

	b.WriteString(helpStylePrompt.Render("cursor mode is "))
	b.WriteString(cursorModeHelpStyle.Render(m.cursorMode.String()))
	b.WriteString(helpStylePrompt.Render(" (ctrl+r to change style)"))

	return b.String()
}

func PromptTextInput(promptvalue *reflect.Value) error {

	x, err := tea.NewProgram(initialModel(promptvalue)).Run()
	if err != nil {
		fmt.Printf("could not start program: %s\n", err)
		os.Exit(1)
	}

	if x.(promptmodel).quitting == true {
		return fmt.Errorf(":Prompt Exit")
	}

	m, _ := x.(promptmodel)
	{
		for i, promptinput := range m.inputs {
			field := promptvalue.Field(i)

			if promptinput.Value() == "" {
				validateinput(&field, field.String())
			} else {
				validateinput(&field, promptinput.Value())
			}

		}

	}

	return nil
}

func validateinput(field *reflect.Value, promptValue string) error {
	fieldKind := field.Kind()

	switch fieldKind {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		d, err := strconv.ParseInt(promptValue, 0, 64)
		if err != nil {
			return fmt.Errorf("string to int64 conversion error %q", err.Error())
		}
		field.SetInt(reflect.ValueOf(d).Int())

	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		d, err := strconv.ParseUint(promptValue, 0, 64)
		if err != nil {
			return fmt.Errorf("string to uint64 conversion error %q", err.Error())
		}
		field.SetUint(reflect.ValueOf(d).Uint())

	case reflect.Float32, reflect.Float64:
		d, err := strconv.ParseFloat(promptValue, 64)
		if err != nil {
			return fmt.Errorf("string to float64 conversion error %q", err.Error())
		}
		field.SetFloat(reflect.ValueOf(d).Float())

	case reflect.String:
		field.Set(reflect.ValueOf(promptValue))

	default:
		return fmt.Errorf("unsupported field type provided %q. supports int, int64, uint64, float64, string", fieldKind)
	}
	return nil
}

func Openbrowser(url string) error {
	var err error

	switch runtime.GOOS {
	case "linux":
		err = exec.Command("xdg-open", url).Start()
	case "windows":
		err = exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
	case "darwin":
		err = exec.Command("open", url).Start()
	default:
		err = fmt.Errorf("unsupported platform")
	}

	return err
}

var baseStyle = lipgloss.NewStyle().
	BorderStyle(lipgloss.NormalBorder()).
	BorderForeground(lipgloss.Color("240"))

type modeltable struct {
	table table.Model
	choice table.Row
}

func (m modeltable) Init() tea.Cmd { return nil }

func (m modeltable) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			if m.table.Focused() {
				m.table.Blur()
			} else {
				m.table.Focus()
			}
		case "q", "ctrl+c":
			return m, tea.Quit
		case "enter":
			m.choice=m.table.SelectedRow()
			return m,tea.Quit
		}
	}
	m.table, cmd = m.table.Update(msg)
	return m, cmd
}

func (m modeltable) View() string {
	return baseStyle.Render(m.table.View()) + "\n"
}

func TableInput(columninput []table.Column,rowinput[]table.Row)table.Row{
	columns := columninput

	rows := rowinput

	t := table.New(
		table.WithColumns(columns),
		table.WithRows(rows),
		table.WithFocused(true),
		table.WithHeight(7),
	)

	s := table.DefaultStyles()
	s.Header = s.Header.
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("240")).
		BorderBottom(true).
		Bold(false)
	s.Selected = s.Selected.
		Foreground(lipgloss.Color("50")).
		Background(lipgloss.Color("240")).
		Bold(true)
	t.SetStyles(s)

	m := modeltable{t,table.Row{}}
	x, err := tea.NewProgram(m).Run(); 
	if err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
	}

	if m, ok := x.(modeltable); ok {
	return m.choice
	}


	return table.Row{}
}