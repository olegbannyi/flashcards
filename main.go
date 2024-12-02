package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"math/rand"
	"os"
	"strconv"
	"strings"
)

const notFound = -1

type App struct {
	cards       []Card
	menu        map[string]func()
	appLog      []string
	importFile  string
	exmportFile string
}

type Card struct {
	Term       string
	Definition string
	Mistakes   int
}

func main() {
	NewApp().run()

}

func NewApp() *App {
	app := App{
		cards:  []Card{},
		menu:   map[string]func(){},
		appLog: []string{}}

	app.menu["add"] = app.addHandler
	app.menu["remove"] = app.removeHandler
	app.menu["import"] = app.importHandler
	app.menu["export"] = app.exportHandler
	app.menu["ask"] = app.askHandler
	app.menu["exit"] = app.exitHandler
	app.menu["log"] = app.logHandler
	app.menu["hardest card"] = app.hardestCardHandler
	app.menu["reset stats"] = app.resetStatsHandler

	flag.StringVar(&app.importFile, "import_from", "", "Import data on app launch")
	flag.StringVar(&app.exmportFile, "export_to", "", "Export data on app exit")

	flag.Parse()

	return &app
}

func (a *App) run() {
	if a.importFile != "" {
		a.doImport(a.importFile)
	}

	reader := bufio.NewReader(os.Stdin)
	for {
		a.Println("Input the action (add, remove, import, export, ask, exit, log, hardest card, reset stats):")
		action, _ := reader.ReadString('\n')
		a.writeLog(action)
		action = strings.TrimSpace(action)

		if handler, ok := a.menu[action]; ok {
			handler()
		}
	}
}

func (a *App) Println(line string) {
	fmt.Println(line)
	a.writeLog(line)
}

func (a *App) Printf(template string, params ...interface{}) {
	line := fmt.Sprintf(template, params...)
	fmt.Print(line)
	a.writeLog(line)
}

func (a *App) writeLog(line string) {
	if line[len(line)-1] != '\n' {
		line += "\n"
	}
	a.appLog = append(a.appLog, line)
}

func (a *App) addHandler() {
	var term, definition string
	reader := bufio.NewReader(os.Stdin)

	a.Println("The card:")

	for {
		term, _ = reader.ReadString('\n')
		a.writeLog(string(term))
		term = strings.TrimSpace(term)

		if notFound != a.findCardByTerm(term) {
			a.Printf("The term \"%s\" already exists. Try again:\n", term)
		} else {
			break
		}
	}

	a.Println("The definition of the card:")

	for {
		definition, _ = reader.ReadString('\n')
		a.writeLog(string(definition))
		definition = strings.TrimSpace(definition)

		if notFound != a.findCardByDefinition(definition) {
			a.Printf("The definition \"%s\" already exists. Try again:\n", definition)
		} else {
			break
		}
	}

	a.cards = append(a.cards, Card{Term: term, Definition: definition, Mistakes: 0})

	a.Printf("The pair (\"%s\":\"%s\") has been added.\n", term, definition)
}

func (a *App) removeHandler() {
	a.Println("Which card?")
	var term string
	fmt.Scan(&term)
	a.writeLog("> " + term)

	index := a.findCardByTerm(term)

	if notFound == index {
		a.Printf("Can't remove \"%s\": there is no such card.\n", term)
	} else {
		switch index {
		case 0:
			a.cards = a.cards[1:]
		case len(a.cards) - 1:
			a.cards = a.cards[:len(a.cards)-1]
		default:
			a.cards = append(a.cards[:index], a.cards[index+1:]...)
		}
		a.Println("The card has been removed.")
	}
}

func (a *App) importHandler() {
	a.Println("File name:")
	var filename string
	fmt.Scan(&filename)
	a.writeLog("> " + filename)

	a.doImport(filename)
}

func (a *App) doImport(filename string) {
	if content, err := os.ReadFile(filename); err != nil {
		a.Println("File not found.")
	} else {
		var cards []Card
		if err := json.Unmarshal(content, &cards); err != nil {
			a.Println(err.Error())
		} else {
			a.cards = cards
			a.cards = make([]Card, len(cards))
			copy(a.cards, cards)
			a.Printf("%d cards have been loaded.\n", len(cards))
		}
	}
}

func (a *App) exportHandler() {
	a.Println("File name:")
	var filename string
	fmt.Scan(&filename)
	a.writeLog("> " + filename)

	a.doExport(filename)
}

func (a *App) doExport(filename string) {
	if content, err := json.Marshal(a.cards); err != nil {
		a.Println(err.Error())
	} else if err = os.WriteFile(filename, content, 0644); err != nil {
		a.Println(err.Error())
	} else {
		a.Printf("%d cards have been saved.\n", len(a.cards))
	}
}

func (a *App) askHandler() {
	a.Println("How many times to ask?")
	var num int
	fmt.Scan(&num)
	a.writeLog("> " + strconv.Itoa(num))

	var answer string
	reader := bufio.NewReader(os.Stdin)

	for i := 0; i < num; i++ {
		index := rand.Intn(len(a.cards))
		card := a.cards[index]

		a.Printf("Print the definition of \"%s\"\n", card.Term)

		answer, _ = reader.ReadString('\n')
		a.writeLog(string(answer))
		answer = strings.TrimSpace(answer)

		if answer == card.Definition {
			a.Println("Correct!")
		} else {
			a.cards[index].Mistakes++
			if n := a.findCardByDefinition(answer); n != notFound {
				a.Printf("Wrong. The right answer is \"%s\", but your definition is correct for \"%s\".\n",
					card.Definition, a.cards[n].Term)
			} else {
				a.Printf("Wrong. The right answer is \"%s\".\n", card.Definition)
			}
		}

	}
}

func (a *App) exitHandler() {
	if a.exmportFile != "" {
		a.doExport(a.exmportFile)
	}

	a.Println("Bye bye!")
	os.Exit(0)
}

func (a *App) logHandler() {
	a.Println("File name:")
	var filename string
	fmt.Scan(&filename)
	a.writeLog("> " + filename)

	a.Println("The log has been saved.")
	os.WriteFile(filename, []byte(strings.Join(a.appLog, "")), 0644)
}

func (a *App) hardestCardHandler() {
	max := 0

	for _, card := range a.cards {
		if card.Mistakes > max {
			max = card.Mistakes
		}
	}

	hardestTerms := []string{}
	for _, card := range a.cards {
		if card.Mistakes == max {
			hardestTerms = append(hardestTerms, "\""+card.Term+"\"")
		}
	}

	if max == 0 {
		a.Println("There are no cards with errors.")
	} else if len(hardestTerms) == 1 {
		errorStr := "error"
		if max > 1 {
			errorStr = "errors"
		}
		a.Printf("The hardest card is %s. You have %d %s answering it.\n", hardestTerms[0], max, errorStr)
	} else {
		a.Printf("The hardest cards are %s.\n", strings.Join(hardestTerms, ", "))
	}
}

func (a *App) resetStatsHandler() {
	for i := range a.cards {
		a.cards[i].Mistakes = 0
	}
	a.Println("Card statistics have been reset.")
}

func (a *App) findCardByTerm(term string) int {
	for i, card := range a.cards {
		if card.Term == term {
			return i
		}
	}
	return notFound
}

func (a *App) findCardByDefinition(definition string) int {
	for i, card := range a.cards {
		if card.Definition == definition {
			return i
		}
	}
	return notFound
}
