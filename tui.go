package main

import (
	"bytes"
	"fmt"
	"io"
	"path/filepath"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	tuiPrimary = lipgloss.AdaptiveColor{Light: "#5A56E0", Dark: "#7D7AFF"}
	tuiSuccess = lipgloss.AdaptiveColor{Light: "#047857", Dark: "#5EEAD4"}
	tuiWarning = lipgloss.AdaptiveColor{Light: "#B45309", Dark: "#FBBF24"}
	tuiMuted   = lipgloss.AdaptiveColor{Light: "#6B7280", Dark: "#9CA3AF"}
	tuiDanger  = lipgloss.AdaptiveColor{Light: "#B42318", Dark: "#FF8A80"}
	tuiBorder  = lipgloss.AdaptiveColor{Light: "#D1D5DB", Dark: "#3F3F46"}
)

var tuiSections = []string{"Overview", "Files", "Duplicates", "Security", "Organize", "Undo"}

type tuiModel struct {
	directory string
	active    int
	width     int
	height    int
	title     string
	content   []string
	scroll    int
	status    string
	modal     string // "", "folder", "move", or "undo"
	input     string
	loading   bool
	loadText  string
	loadFrame int
}

type tuiLoadedMsg struct {
	section int
	model   tuiModel
}

type tuiTickMsg struct{}

func runTUI(in io.Reader, out io.Writer) error {
	directory, err := defaultDirectory()
	if err != nil {
		return err
	}
	model := newTUIModel(directory)
	program := tea.NewProgram(model, tea.WithInput(in), tea.WithOutput(out), tea.WithAltScreen())
	_, err = program.Run()
	return err
}

func newTUIModel(directory string) tuiModel {
	return tuiModel{directory: directory, title: "Overview", status: "Ready", loading: true, loadText: "Scanning your Downloads"}
}

func (m tuiModel) Init() tea.Cmd { return tea.Batch(m.loadCmd(), tuiSpinnerCmd()) }

func (m tuiModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tuiLoadedMsg:
		if msg.section != m.active {
			return m, nil
		}
		m.content, m.title, m.status = msg.model.content, msg.model.title, msg.model.status
		m.scroll, m.loading = 0, false
		return m, nil
	case tuiTickMsg:
		if m.loading {
			m.loadFrame = (m.loadFrame + 1) % 4
			return m, tuiSpinnerCmd()
		}
		return m, nil
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
		return m, nil
	case tea.KeyMsg:
		if m.loading && msg.String() != "q" && msg.String() != "ctrl+c" {
			return m, nil
		}
		if m.modal != "" {
			return m.updateModal(msg)
		}
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "up", "k":
			m.active = (m.active + len(tuiSections) - 1) % len(tuiSections)
			return m.beginLoad("Opening " + tuiSections[m.active])
		case "down", "j", "tab":
			m.active = (m.active + 1) % len(tuiSections)
			return m.beginLoad("Opening " + tuiSections[m.active])
		case "1", "2", "3", "4", "5", "6":
			m.active = int(msg.String()[0] - '1')
			return m.beginLoad("Opening " + tuiSections[m.active])
		case "ctrl+f", "pgdown", "right":
			m.scroll = min(m.scroll+max(3, m.height/2), max(0, len(m.content)-1))
			return m, nil
		case "ctrl+b", "pgup", "left":
			m.scroll = max(0, m.scroll-max(3, m.height/2))
			return m, nil
		case "home":
			m.scroll = 0
			return m, nil
		case "end":
			m.scroll = max(0, len(m.content)-1)
			return m, nil
		case "enter", "r":
			return m.beginLoad("Refreshing " + tuiSections[m.active])
		case "d":
			m.modal, m.input = "folder", m.directory
			return m, nil
		case "o":
			m.modal, m.input = "move", ""
			return m, nil
		case "u":
			m.modal, m.input = "undo", ""
			return m, nil
		case "?":
			m.status = "j/k panels • ←/→ scroll • enter refresh • d folder • o organize • u undo • q quit"
			return m, nil
		}
	}
	return m, nil
}

func (m tuiModel) updateModal(key tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch key.String() {
	case "esc":
		m.modal, m.input = "", ""
		m.status = "Action cancelled."
		return m, nil
	case "backspace":
		if len(m.input) > 0 {
			m.input = m.input[:len(m.input)-1]
		}
		return m, nil
	case "enter":
		switch m.modal {
		case "folder":
			directory, err := resolveTargetDirectory([]string{m.input})
			if err != nil {
				m.status = "Directory error: " + err.Error()
			} else {
				m.directory = directory
				m.status = "Folder changed to " + directory
				m.modal, m.input = "", ""
				return m.beginLoad("Scanning the new folder")
			}
		case "move":
			if m.input != "MOVE" {
				m.status = "Not organized: type MOVE exactly to confirm."
			} else {
				var out bytes.Buffer
				err := applyOrganizationTo(m.directory, &out)
				m.status = operationStatus(out.String())
				if err != nil {
					m.status = "Error: " + err.Error() + "\n" + m.status
				}
				m.modal, m.input = "", ""
				return m.beginLoad("Refreshing organization preview")
			}
		case "undo":
			if m.input != "UNDO" {
				m.status = "Not undone: type UNDO exactly to confirm."
			} else {
				var out bytes.Buffer
				err := undoOrganizationTo(&out)
				m.status = operationStatus(out.String())
				if err != nil {
					m.status = "Error: " + err.Error() + "\n" + m.status
				}
			}
		}
		m.modal, m.input = "", ""
		return m, nil
	default:
		if len(key.Runes) > 0 {
			m.input += string(key.Runes)
		}
	}
	return m, nil
}

func (m tuiModel) beginLoad(label string) (tea.Model, tea.Cmd) {
	m.loading, m.loadText, m.loadFrame = true, label, 0
	return m, tea.Batch(m.loadCmd(), tuiSpinnerCmd())
}

func (m tuiModel) loadCmd() tea.Cmd {
	section := m.active
	return func() tea.Msg {
		started := time.Now()
		loaded := m.loadSection()
		if remaining := 350*time.Millisecond - time.Since(started); remaining > 0 {
			time.Sleep(remaining)
		}
		return tuiLoadedMsg{section: section, model: loaded}
	}
}

func tuiSpinnerCmd() tea.Cmd {
	return tea.Tick(90*time.Millisecond, func(time.Time) tea.Msg { return tuiTickMsg{} })
}

func (m tuiModel) loadSection() tuiModel {
	m.title = tuiSections[m.active]
	m.scroll = 0
	switch m.active {
	case 0, 1:
		result, err := scanResult(m.directory)
		if err != nil {
			m.content, m.status = []string{"Unable to scan this folder."}, "Error: "+err.Error()
			return m
		}
		if m.active == 0 {
			m.content = []string{
				metricLine("Items", fmt.Sprint(result.ItemCount), "Files and folders at this level"),
				metricLine("File size", humanSize(result.TotalBytes), "Total size of direct files"),
				metricLine("Folders", fmt.Sprint(len(result.Directories)), "Directories are never moved automatically"),
				"",
				"Largest files",
			}
			for _, file := range limitFiles(result.Files, 8) {
				m.content = append(m.content, fmt.Sprintf("%-10s %s", humanSize(file.Size), file.Name))
			}
		} else {
			m.content = []string{"Directories"}
			for _, name := range result.Directories {
				m.content = append(m.content, "▸ "+name)
			}
			m.content = append(m.content, "", "Files by size")
			for _, file := range limitFiles(result.Files, 18) {
				m.content = append(m.content, fmt.Sprintf("%-10s %s", humanSize(file.Size), file.Name))
			}
		}
	case 2:
		result, err := duplicateResult(m.directory)
		if err != nil {
			m.content, m.status = []string{"Unable to compare files."}, "Error: "+err.Error()
			return m
		}
		m.content = []string{fmt.Sprintf("%d duplicate groups • %s reclaimable", len(result.Groups), humanSize(result.ReclaimableBytes)), ""}
		for i, group := range result.Groups {
			m.content = append(m.content, fmt.Sprintf("Group %d  •  %s each", i+1, humanSize(group[0].Size)))
			for _, file := range group {
				m.content = append(m.content, "  "+file.Name)
			}
		}
		if len(result.Groups) == 0 {
			m.content = append(m.content, "No exact duplicate files found.")
		}
	case 3:
		result, err := securityResult(m.directory)
		if err != nil {
			m.content, m.status = []string{"Unable to run the safety review."}, "Error: "+err.Error()
			return m
		}
		m.content = []string{fmt.Sprintf("%d files need review", len(result.Findings)), ""}
		for _, finding := range result.Findings {
			m.content = append(m.content, fmt.Sprintf("[%3d] %s", finding.Score, finding.Name), "      "+strings.Join(finding.Reasons, " • "))
		}
		if len(result.Findings) == 0 {
			m.content = append(m.content, "No files need review.")
		}
	case 4:
		plans, err := organizationPlans(m.directory)
		if err != nil {
			m.content, m.status = []string{"Unable to create an organization preview."}, "Error: "+err.Error()
			return m
		}
		m.content = []string{fmt.Sprintf("%d files match organization rules", len(plans)), ""}
		for _, plan := range plans {
			destination, _ := filepath.Rel(m.directory, plan.Destination)
			m.content = append(m.content, fmt.Sprintf("%-35s → %s", filepath.Base(plan.Source), destination))
		}
		m.content = append(m.content, "", "Press o to organize. You must type MOVE to confirm.")
	case 5:
		m.content = []string{"Undo reverses only the latest successful organization.", "", "Press u to undo. You must type UNDO to confirm."}
	}
	return m
}

func (m tuiModel) View() string {
	if m.width == 0 {
		return "Loading Download Inbox Cleaner…"
	}
	sidebarWidth := 24
	contentWidth := max(40, m.width-sidebarWidth-5)
	header := lipgloss.NewStyle().Bold(true).Foreground(tuiPrimary).Render("Download Inbox Cleaner") + "  " +
		lipgloss.NewStyle().Foreground(tuiMuted).Render("safe Downloads workspace")
	path := lipgloss.NewStyle().Foreground(tuiMuted).Render("Folder  " + m.directory)

	menu := make([]string, 0, len(tuiSections)+2)
	menu = append(menu, lipgloss.NewStyle().Bold(true).Foreground(tuiMuted).Render("WORKSPACE"))
	for i, section := range tuiSections {
		item := fmt.Sprintf(" %d  %s", i+1, section)
		if i == m.active {
			item = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#FFFFFF")).Background(tuiPrimary).Width(sidebarWidth - 2).Render(item)
		} else {
			item = lipgloss.NewStyle().Foreground(tuiMuted).Render(item)
		}
		menu = append(menu, item)
	}
	menu = append(menu, "", lipgloss.NewStyle().Foreground(tuiMuted).Render("d  change folder"))
	sidebar := lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(tuiBorder).Padding(1, 1).Width(sidebarWidth).Render(strings.Join(menu, "\n"))

	maxLines := max(5, m.height-12)
	visible := []string{}
	if m.loading {
		frames := []string{"◐", "◓", "◑", "◒"}
		visible = []string{"", "", lipgloss.NewStyle().Bold(true).Foreground(tuiPrimary).Render(frames[m.loadFrame] + "  " + m.loadText), lipgloss.NewStyle().Foreground(tuiMuted).Render("Please wait — checking files safely.")}
	} else {
		start := min(m.scroll, max(0, len(m.content)-1))
		end := min(len(m.content), start+maxLines)
		visible = append([]string{}, m.content[start:end]...)
		if start > 0 {
			visible = append([]string{lipgloss.NewStyle().Foreground(tuiMuted).Render("↑ more above")}, visible...)
		}
		if end < len(m.content) {
			visible = append(visible, lipgloss.NewStyle().Foreground(tuiMuted).Render("↓ more below  (←/→ to scroll)"))
		}
	}
	lines := append([]string{lipgloss.NewStyle().Bold(true).Foreground(tuiPrimary).Render(m.title), ""}, visible...)
	content := lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(tuiBorder).Padding(1, 2).Width(contentWidth).Height(max(12, m.height-8)).Render(strings.Join(lines, "\n"))
	statusColor := tuiMuted
	if strings.HasPrefix(m.status, "Error:") {
		statusColor = tuiDanger
	} else if strings.HasPrefix(m.status, "Moved ") || strings.HasPrefix(m.status, "Organization undone") {
		statusColor = tuiSuccess
	} else if strings.HasPrefix(m.status, "Not ") {
		statusColor = tuiWarning
	}
	status := lipgloss.NewStyle().Foreground(statusColor).Render(m.status)
	footer := lipgloss.NewStyle().Foreground(tuiMuted).Render("j/k panels • ←/→ scroll • enter refresh • d folder • o organize • u undo • q quit")
	view := lipgloss.JoinVertical(lipgloss.Left, header, path, "", lipgloss.JoinHorizontal(lipgloss.Top, sidebar, " ", content), "", status, footer)
	if m.modal != "" {
		prompt := "Folder path"
		if m.modal == "move" {
			prompt = "Type MOVE to organize files"
		}
		if m.modal == "undo" {
			prompt = "Type UNDO to restore the latest organization"
		}
		dialog := lipgloss.NewStyle().Border(lipgloss.DoubleBorder()).BorderForeground(tuiPrimary).Padding(1, 2).Width(58).Render(
			lipgloss.NewStyle().Bold(true).Render(prompt) + "\n\n" + m.input + "▏\n\n" + lipgloss.NewStyle().Foreground(tuiMuted).Render("enter confirm • esc cancel"))
		return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, dialog)
	}
	return view
}

func metricLine(label, value, detail string) string {
	return fmt.Sprintf("%-12s %-10s %s", label, value, detail)
}

func limitFiles(files []File, limit int) []File {
	if len(files) <= limit {
		return files
	}
	return files[:limit]
}

func humanSize(size int64) string {
	if size < 1024 {
		return fmt.Sprintf("%d B", size)
	}
	if size < 1024*1024 {
		return fmt.Sprintf("%.1f KB", float64(size)/1024)
	}
	if size < 1024*1024*1024 {
		return fmt.Sprintf("%.1f MB", float64(size)/(1024*1024))
	}
	return fmt.Sprintf("%.1f GB", float64(size)/(1024*1024*1024))
}

func operationStatus(output string) string {
	lines := strings.Split(strings.TrimSpace(output), "\n")
	for i := len(lines) - 1; i >= 0; i-- {
		if strings.TrimSpace(lines[i]) != "" {
			return strings.TrimSpace(lines[i])
		}
	}
	return "Action completed."
}
