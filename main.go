package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"sync"

	"golang.org/x/sys/unix"
)

var SupportLink *string
var Script *string

func init() {
	SupportLink = flag.String("support-link", "https://example.com", "https://example.com")
	Script = flag.String("script", "", "/path/to/executable")

	flag.Parse()
}

func main() {

	userInput := make(chan string)
	notifyChan := make(chan string)

	go func() {

		reader := bufio.NewScanner(os.Stdin)

		for reader.Scan() {

			//Because of Pterodactyl works, we can always expect the string to not be empty.
			text := reader.Text()

			//Because of an issue with the wings, it cannot process SIGINT, and will not send any signals to the applicatiion.
			//GitHub Issue: https://github.com/pterodactyl/panel/issues/4783
			if strings.HasPrefix(text, "EGG_SIGNAL") && strings.HasSuffix(text, "_SIGINT") {

				currentPid := os.Getpid()

				if err := unix.Kill(currentPid, unix.SIGINT); err != nil {
					log.Fatalln(err)
				}

				break

			}

			userInput <- text
		}

		close(userInput)
	}()

	go startMainProcess(userInput, notifyChan)

	for {

		select {
		case <-notifyChan:
			os.Exit(0)
		}

	}

}

func startMainProcess(userInput chan string, notifyChan chan string) {

	fmt.Println(*Script)

	cmd := exec.Command("bash", "-c", *Script)

	cmd.SysProcAttr = &unix.SysProcAttr{Setsid: true}

	//Channels to notify if parent has call to shut down
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, unix.SIGINT, unix.SIGTERM)

	stdin, err := cmd.StdinPipe()
	if err != nil {
		fmt.Println("Uh oh! I seem to have run into an error!")
		fmt.Printf("Please contact support at %s\n", *SupportLink)
		log.Fatalln(err)
	}

	defer stdin.Close()

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		fmt.Println("Uh oh! I seem to have run into an error!")
		fmt.Printf("Please contact support at %s\n", *SupportLink)
		log.Fatalln(err)
	}

	defer stdout.Close()

	stderr, err := cmd.StderrPipe()
	if err != nil {
		fmt.Println("Uh oh! I seem to have run into an error!")
		fmt.Printf("Please contact support at %s\n", *SupportLink)
		log.Fatalln(err)
	}

	defer stderr.Close()

	if err := cmd.Start(); err != nil {
		fmt.Println("Uh oh! I seem to have run into an error!")
		fmt.Printf("Please contact support at %s\n", *SupportLink)
		log.Fatalln(err)
	}

	go func() {
		for input := range userInput {
			if _, err := fmt.Fprintln(stdin, input); err != nil {
				fmt.Println("Uh oh! I seem to have run into an error!")
				fmt.Printf("Please contact support at %s\n", *SupportLink)
				log.Fatalln(err)
			}
		}
	}()

	go func() {
		<-sigChan
		if err := unix.Kill(-cmd.Process.Pid, unix.SIGINT); err != nil {
			fmt.Println("Uh oh! I seem to have run into an error!")
			fmt.Printf("Please contact support at %s\n", *SupportLink)
			log.Fatalln(err)
		}
	}()

	var wg sync.WaitGroup
	wg.Add(2)

	//Handle stdout
	go func() {
		defer wg.Done()
		io.Copy(os.Stdout, stdout)
	}()

	go func() {
		defer wg.Done()
		io.Copy(os.Stderr, stderr)
	}()

	wg.Wait()

	_ = cmd.Wait()

	notifyChan <- "closed"

}
