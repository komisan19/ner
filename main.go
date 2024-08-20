package main

import (
	"context"
	"log/slog"
	"os"
	"strings"

	"github.com/nbd-wtf/go-nostr"
)

func maskKey(key string) string {
	return strings.Repeat("*", 4)
}

var (
	urls = [2]string{"wss://relay-jp.nostr.wirednet.jp", "wss://relay.nostr.wirednet.jp"}
	sk   string
)

func init() {
	sk = os.Getenv("NOSTR_SK")
}

func main() {
	publishRelay(urls)
}

func publishRelay(urls [2]string) {
	pub, err := nostr.GetPublicKey(sk)
	if err != nil {
		slog.Warn("error: ", err)
		return
	}

	ev := nostr.Event{
		PubKey:    pub,
		CreatedAt: nostr.Now(),
		Kind:      nostr.KindTextNote,
		Tags:      nil,
		Content:   "Hello World for coudflare",
	}

	ev.Sign(sk)

	ctx := context.TODO()

	for _, url := range urls {
		relay, err := nostr.RelayConnect(ctx, url)
		if err != nil {
			slog.Warn("error: ", err)
			continue
		}
		if err := relay.Publish(ctx, ev); err != nil {
			slog.Warn("error: ", err)
			continue
		}

		slog.Warn("sucess: ", url)
	}
}
