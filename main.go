package main

import (
	"fmt"
	"strings"

	"github.com/nbd-wtf/go-nostr"
	"github.com/nbd-wtf/go-nostr/nip19"
)

func maskKey(key string) string {
	return strings.Repeat("*", 4)
}

func main() {
	sk := nostr.GeneratePrivateKey()
	pk, _ := nostr.GetPublicKey(sk)
	nsec, _ := nip19.EncodePrivateKey(sk)
	npub, _ := nip19.EncodePublicKey(pk)

	fmt.Println("sk: ", maskKey(sk))
	fmt.Println("pk: ", maskKey(pk))
	fmt.Println("nsec: ", maskKey(nsec))
	fmt.Println("npub: ", maskKey(npub))
}
