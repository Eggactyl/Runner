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

	"github.com/dustin/go-humanize"
	seccomp "github.com/elastic/go-seccomp-bpf"
	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/mem"
	"golang.org/x/sys/unix"
)

var SupportLink *string
var Script *string
var ScriptArgs *string
var AntiDiskFill *bool
var ShowHWInfo *bool

func init() {
	SupportLink = flag.String("support-link", "", "https://example.com")
	Script = flag.String("script", "", "/path/to/executable")
	ScriptArgs = flag.String("script-args", "", "--enable-something")
	AntiDiskFill = flag.Bool("anti-disk-fill", true, "--anti-disk-fill")
	ShowHWInfo = flag.Bool("show-hw-info", false, "--show-hw-info")

	flag.Parse()
}

func main() {

	fmt.Println("\033[H\033[2J")

	if *ShowHWInfo {
		memInfo, err := mem.VirtualMemory()
		if err != nil {
			return
		}

		swapInfo, err := mem.SwapMemory()
		if err != nil {
			return
		}

		cpuInfo, err := cpu.Info()
		if err != nil {
			return
		}

		totalRam := humanize.IBytes(memInfo.Total)
		usedRam := humanize.IBytes(memInfo.Used)

		totalSwap := humanize.IBytes(swapInfo.Total)
		usedSwap := humanize.IBytes(swapInfo.Used)

		fmt.Println("\033[4m\033[1m\033[38;5;33mHardware Information:\033[0m")

		fmt.Printf("  \033[38;5;33mCPU Model: %s\033[0m\n", cpuInfo[0].ModelName)

		var memUsagePercent string
		var swapUsagePercent string

		if memInfo.UsedPercent >= 90.0 {
			memUsagePercent = fmt.Sprintf("\033[3m\033[1m\033[38;5;210m(%.2f%%)\033[0m", memInfo.UsedPercent)
		} else if memInfo.UsedPercent >= 70.0 {
			memUsagePercent = fmt.Sprintf("\033[3m\033[1m\033[38;5;214m(%.2f%%)\033[0m", memInfo.UsedPercent)
		} else {
			memUsagePercent = fmt.Sprintf("\033[3m\033[1m\033[38;5;34m(%.2f%%)\033[0m", memInfo.UsedPercent)
		}

		if swapInfo.UsedPercent >= 90.0 {
			swapUsagePercent = fmt.Sprintf("\033[3m\033[1m\033[38;5;210m(%.2f%%)\033[0m", swapInfo.UsedPercent)
		} else if swapInfo.UsedPercent >= 70.0 {
			swapUsagePercent = fmt.Sprintf("\033[3m\033[1m\033[38;5;214m(%.2f%%)\033[0m", swapInfo.UsedPercent)
		} else {
			swapUsagePercent = fmt.Sprintf("\033[3m\033[1m\033[38;5;34m(%.2f%%)\033[0m", swapInfo.UsedPercent)
		}

		fmt.Printf("  \033[38;5;33mRAM Usage: %s / %s \033[0m%s \n", usedRam, totalRam, memUsagePercent)
		fmt.Printf("  \033[38;5;33mSwap Usage: %s / %s \033[0m%s \n", usedSwap, totalSwap, swapUsagePercent)
	}

	if *AntiDiskFill {

		filter := seccomp.Filter{
			NoNewPrivs: true,
			Flag:       seccomp.FilterFlagTSync,
			Policy: seccomp.Policy{
				DefaultAction: seccomp.ActionAllow,
				Syscalls: []seccomp.SyscallGroup{
					{
						Action: seccomp.ActionKillProcess,
						Names: []string{
							"fallocate",
						},
					},
				},
			},
		}

		if err := seccomp.LoadFilter(filter); err != nil {
			log.Fatalln(err)
		}

	}

	userInput := make(chan string)
	notifyChan := make(chan string)

	go func() {

		reader := bufio.NewScanner(os.Stdin)

		for reader.Scan() {

			//Because of Pterodactyl works, we can always expect the string to not be empty.
			text := reader.Text()

			//Because of an issue with the wings, it cannot process SIGINT, and will not send any signals to the applicatiion.
			//GitHub Issue: https://github.com/pterodactyl/panel/issues/4783
			if text == "EGG_SIGNAL_SIGINT" || text == "^C" {

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

	cmdWithArgs := strings.Join(append([]string{*Script}, *ScriptArgs), " ")

	cmd := exec.Command("/bin/bash", "-c", cmdWithArgs)
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
		if len(*SupportLink) > 0 {
			fmt.Printf("Please contact support at %s\n", *SupportLink)
		}
		log.Fatalln(err)
	}

	defer stdout.Close()

	stderr, err := cmd.StderrPipe()
	if err != nil {
		fmt.Println("Uh oh! I seem to have run into an error!")
		if len(*SupportLink) > 0 {
			fmt.Printf("Please contact support at %s\n", *SupportLink)
		}
		log.Fatalln(err)
	}

	defer stderr.Close()

	if err := cmd.Start(); err != nil {
		fmt.Println("Uh oh! I seem to have run into an error!")
		if len(*SupportLink) > 0 {
			fmt.Printf("Please contact support at %s\n", *SupportLink)
		}
		log.Fatalln(err)
	}

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

	go func() {
		for input := range userInput {
			if _, err := fmt.Fprintln(stdin, input); err != nil {
				fmt.Println("Uh oh! I seem to have run into an error!")
				if len(*SupportLink) > 0 {
					fmt.Printf("Please contact support at %s\n", *SupportLink)
				}
				log.Fatalln(err)
			}
		}
	}()

	go func() {
		<-sigChan
		if err := unix.Kill(-cmd.Process.Pid, unix.SIGINT); err != nil {
			fmt.Println("Uh oh! I seem to have run into an error!")
			if len(*SupportLink) > 0 {
				fmt.Printf("Please contact support at %s\n", *SupportLink)
			}
			log.Fatalln(err)
		}
	}()

	wg.Wait()

	if err := cmd.Wait(); err != nil {

		if exitErr, ok := err.(*exec.ExitError); ok {

			//Ignore interrupts
			if exitErr.Error() != "signal: interrupt" {
				fmt.Println("Uh oh! I seem to have run into an error!")
				if len(*SupportLink) > 0 {
					fmt.Printf("Please contact support at %s\n", *SupportLink)
				}
				log.Fatalln(err)
			}

		}

	}

	notifyChan <- "closed"

}
