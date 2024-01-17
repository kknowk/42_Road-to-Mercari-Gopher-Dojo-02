// package main
package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"sync"

	"net/url"
	"os/signal"
	"path"
	"syscall"
	"bufio"

	"github.com/fatih/color"
	"golang.org/x/sync/errgroup"
)

const (
	numberOfParts = 100 // ダウンロードを分割する部分の数
)

func main() {
	if len(os.Args) < 2 {
		color.Red("Argument is missing\n")
		color.HiMagenta("Usage: ./download <url>\n")
		os.Exit(1)
	}
	rawURL := os.Args[1]

	// Get Content-Length
	resp, err := http.Head(rawURL)
	if err != nil {
		color.Red("Error: %s\n", err)
		os.Exit(1)
	}
	size, err := strconv.Atoi(resp.Header.Get("Content-Length"))
	if err != nil {
		color.Red("Error: %s\n", err)
		os.Exit(1)
	}

	// ファイルサイズを分割数で割って、各部分のサイズを計算
	partSize := size / numberOfParts

	// errgroupを使って並行処理
	var eg errgroup.Group
	// ロックを使ってスライスを同期
	var mu sync.Mutex

	// ダウンロードしたファイルの一部を格納するスライス
	parts := make([][]byte, numberOfParts)

	// シグナル通知用のチャネルを作成
	sigs := make(chan os.Signal, 1)

	// SIGINT（Ctrl+C）の通知を設定
	signal.Notify(sigs, syscall.SIGINT)

	go func() {
		sig := <-sigs
		if sig == syscall.SIGINT {
			color.Red("SIGINT received")
			color.Red("Download interrupted")
			os.Exit(1)
		}
	}()

	// http.Client のインスタンスを作成
	client := &http.Client{}

	for i := 0; i < numberOfParts; i++ {
		i := i // ゴルーチン内でループ変数をキャプチャする
		eg.Go(func() error {
			// ダウンロードする部分の範囲を計算
			start := i * partSize
			end := start + partSize - 1

			// 最後の部分の場合は、最後までダウンロードする
			if i == numberOfParts-1 {
				end = size
			}

			// ダウンロード
			req, err := http.NewRequest("GET", rawURL, nil)
			if err != nil {
				return err
			}
			// Rangeヘッダーを設定
			req.Header.Set("Range", fmt.Sprintf("bytes=%d-%d", start, end))

			// リクエストを送信
			// http.DefaultClient の代わりに client を使用
			resp, err := client.Do(req)
			if err != nil {
				return err
			}
			defer resp.Body.Close()

			// レスポンスボディを読み込む
			part, err := io.ReadAll(resp.Body)
			if err != nil {
				return err
			}

			// ミューテックスを使ってスライスを同期
			mu.Lock()
			parts[i] = part
			mu.Unlock()
			return nil
		})
	}

	// エラーをチェック
	if err := eg.Wait(); err != nil {
		color.Red("Error: %s\n", err)
		os.Exit(1)
	}

	parsedUrl, err := url.Parse(rawURL)
	if err != nil {
		color.Red("Error parsing URL: %s\n", err)
		os.Exit(1)
	}
	filename := path.Base(parsedUrl.Path)

	// ファイルをマージ
	file, err := os.Create(filename)
	if err != nil {
		color.Red("Error: %s\n", err)
		os.Exit(1)
	}
	defer file.Close()

	// バッファライターを作成
	writer := bufio.NewWriter(file)

	for _, part := range parts {
		if _, err := writer.Write(part); err != nil {
			color.Red("Error: %s\n", err)
			os.Exit(1)
		}
	}

	if err := writer.Flush(); err != nil {
		color.Red("Error: %s\n", err)
		os.Exit(1)
	}


	color.Green("Downloaded %s\n", filename)
}

// test command
// ./download "https://img.cpcdn.com/users/5572512/180x180c/3fe22dcf64832a045a6a12b54f04e8b3?u=5572512&p=1359025345"
