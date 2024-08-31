package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/nbd-wtf/go-nostr"
)

const (
	BaseURI = "https://public.bitbank.cc/"
)

var (
	urls  = [2]string{"wss://relay-jp.nostr.wirednet.jp", "wss://relay.nostr.wirednet.jp"}
	nsec  = os.Getenv("NOSTR_SK")
	pairs = [4]string{"btc_jpy", "eth_jpy", "bcc_jpy", "bat_jpy"}
)

type (
	Ticker struct {
		Success int   `json:"success"`
		Data    *Data `json:"data"`
	}
	Data struct {
		Sell      string `json:"sell"`
		Buy       string `json:"buy"`
		High      string `json:"high"`
		Low       string `json:"low"`
		Last      string `json:"last"`
		Vol       string `json:"vol"`
		TimeStamp int64  `json:"timestamp"`
	}

	TickerResponse struct {
		Pair      string
		Last      string
		TimeStamp string
	}
)

func init() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)
}

func main() {
	pub, err := nostr.GetPublicKey(nsec)
	if err != nil {
		slog.Warn("message", "err", err)
		return
	}

	results := make(chan *TickerResponse, len(pairs))
	errors := make(chan error, len(pairs))
	for _, pair := range pairs {
		go fetchExchangeRate(pair, results, errors)
	}

	posts := make([]string, 0, len(pairs))
	for i := 0; i < len(pairs); i++ {
		select {
		case res := <-results:
			posts = append(posts, fmt.Sprintf("Pair: %v, Last Price: %v, Timestamp: %v\n", res.Pair, res.Last, res.TimeStamp))
		case err := <-errors:
			fmt.Println("Error:", err)
		}
	}

	result := fmt.Sprintln("仮想通貨の価格\n" + strings.Join(posts, " "))

	publishRelay(pub, urls, result)
}

func fetchExchangeRate(pair string, results chan<- *TickerResponse, errors chan<- error) {
	resp, err := http.Get(BaseURI + pair + "/ticker")
	if err != nil {
		errors <- err
		return
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		errors <- err
		return
	}

	var ticker Ticker
	if err := json.Unmarshal(data, &ticker); err != nil {
		errors <- err
		return
	}

	timeStamp := time.Unix(ticker.Data.TimeStamp/1000, 0).Format("2006-01-02 15:04:05")

	results <- &TickerResponse{
		Pair:      pair,
		Last:      ticker.Data.Last,
		TimeStamp: timeStamp,
	}

}

func publishRelay(pub string, urls [2]string, post string) {
	ev := nostr.Event{
		PubKey:    pub,
		CreatedAt: nostr.Now(),
		Kind:      nostr.KindTextNote,
		Tags:      nil,
		Content:   post,
	}

	if err := ev.Sign(nsec); err != nil {
		slog.Warn("message", "err", err)
		return
	}

	ctx := context.Background()

	for _, url := range urls {
		relay, err := nostr.RelayConnect(ctx, url)
		if err != nil {
			slog.Warn("message", "err", err)
			continue
		}
		if err := relay.Publish(ctx, ev); err != nil {
			slog.Warn("message", "err", err)
			continue
		}
	}
	slog.Info("message", "INFO", "Success to published!")
}
