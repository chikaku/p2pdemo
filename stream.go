package main

import (
	"bufio"
	"fmt"
	"log/slog"
	"os"
	"strings"
	"sync"

	"github.com/fatih/color"
	"github.com/libp2p/go-libp2p/core/network"
)

func handleStream(s network.Stream) {
	slog.Info("😼 stream connected")
	rw := bufio.NewReadWriter(bufio.NewReader(s), bufio.NewWriter(s))
	wg := new(sync.WaitGroup)
	wg.Add(2)
	go readData(rw, wg)
	go writeData(rw, wg)
	wg.Wait()
}

func readData(rw *bufio.ReadWriter, wg *sync.WaitGroup) {
	defer wg.Done()
	for {
		str, err := rw.ReadString('\n')
		if err != nil {
			fmt.Print(color.RedString("Error reading from buffer: " + err.Error()))
			return
		}
		if str = strings.TrimSpace(str); str != "" {
			fmt.Printf("recv: %s\n", color.GreenString(str))
			fmt.Print("> ")
		}
	}
}

func writeData(rw *bufio.ReadWriter, wg *sync.WaitGroup) {
	defer wg.Done()
	stdReader := bufio.NewReader(os.Stdin)
	for {
		fmt.Printf("> ")
		sendData, err := stdReader.ReadString('\n')
		if err != nil {
			fmt.Print(color.RedString("😿 Error reading from stdin: " + err.Error()))
			return
		}
		_, err = rw.WriteString(sendData)
		if err != nil {
			fmt.Print(color.RedString("😿 Error writing to buffer: " + err.Error()))
			return
		}
		err = rw.Flush()
		if err != nil {
			fmt.Print(color.RedString("😿 Error flushing buffer: " + err.Error()))
			return
		}
	}
}
