package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

type Board struct {
	Width  int
	Height int
}

type Model struct {
	Board       Board
	Body        *strings.Builder
	err         error
	prompt      textinput.Model
	Table       table
	startTime   time.Time
	logFile     string
	isWhiteTurn bool
}

type (
	errMsg error
)

type table map[[2]int]rune

func createInitialTableMap(width, height int) table {
	m := make(table)
	if width >= 3 && height >= 1 {
		m[[2]int{width - 3, 0}] = BlackHorse
		m[[2]int{width - 2, 0}] = BlackTower
		m[[2]int{width - 1, 0}] = BlackKing
	}
	if width >= 3 && height >= 1 {
		m[[2]int{0, height - 1}] = WhiteKing
		m[[2]int{1, height - 1}] = WhiteTower
		m[[2]int{2, height - 1}] = WhiteHorse
	}
	return m
}

const TLC = '\u250C' // ┌ top left corner
const TRC = '\u2510' // ┐ top right corner
const BLC = '\u2514' // └ bottom left corner
const BRC = '\u2518' // ┘ bottom right corner
const HL = '\u2500'  // ─ Horizontal Line
const VL = '\u2502'  // │ Vertical Line
const CR = '\u253C'  // ┼ Cross
const EC = ' '       //   Empty Cell
const RC = '\u2524'  // ┤ Right Cell
const LC = '\u251C'  // ├ Left Cell
const TC = '\u252C'  // ┬ Top Cell
const BC = '\u2534'  // ┴ Bottom Cell
const EOL = "\n"     // End of Line

const WhiteHorse = '\u2658' // ♘ White Horse (Unicode chess knight)
const WhiteTower = '\u2656' // ♖ White Tower (Unicode chess rook)
const WhiteKing = '\u2654'  // ♔ White King  (Unicode chess king)
const BlackHorse = '\u265E' // ♞ Black Horse (Unicode black knight)
const BlackTower = '\u265C' // ♜ Black Tower (Unicode black rook)
const BlackKing = '\u265A'  // ♚ Black King  (Unicode black king)

const minValidSize = 6
const maxValidSize = 12

/*
 * Game messages
 */

const welcomeMessage = "Welcome to xxxxxx Chess"
const initialBoardMessage = "\nSelect board size to start.\nValues must be between 6 and 12 on each dimension.\n\n"
const promptWidthMsg = "Enter Board width (X): "
const promptHeightMsg = "Enter Board height (Y): "
const promptContinueMsg = "\nType a command (type help for options): \n> "
const invalidInputMsg = "\n\nInvalid input. Please enter a number between %d and %d.\n"
const creatingBoardMsg = "\n\nCreating board of size %dx%d\n"
const invalidCommandMsg = "\n\nUnknown command. Type 'help' for available commands.\n"
const invalidCoordinatesMsg = "\n\nInvalid coordinates.\n"
const noPieceMsg = "\n\nNo piece at the source coordinate.\n"
const whiteTurnMsg = "\n\nIt's White's turn. You can only move white pieces.\n"
const blackTurnMsg = "\n\nIt's Black's turn. You can only move black pieces.\n"
const unknownPieceMsg = "\n\nUnknown piece type.\n"
const invalidMoveMsg = "\n\nInvalid move for this piece type.\n"
const cannotCaptureSelfMsg = "\n\nCannot capture your own piece.\n"
const moveUsageMsg = "\n\nUsage: move <from> <to>\n"
const gameEndedMsg = "Game ended by player"
const gameOverMsg = "Game Over!"
const gameOverThanksMsg = "\n\nGame Over! Thanks for playing!"
const blackWinsMsg = "⬛ Black wins! 🎉"
const whiteWinsMsg = "⬜ White wins! 🎉"
const gameResetMsg = "Game reset"
const whiteTurnIndicator = "\n\n⬜ Turn: White\n"
const blackTurnIndicator = "\n\n⬛ Turn: Black\n"

const helpMessage = `

Available commands:
  move <from> <to>       Move a piece (e.g. move B1 C3)
  restart                Restart the match
  exit                   Exit the game
  help                   Show this list`

func main() {
	ti := textinput.New()
	ti.Prompt = promptWidthMsg
	ti.CharLimit = 20
	ti.Width = 20

	p := tea.NewProgram(Model{
		Board: Board{
			Width:  0,
			Height: 0,
		},
		Body:        new(strings.Builder),
		prompt:      ti,
		Table:       nil,
		startTime:   time.Time{},
		logFile:     "",
		isWhiteTurn: true,
	})

	if _, err := p.Run(); err != nil {
		fmt.Printf("Error starting program: %v\n", err)
	}
}

/*
 * bubble definitions
 */

func (m Model) View() string {
	var sb strings.Builder

	sb.WriteString(m.Body.String())

	return sb.String()
}

func (m Model) Init() tea.Cmd {
	// needed by model.
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {

		case tea.KeyCtrlC, tea.KeyEsc:
			return m, tea.Quit

		case tea.KeyEnter:
			if m.Board.Width == 0 || m.Board.Height == 0 {
				if m.Board.Width == 0 {
					w, err := strconv.Atoi(m.prompt.Value())
					if err != nil || !validateBoardSize(w) {
						if !strings.Contains(m.Body.String(), invalidInputMsg) {
							m.Body.WriteString(fmt.Sprintf(invalidInputMsg, minValidSize, maxValidSize))
						}
						m.prompt.SetValue("")
						return m, cmd
					} else {
						m.Board.Width = w
						m.prompt.SetValue("")
					}
				} else if m.Board.Height == 0 {
					h, err := strconv.Atoi(m.prompt.Value())
					if err != nil || !validateBoardSize(h) {
						if !strings.Contains(m.Body.String(), invalidInputMsg) {
							m.Body.WriteString(fmt.Sprintf(invalidInputMsg, minValidSize, maxValidSize))
						}
						m.prompt.SetValue("")
						return m, cmd
					} else {
						m.Board.Height = h
						m.prompt.SetValue("")
						m.Body.WriteString(fmt.Sprintf(creatingBoardMsg, m.Board.Width, m.Board.Height))
						m.Table = createInitialTableMap(m.Board.Width, m.Board.Height)
						m.startTime = time.Now()
						m.logFile = m.createNewLogFile()

						writeToHistory(fmt.Sprintf("Game started with board size %dx%d\n", m.Board.Width, m.Board.Height), m.logFile)
					}
				}
			} else {
				switch strings.Split(strings.ToLower(m.prompt.Value()), " ")[0] {
				case "restart":
					m = resetGame(m)
					return m, cmd

				case "exit":
					if m.logFile != "" {
						writeToHistory(gameEndedMsg, m.logFile)
					}
					return m, tea.Quit

				case "help", "h":
					if !strings.Contains(m.Body.String(), helpMessage) {
						m.Body.WriteString(helpMessage)
					}
					return m, cmd

				case "move", "mv":
					base := strings.Split(strings.ToLower(m.prompt.Value()), " ")
					if len(base) < 3 {
						m.Body.WriteString(moveUsageMsg)
						m.prompt.SetValue("")
						return m, cmd
					}
					from := base[1]
					to := base[2]
					m, msg := movePiece(from, to, m)
					if msg != "" {
						m.Body.WriteString(msg)
					}
					m.prompt.SetValue("")
					return m, cmd

				default:
					if !strings.Contains(m.Body.String(), invalidCommandMsg) {
						m.Body.WriteString(invalidCommandMsg)
					}
					return m, cmd
				}
			}
		}
	case errMsg:
		m.err = msg
		return m, nil
	}

	m.prompt, cmd = m.prompt.Update(msg)

	if m.Board.Width == 0 || m.Board.Height == 0 {
		m.Body.Reset()

		m.Body.Write([]byte(drawBoxMessage(welcomeMessage)))
		m.Body.Write([]byte(initialBoardMessage))

		if m.Board.Width != 0 {
			m.prompt.Prompt = promptHeightMsg
		} else {
			m.prompt.Prompt = promptWidthMsg
		}

		m.Body.WriteString(m.prompt.View())
		m.prompt.Focus()
	} else {
		m.Body.Reset()
		m.Body.WriteString("\n\n")
		m.Body.WriteString(drawTableWithMap(m.Board.Height, m.Board.Width, m.Table))

		if m.isWhiteTurn {
			m.Body.WriteString(whiteTurnIndicator)
		} else {
			m.Body.WriteString(blackTurnIndicator)
		}

		m.prompt.Prompt = promptContinueMsg
		m.Body.WriteString(m.prompt.View())
		m.prompt.Focus()
	}

	return m, cmd
}

/*
 * Validations
 */

func isValidKingMove(fromCol, fromRow, toCol, toRow int) bool {
	colDiff := abs(toCol - fromCol)
	rowDiff := abs(toRow - fromRow)
	return colDiff <= 1 && rowDiff <= 1 && !(colDiff == 0 && rowDiff == 0)
}

func isValidTowerMove(fromCol, fromRow, toCol, toRow int) bool {
	colDiff := abs(toCol - fromCol)
	rowDiff := abs(toRow - fromRow)

	straightMove := (colDiff == 0 && rowDiff > 0 && rowDiff <= 3) || (rowDiff == 0 && colDiff > 0 && colDiff <= 3)
	diagonalMove := (colDiff == rowDiff) && colDiff > 0 && colDiff <= 3

	return straightMove || diagonalMove
}

func isValidHorseMove(fromCol, fromRow, toCol, toRow int) bool {
	colDiff := abs(toCol - fromCol)
	rowDiff := abs(toRow - fromRow)

	return (colDiff == 2 && rowDiff == 1) || (colDiff == 1 && rowDiff == 2)
}

func validateCoordinate(coord string, m Model) bool {
	height := m.Board.Height
	width := m.Board.Width

	if len(coord) != 2 {
		return false
	}

	col := coord[0]
	row := coord[1]

	col = strings.ToUpper(string(col))[0]
	if col < 'A' || col >= byte('A'+width) {
		return false
	}

	if row < '1' || row > byte('0'+height) {
		return false
	}
	return true
}

func validateBoardSize(side int) bool {
	if side < minValidSize || side > maxValidSize {
		return false
	}

	return true
}

func isWhitePiece(piece rune) bool {
	return piece == WhiteKing || piece == WhiteTower || piece == WhiteHorse
}

func isBlackPiece(piece rune) bool {
	return piece == BlackKing || piece == BlackTower || piece == BlackHorse
}

/*
 * helpers
 */

func abs(x int) int {
	if x < 0 {
		return -x
	}

	return x
}

func (m *Model) createNewLogFile() string {
	historyDir := "history"
	if err := os.MkdirAll(historyDir, 0755); err != nil {
		return ""
	}

	filename := fmt.Sprintf("game_%02d_%02d_%02d_%02d_%02d.txt",
		m.startTime.Day(), m.startTime.Month(),
		m.startTime.Hour(), m.startTime.Minute(), m.startTime.Second())

	return filepath.Join(historyDir, filename)
}

func writeToHistory(message string, logFile string) error {
	if logFile == "" {
		return fmt.Errorf("no log file specified")
	}

	historyDir := "history"

	if err := os.MkdirAll(historyDir, 0755); err != nil {
		return err
	}

	f, err := os.OpenFile(logFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)

	if err != nil {
		return err
	}

	defer f.Close()

	if _, err := f.WriteString(message); err != nil {
		return err
	}

	return nil
}

/*
 * game handlers
 */

func movePiece(from, to string, m Model) (Model, string) {
	if !validateCoordinate(from, m) || !validateCoordinate(to, m) {
		return m, invalidCoordinatesMsg
	}
	fromCol := int(strings.ToUpper(string(from[0]))[0] - 'A')
	fromRow := m.Board.Height - int(from[1]-'0')
	toCol := int(strings.ToUpper(string(to[0]))[0] - 'A')
	toRow := m.Board.Height - int(to[1]-'0')
	piece, ok := m.Table[[2]int{fromCol, fromRow}]

	if !ok || piece == EC {
		return m, noPieceMsg
	}

	if m.isWhiteTurn && !isWhitePiece(piece) {
		return m, whiteTurnMsg
	}
	if !m.isWhiteTurn && !isBlackPiece(piece) {
		return m, blackTurnMsg
	}

	var validMove bool

	switch piece {
	case WhiteKing, BlackKing:
		validMove = isValidKingMove(fromCol, fromRow, toCol, toRow)
	case WhiteTower, BlackTower:
		validMove = isValidTowerMove(fromCol, fromRow, toCol, toRow)
	case WhiteHorse, BlackHorse:
		validMove = isValidHorseMove(fromCol, fromRow, toCol, toRow)
	default:
		return m, unknownPieceMsg
	}

	if !validMove {
		return m, invalidMoveMsg
	}

	captured := m.Table[[2]int{toCol, toRow}]

	if captured != 0 && captured != EC {
		if m.isWhiteTurn && isWhitePiece(captured) {
			return m, cannotCaptureSelfMsg
		}
		if !m.isWhiteTurn && isBlackPiece(captured) {
			return m, cannotCaptureSelfMsg
		}
	}

	msg := ""
	isGameOver := false
	if captured != 0 && captured != EC {
		msg = fmt.Sprintf("Moved %c from %s to %s. Captured %c \n", piece, from, to, captured)

		if captured == WhiteKing {
			msg += drawBoxMessage(fmt.Sprintf("blackWinsMsg"))
			isGameOver = true
		} else if captured == BlackKing {
			msg += drawBoxMessage(fmt.Sprintf("whiteWinsMsg"))
			isGameOver = true
		}
	} else {
		msg = fmt.Sprintf("Moved %c from %s to %s.", piece, from, to)
	}

	writeToHistory(msg, m.logFile)

	if isGameOver {
		writeToHistory(gameOverMsg, m.logFile)
	}

	delete(m.Table, [2]int{fromCol, fromRow})
	m.Table[[2]int{toCol, toRow}] = piece
	m.Body.Reset()
	m.Body.WriteString("\n\n")
	m.Body.WriteString(drawTableWithMap(m.Board.Height, m.Board.Width, m.Table))

	if isGameOver {
		m.Body.WriteString("\n\n")
		m.Body.WriteString(msg)
		m.Body.WriteString(gameOverThanksMsg)
		m.prompt.SetValue("")
		m.prompt.Prompt = promptContinueMsg
		m.Body.WriteString(m.prompt.View())
		m.prompt.Focus()
		return m, ""
	}

	m.isWhiteTurn = !m.isWhiteTurn

	if m.isWhiteTurn {
		m.Body.WriteString(whiteTurnIndicator)
	} else {
		m.Body.WriteString(blackTurnIndicator)
	}

	m.prompt.SetValue("")
	m.prompt.Prompt = promptContinueMsg
	m.Body.WriteString(m.prompt.View())
	m.prompt.Focus()

	if msg != "" {
		m.Body.WriteString("\n\n" + msg)
	}

	return m, ""
}

func resetGame(m Model) Model {
	if m.logFile != "" {
		writeToHistory(gameResetMsg, m.logFile)
	}

	width := m.Board.Width
	height := m.Board.Height

	m.Body.Reset()
	m.Table = createInitialTableMap(width, height)
	m.startTime = time.Now()
	m.logFile = m.createNewLogFile()
	m.isWhiteTurn = true

	writeToHistory(fmt.Sprintf("Game started with board size %dx%d\n", width, height), m.logFile)

	m.Body.WriteString("\n\n")
	m.Body.WriteString(drawTableWithMap(height, width, m.Table))
	m.Body.WriteString(whiteTurnIndicator)
	m.prompt.SetValue("")
	m.prompt.Prompt = promptContinueMsg

	return m
}

func getCellValue(x, y int, t table) rune {
	if piece, ok := t[[2]int{x, y}]; ok {
		return piece
	}

	return EC
}

/*
 * drawings
 */

func drawTableWithMap(height, width int, t table) string {
	var tableBuilder strings.Builder

	buildTableTopLine(width, &tableBuilder)
	buildTableMiddleLineWithMap(width, height, &tableBuilder, t)
	buildTableBottomLine(width, &tableBuilder)

	return tableBuilder.String()
}

func drawBoxMessage(msg string) string {
	padding := 5
	emojiPad := 0
	contentLen := len(msg) + padding*2

	if strings.Contains(msg, "⬛") || strings.Contains(msg, "⬜") {
		emojiPad = 1
		padding += 1
	}

	var box strings.Builder

	box.WriteString(fmt.Sprintf("%c%s%c%s", TLC, strings.Repeat(string(HL), contentLen), TRC, EOL))
	box.WriteString(fmt.Sprintf("%c%s%s%s%c%s", VL, strings.Repeat(" ", padding), msg, strings.Repeat(" ", padding+emojiPad), VL, EOL))
	box.WriteString(fmt.Sprintf("%c%s%c%s", BLC, strings.Repeat(string(HL), contentLen), BRC, EOL))

	return box.String()
}

func drawTable(height, width int) string {
	table := strings.Builder{}

	buildTableTopLine(width, &table)
	buildTableMiddleLine(width, height, &table)
	buildTableBottomLine(width, &table)

	return table.String()
}

func buildTableTopLine(width int, table *strings.Builder) {
	table.WriteString(string("    "))
	for i := 0; i < width; i++ {
		table.WriteString(fmt.Sprintf("  %c ", 'A'+i))
	}

	table.WriteString(string(EOL))
	table.WriteString(string("    "))
	table.WriteString(string(TLC))

	for i := 0; i < width; i++ {
		table.WriteString(strings.Repeat(string(HL), 3))
		if i < width-1 {
			table.WriteString(string(TC))
		} else {
			table.WriteString(string(TRC))
			table.WriteString(string(EOL))
		}
	}
}

func buildTableMiddleLine(width, height int, tableBuilder *strings.Builder) {
	chars := []struct {
		left, center, right, accross rune
	}{
		{VL, EC, VL, VL},
		{LC, HL, RC, CR},
	}

	t := createInitialTableMap(width, height)

	for h := 0; h < height*2-1; h++ {
		if h%2 == 0 {
			tableBuilder.WriteString(fmt.Sprintf(" %2d ", height-h/2))
		} else {
			tableBuilder.WriteString("    ")
		}

		tableBuilder.WriteString(string(chars[h%2].left))

		for w := 0; w < width; w++ {
			if h%2 == 0 {
				tableBuilder.WriteString(fmt.Sprintf(" %c ", getCellValue(w, h/2, t)))
			} else {
				tableBuilder.WriteString(strings.Repeat(string(HL), 3))
			}
			if w == width-1 {
				tableBuilder.WriteString(string(chars[h%2].right))
				tableBuilder.WriteString(string(EOL))
			} else {
				tableBuilder.WriteString(string(chars[h%2].accross))
			}
		}
	}
}

func buildTableMiddleLineWithMap(width, height int, tableBuilder *strings.Builder, t table) {
	chars := []struct {
		left, center, right, accross rune
	}{
		{VL, EC, VL, VL},
		{LC, HL, RC, CR},
	}

	for h := 0; h < height*2-1; h++ {
		if h%2 == 0 {
			tableBuilder.WriteString(fmt.Sprintf(" %2d ", height-h/2))
		} else {
			tableBuilder.WriteString("    ")
		}
		tableBuilder.WriteString(string(chars[h%2].left))
		for w := 0; w < width; w++ {
			if h%2 == 0 {
				y := h / 2
				tableBuilder.WriteString(fmt.Sprintf(" %c ", getCellValue(w, y, t)))
			} else {
				tableBuilder.WriteString(strings.Repeat(string(HL), 3))
			}
			if w == width-1 {
				tableBuilder.WriteString(string(chars[h%2].right))
				tableBuilder.WriteString(string(EOL))
			} else {
				tableBuilder.WriteString(string(chars[h%2].accross))
			}
		}
	}
}

func buildTableBottomLine(width int, table *strings.Builder) {
	table.WriteString(string("    "))
	table.WriteString(string(BLC))

	for i := 0; i < width; i++ {
		table.WriteString(strings.Repeat(string(HL), 3))
		if i < width-1 {
			table.WriteString(string(BC))
		} else {
			table.WriteString(string(BRC))
			table.WriteString(string(EOL))
		}
	}
}
