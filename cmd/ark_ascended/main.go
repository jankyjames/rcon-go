package main

import (
	"os"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/jankyjames/rcon-go"
	"github.com/rivo/tview"
	"github.com/sirupsen/logrus"
)

var (
	app   = tview.NewApplication()
	debug = func() bool {
		switch strings.ToLower(os.Getenv("DEBUG")) {
		case "1", "true", "t", "yes", "y":
			return true
		default:
			return false
		}
	}()
	logrusView = func() *tview.TextView {
		logView := tview.NewTextView().SetScrollable(true)
		logView.SetTitle("Logs")
		logView.SetChangedFunc(func() {
			app.Draw()
			logView.ScrollToEnd()
		})
		logrus.StandardLogger().SetOutput(logView)
		return logView
	}()
)

func main() {
	if debug {
		logrus.SetLevel(logrus.DebugLevel)
	}

	hostnameInput := tview.NewInputField().
		SetLabel("Hostname").
		SetText("").
		SetFieldWidth(20)
	passwordInput := tview.NewInputField().
		SetLabel("Password").
		SetText("").
		SetFieldWidth(20).
		SetMaskCharacter('*')

	form := tview.NewForm().AddFormItem(hostnameInput).AddFormItem(passwordInput).
		AddButton("Connect", func() {
			app.SetRoot(consoleView(hostnameInput.GetText(), passwordInput.GetText()), true)
		})
	form.SetBorder(true)
	if err := app.SetRoot(form, true).EnableMouse(true).Run(); err != nil {
		panic(err)
	}
}

func consoleView(host, password string) tview.Primitive {
	client := rcon.New(host, password)
	responseView := tview.NewTextView()

	input := tview.NewInputField().SetLabel("Command: ")
	inactiveColor := tcell.ColorGray
	activeColor := tcell.ColorBlue
	input.SetDoneFunc(func(key tcell.Key) {
		commandString := input.GetText()
		input.SetFieldBackgroundColor(inactiveColor)
		input.SetText("")
		input.SetDisabled(true)
		go func() {
			do, err := client.Do(commandString)
			if err != nil {
				panic(err.Error())
			}
			app.QueueUpdateDraw(func() {
				responseView.SetText(do)
				input.SetDisabled(false)
				input.SetFieldBackgroundColor(activeColor)
			})
		}()
	})

	grid := tview.NewGrid().
		SetRows(0, 1).
		SetBorders(true).
		AddItem(responseView, 0, 0, 1, 1, 0, 0, false).
		AddItem(input, 1, 0, 1, 1, 0, 0, true)

	if debug {
		grid.SetColumns(-1, -1).
			AddItem(logrusView, 0, 1, 2, 1, 0, 0, false)
	}

	return grid
}
