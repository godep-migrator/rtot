package main

import (
	"crypto/md5"
	"crypto/rand"
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/modcloth-labs/rtot"
)

var (
	addr   = os.Getenv("RTOT_ADDR")
	secret = os.Getenv("RTOT_SECRET")
)

func makeSecret() string {
	buf := make([]byte, 16)
	_, err := rand.Read(buf)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to generate a secret! %v", err)
		os.Exit(1)
	}
	hash := md5.New()
	io.WriteString(hash, string(buf))
	return fmt.Sprintf("%x", hash.Sum(nil))
}

func main() {
	if addr == "" {
		addr = ":8457"
	}

	flag.StringVar(&addr,
		"a", addr, "HTTP Server address [RTOT_ADDR]")
	flag.StringVar(&secret,
		"s", secret, "Secret string for secret stuff [RTOT_SECRET]")
	versionFlag := flag.Bool("v", false, "Show version and exit")

	flag.Parse()

	if *versionFlag {
		fmt.Printf("rtot-server %v\n", rtot.VersionString)
		os.Exit(0)
	}

	if secret == "" {
		secret = makeSecret()
		fmt.Printf("[rtot] No secret given, so generated %q\n", secret)
	}

	_, err := rtot.NewJobGroup("main", "memory")
	if err != nil {
		fmt.Fprintf(os.Stderr, "[rtot] Failed to init job store: %v\n", err)
		os.Exit(1)
	}

	rtot.ServerMain(addr, secret)
}
