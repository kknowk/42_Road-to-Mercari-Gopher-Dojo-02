// package main
package main

import (
	"bufio"
	"fmt"
	"math/rand"
	"os"
	"time"

	"os/signal"
	"syscall"

	"github.com/fatih/color"
)

func CreateWord() ([]string, error) {
	file, err := os.Open("words.txt")
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)

	var words []string

	for scanner.Scan() {
		// add at the end of the slice
		words = append(words, scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		fmt.Println(err)
		return nil, err
	}

	return words, nil
}

func main() {

	words, err := CreateWord()
	if err != nil {
		os.Exit(1)
	}
	totalScore := 0

	// create timer
	timer := time.NewTimer((30 * time.Second))

	// シグナル通知用のチャネルを作成
	sigs := make(chan os.Signal, 1)

	// SIGINT（Ctrl+C）の通知を設定
	signal.Notify(sigs, syscall.SIGINT)

	// シグナル用のゴルーチン
	go func() {
		<-sigs
		color.Red("SIGINT")
		os.Exit(1)
	}()

	// async timer
	go func() {
		<-timer.C

		// timerに代入せず、直接受け取ることもできる
		// <-time.After(30 * time.Second)

		// print score
		fmt.Printf("\nTime's up! Score: %d\n", totalScore)
		os.Exit(0)
	}()

	scanner := bufio.NewScanner(os.Stdin)

	for {
		word := words[rand.Intn(len(words))]
		color.HiMagenta("Type the word: %s", word)
		fmt.Printf("-> ")

		for {
			if !scanner.Scan() {
				if err := scanner.Err(); err != nil {
					color.Red("Error reading input:", err)
				} else {
					color.Red("No more input available.")
				}
				return // EOFまたはエラーがあればループを抜けてプログラムを終了
			}

			input := scanner.Text()
			if input == word {
				totalScore++
				color.Green("Correct!\n")
				break // 正しい単語を入力したら内側のループを抜ける
			} else {
				color.Red("Typing error! one more please!\n")
				color.HiMagenta("Type the word: %s", word)
				fmt.Printf("-> ")
			}
		}
	}
}
