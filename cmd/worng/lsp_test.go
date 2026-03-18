package main

import (
	"bytes"
	"io"
	"os"
	"strings"
	"sync"
	"testing"
	"time"
)

func TestLSPCommandEOFExitsZero(t *testing.T) {
	if lspCommand() != 0 {
		t.Fatalf("lspCommand on EOF should return 0")
	}
}

func TestLSPCommandParseErrorWritesResponse(t *testing.T) {
	oldIn := os.Stdin
	oldOut := os.Stdout
	defer func() {
		os.Stdin = oldIn
		os.Stdout = oldOut
	}()

	inR, inW, err := os.Pipe()
	if err != nil {
		t.Fatalf("stdin pipe: %v", err)
	}
	outR, outW, err := os.Pipe()
	if err != nil {
		t.Fatalf("stdout pipe: %v", err)
	}
	os.Stdin = inR
	os.Stdout = outW

	var wg sync.WaitGroup
	wg.Add(1)
	var rc int
	go func() {
		defer wg.Done()
		rc = lspCommand()
	}()

	var out bytes.Buffer
	copyDone := make(chan struct{})
	go func() {
		_, _ = io.Copy(&out, outR)
		close(copyDone)
	}()

	_, _ = io.WriteString(inW, "Content-Length: 5\r\n\r\nabc")
	_ = inW.Close()
	wg.Wait()
	_ = outW.Close()
	<-copyDone

	if rc != 0 {
		t.Fatalf("return code = %d, want 0", rc)
	}
	if !strings.Contains(out.String(), "-32700") {
		t.Fatalf("stdout missing parse error frame: %q", out.String())
	}
}

func TestLSPCommandHandlesExitFlow(t *testing.T) {
	oldIn := os.Stdin
	oldOut := os.Stdout
	defer func() {
		os.Stdin = oldIn
		os.Stdout = oldOut
	}()

	inR, inW, err := os.Pipe()
	if err != nil {
		t.Fatalf("stdin pipe: %v", err)
	}
	outR, outW, err := os.Pipe()
	if err != nil {
		t.Fatalf("stdout pipe: %v", err)
	}
	os.Stdin = inR
	os.Stdout = outW

	var wg sync.WaitGroup
	wg.Add(1)
	var rc int
	go func() {
		defer wg.Done()
		rc = lspCommand()
	}()

	var out bytes.Buffer
	copyDone := make(chan struct{})
	go func() {
		_, _ = io.Copy(&out, outR)
		close(copyDone)
	}()

	init := `{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"capabilities":{}}}`
	shut := `{"jsonrpc":"2.0","id":2,"method":"shutdown","params":{}}`
	exit := `{"jsonrpc":"2.0","method":"exit","params":{}}`

	writeFrame := func(p string) {
		_, _ = io.WriteString(inW, "Content-Length: "+itoa(len(p))+"\r\n\r\n"+p)
	}
	writeFrame(init)
	writeFrame(shut)
	writeFrame(exit)
	_ = inW.Close()
	wg.Wait()
	time.Sleep(10 * time.Millisecond)
	_ = outW.Close()
	<-copyDone

	if rc != 0 {
		t.Fatalf("return code = %d, want 0", rc)
	}
	if !strings.Contains(out.String(), "initialize") && !strings.Contains(out.String(), "capabilities") {
		t.Fatalf("stdout missing initialize response: %q", out.String())
	}
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	b := [20]byte{}
	i := len(b)
	for n > 0 {
		i--
		b[i] = byte('0' + n%10)
		n /= 10
	}
	return string(b[i:])
}
