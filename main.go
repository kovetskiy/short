package main

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/big"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/docopt/docopt-go"
	"github.com/nsf/termbox-go"
)

const (
	usage = `Short 1.0, short term memory tester.

Usage:
    ./short [options]

Options:
    -f <file>     use specified file as database [default: ~/.config/short-term].
    -n <number>   show specified count of tests [default: 20].
    -c <count>    show specified count of numbers in tests [default: 7].
    -i <min>      use specified number as minimum value of number [default: 10]
    -a <max>      use specified number as maximum value of number [default: 99]
`
)

// test result
type Result struct {
	Score    int     `json:"score"`
	Duration float64 `json:"duration"`
	Count    int     `json:"count"`
}

func main() {
	args, _ := docopt.Parse(usage, nil, true, "1.0", false)

	file := args["-f"].(string)
	if file[:2] == "~/" {
		file = os.Getenv("HOME") + file[1:]
	}

	var (
		testsCount, _   = strconv.Atoi(args["-n"].(string))
		numbersCount, _ = strconv.Atoi(args["-c"].(string))
		minNumber, _    = strconv.Atoi(args["-i"].(string))
		maxNumber, _    = strconv.Atoi(args["-a"].(string))
	)

	err := termbox.Init()
	if err != nil {
		panic(err)
	}

	clearScreen()

	results := []Result{}
	for i := 0; i < testsCount; i++ {
		result := runTest(minNumber, maxNumber, numbersCount)
		results = append(results, result)
	}

	var (
		sumScore    int
		sumDuration float64
	)

	for _, result := range results {
		sumScore += result.Score
		sumDuration += result.Duration
	}

	avgDuration := sumDuration / float64(len(results))
	avgScore := float64(sumScore) / float64(len(results))

	termbox.Close()

	fmt.Printf("Score: %.2f (%.2f sec)\n", avgScore, avgDuration)

	saveResults(file, results, sumScore, avgDuration)
}

func saveResults(
	file string, results []Result, totalScore int, avgDuration float64,
) {
	type DatabaseItem struct {
		Date        string   `json:"date"`
		AvgDuration float64  `json:"avg_duration"`
		TotalScore  int      `json:"total_score"`
		Results     []Result `json:"results"`
	}

	fd, err := os.OpenFile(file, os.O_RDWR|os.O_CREATE, 0600)
	if err != nil {
		panic(err)
	}
	defer fd.Close()
	content, err := ioutil.ReadAll(fd)
	if err != nil {
		panic(err)
	}

	database := []DatabaseItem{}
	json.Unmarshal(content, &database)

	database = append(database, DatabaseItem{
		Date:        time.Now().String(),
		AvgDuration: avgDuration,
		TotalScore:  totalScore,
		Results:     results,
	})

	content, err = json.Marshal(database)
	if err != nil {
		panic(err)
	}

	fd.WriteAt(content, 0)
}

func runTest(minNumber, maxNumber, numbersCount int) Result {
	validNumbers := generateRandomNumbers(
		minNumber, maxNumber, numbersCount,
	)

	numberStrings := []string{}
	for _, number := range validNumbers {
		numberStrings = append(numberStrings, strconv.Itoa(number))
	}

	wholeTest := strings.Join(numberStrings, " ")

	width, height := termbox.Size()

	timeStart := time.Now()

	x := width/2 - len(wholeTest)/2
	y := height / 2

	termbox.SetCursor(x, y)

	for _, symbol := range wholeTest {
		x += 1
		termbox.SetCell(
			x, y, symbol, termbox.ColorDefault, termbox.ColorDefault,
		)
	}

	termbox.HideCursor()
	termbox.Flush()

	wait() //wait for input 'Enter'

	timeFinish := time.Now()

	clearScreen()

	termbox.SetCursor(x-len(wholeTest)+1, y)
	termbox.Flush()
	userNumbers := getNumbers(x-len(wholeTest), y)

	clearScreen()

	score := compare(validNumbers, userNumbers)
	duration := timeFinish.Sub(timeStart).Seconds()

	return Result{
		score, duration, numbersCount,
	}
}

func generateRandomNumbers(min, max, count int) []int {
	numbers := []int{}
	for i := 0; i < count; i++ {
		bigNumber, _ := rand.Int(rand.Reader, big.NewInt(int64(max)))
		number := int(bigNumber.Int64())
		if number < min {
			i--
			continue
		}

		numbers = append(numbers, number)
	}

	return numbers
}

func getNumbers(x, y int) []int {
	numbers := []int{}
	text := readText(x, y)

	pieces := strings.Split(text, " ")
	for _, piece := range pieces {
		number, _ := strconv.Atoi(piece)
		numbers = append(numbers, number)
	}

	return numbers
}

func readText(x, y int) string {
	text := ""
	for {
		event := termbox.PollEvent()
		if event.Type != termbox.EventKey {
			continue
		}

		if event.Ch >= '0' && event.Ch <= '9' {
			text += string(event.Ch)
		}

		switch event.Key {
		case termbox.KeySpace:
			text += " "
		case termbox.KeyBackspace2:
			if len(text) == 0 {
				break
			}
			text = text[0 : len(text)-1]
			clearScreen()
			printText(text, x, y)
		case termbox.KeyEnter:
			return text
		case termbox.KeyCtrlC, termbox.KeyCtrlZ:
			termbox.Close()
			os.Exit(0)
		}

		printText(text, x, y)
	}
}

func compare(validNumbers, inputNumbers []int) (score int) {
	length := len(inputNumbers)
	if len(validNumbers) < length {
		length = len(validNumbers)
	}

	for index := 0; index < length; index++ {
		if validNumbers[index] == inputNumbers[index] {
			score++
		} else {
			break
		}
	}

	return score
}

func clearScreen() {
	termbox.Clear(termbox.ColorDefault, termbox.ColorDefault)
	err := termbox.Flush()
	if err != nil {
		panic(err)
	}
}

// just wait for any user input (like 'Press Enter to continue')
func wait() {
	for {
		event := termbox.PollEvent()
		if event.Type != termbox.EventKey {
			continue
		}

		switch event.Key {
		case termbox.KeyEnter:
			return
		case termbox.KeyCtrlC, termbox.KeyCtrlZ:
			termbox.Close()
			os.Exit(0)
		}
	}
}

func printText(text string, x, y int) {
	termbox.SetCursor(x, y)

	for _, symbol := range text {
		x += 1
		termbox.SetCell(
			x, y, symbol, termbox.ColorDefault, termbox.ColorDefault,
		)
	}

	termbox.SetCursor(x+1, y)
	termbox.Flush()
}
