package main

import (
	"bufio"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/big"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/docopt/docopt-go"
)

const (
	usage = `Short 1.0, short term memory tester.

Usage:
	./short [options]

Options:
	--dry         run tests and show results.
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

	dry := args["--dry"].(bool)

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

	clearScreen()

	results := []Result{}
	for i := 0; i < testsCount; i++ {
		result := runTest(minNumber, maxNumber, numbersCount)
		if dry {
			fmt.Println("Score: ", result.Score)
			fmt.Println("Duration: ", result.Duration)
			wait()
			clearScreen()
		}

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

	if dry {
		fmt.Println("Total score: ", sumScore)
		fmt.Println("Average duration: ", avgDuration)
	}

	saveResults(file, results, sumScore, avgDuration)
}

func saveResults(file string, results []Result, totalScore int, avgDuration float64) {
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

	strs := []string{}
	for _, number := range validNumbers {
		strs = append(strs, strconv.Itoa(number))
	}

	timeStart := time.Now()

	fmt.Println(strings.Join(strs, " "))

	wait() //wait for input 'Enter'
	clearScreen()

	userNumbers := getNumbersFromStdin()

	timeFinish := time.Now()

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

func getNumbersFromStdin() []int {
	numbers := []int{}

	scanner := bufio.NewScanner(os.Stdin)
	if scanner.Scan() {
		pieces := strings.Split(scanner.Text(), " ")
		for _, piece := range pieces {
			number, _ := strconv.Atoi(piece)
			numbers = append(numbers, number)
		}
	}

	return numbers

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
	cmd := exec.Command("clear")
	cmd.Stdout = os.Stdout
	cmd.Run()
}

// just wait for any user input (like 'Press Enter to continue')
func wait() {
	var ready string
	fmt.Scanf("%s", &ready)
}
