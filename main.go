package main

import (
	"flag"
	"io"
	"log"
	"net"
	"strings"
	"sync"
	"time"
)

var (
	verbose = flag.Bool("v", false, "verbose (log reads/writes)")
	src     = flag.String("src", "", "source socket / listen address")
	dst     = flag.String("dst", "", "destination socket / address")
	delay   = flag.Duration("delay", 0, "delay (both ways)")
)

func main() {
	flag.Parse()

	network := "tcp"
	if strings.HasPrefix(*src, "/") {
		network = "unix"
	}
	ln, err := net.Listen(network, *src)
	if err != nil {
		log.Fatal(err)
	}
	for {
		inc, err := ln.Accept()
		if err != nil {
			log.Print(err)
			break
		}
		go func(inc net.Conn) {
			if err := handleConnection(inc); err != nil {
				log.Print(err)
			}
		}(inc)
	}
}

func handleConnection(inc net.Conn) error {
	network := "tcp"
	if strings.HasPrefix(*dst, "/") {
		network = "unix"
	}
	outg, err := net.Dial(network, *dst)
	if err != nil {
		return err
	}
	var wg sync.WaitGroup
	wg.Add(2)
	go slowCopy(">>>", inc, outg, *delay)
	go slowCopy("<<<", outg, inc, *delay)
	wg.Wait()
	return nil
}

type chunk struct {
	data []byte
	ts   time.Time
}

func readChan(pref string, r io.Reader, q chan<- chunk) {
	defer close(q)
	buf := make([]byte, 32*1024)
	for {
		nr, err := r.Read(buf)
		if *verbose {
			log.Printf("%s read %d bytes: %v", pref, nr, err)
		}
		if nr > 0 {
			var ch chunk
			ch.ts = time.Now()
			ch.data = append(ch.data, buf[:nr]...)
			q <- ch
		}
		if err != nil {
			break
		}
	}
}

// slowCopy copies from r to w, adding a delay of d to every
// read, to attempt to simulate a high latency connection.
func slowCopy(pref string, r io.Reader, w io.WriteCloser, d time.Duration) {
	q := make(chan chunk, 1024)
	go readChan(pref, r, q)
	for ch := range q {
		time.Sleep(time.Until(ch.ts.Add(d)))
		nw, err := w.Write(ch.data)
		if *verbose {
			log.Printf("%s wrote %d bytes: %v", pref, nw, err)
		}
		if err != nil {
			break
		}
		if nw < len(ch.data) {
			log.Printf("%s short write", pref)
			break
		}
	}
	w.Close()
	for range q {
	}
}
