package main

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"net/url"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strings"
	"sync"

	"eggactyl.cloud/runner/ui"
	"github.com/dustin/go-humanize"
	seccomp "github.com/elastic/go-seccomp-bpf"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/mem"
	"golang.org/x/sys/unix"
	"gopkg.in/yaml.v3"
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

	if _, err := os.Stat(fmt.Sprintf("%s/eggactyl_config.yml", os.Getenv("HOME"))); os.IsNotExist(err) {
		ConvertConfig()
	}

	repoUrl := os.Getenv("GIT_REPO")
	repoPat := os.Getenv("GIT_PAT")
	repoBranch := os.Getenv("GIT_BRANCH")

	var cfg YamlConfig
	if err := LoadConfig(fmt.Sprintf("%s/eggactyl_config.yml", os.Getenv("HOME")), &cfg); err != nil {
		log.Fatalln(err)
	}

	if strings.HasPrefix(cfg.Software.SoftwareType, "discord_") && len(repoUrl) > 0 {

		spinner := ui.Spinner("Grabbing git repo")

		parsedUrl, err := url.Parse(repoUrl)
		if err != nil {
			log.Fatalf("invalid url: %v\n", err)
		}

		if parsedUrl.Scheme == "ssh" {
			log.Fatalln("ssh urls are currently not supported")
		}

		if parsedUrl.Scheme == "" {
			parsedUrl.Scheme = "https"
		}

		gitRepoUrl := parsedUrl.String()

		cloneOptions := git.CloneOptions{
			URL:           gitRepoUrl,
			ReferenceName: plumbing.ReferenceName(fmt.Sprintf("refs/heads/%s", repoBranch)),
		}

		if len(repoPat) > 0 {
			cloneOptions.Auth = &http.BasicAuth{
				Username: "eggactyl",
				Password: repoPat,
			}
		}

		_, err = git.PlainClone("/opt/git_repo", false, &cloneOptions)
		if err != nil && err == git.ErrRepositoryAlreadyExists {

			repo, err := git.PlainOpen("/opt/git_repo")
			if err != nil {
				log.Fatalln(err)
			}

			rem, err := repo.Remote("origin")
			if err != nil {
				log.Fatalln(err)
			}

			if rem.Config().URLs[0] != gitRepoUrl {

				err = repo.DeleteRemote("origin")
				if err != nil {
					log.Fatalln(err)
				}

				rem, err = repo.CreateRemote(&config.RemoteConfig{
					Name: "origin",
					URLs: []string{gitRepoUrl},
				})

				if err != nil {
					log.Fatalln(err)
				}

			}

			fetchOptions := git.FetchOptions{}

			if len(repoPat) > 0 {
				fetchOptions.Auth = &http.BasicAuth{
					Username: "eggactyl",
					Password: repoPat,
				}
			}

			rem.Fetch(&fetchOptions)

			worktree, err := repo.Worktree()
			if err != nil {
				log.Fatalln(err)
			}

			err = worktree.Checkout(&git.CheckoutOptions{
				Branch: plumbing.ReferenceName(fmt.Sprintf("refs/remotes/origin/%s", repoBranch)),
				Force:  true,
			})
			if err != nil {
				log.Fatalln(err)
			}

		} else if err != nil {
			log.Fatalln(err)
		}

		fileList, err := getFileList("/opt/git_repo")
		if err != nil {
			log.Fatal(err)
		}

		destinationDir := os.Getenv("HOME")

		for _, file := range fileList {
			sourcePath := filepath.Join("/opt/git_repo", file)
			destinationPath := filepath.Join(destinationDir, file)

			if strings.HasSuffix(file, "/") {
				err := os.MkdirAll(destinationPath, os.ModePerm)
				if err != nil {
					log.Fatal(err)
				}
			} else {
				err := os.Symlink(sourcePath, destinationPath)
				if err != nil && !os.IsExist(err) {
					log.Fatal(err)
				}
			}
		}

		err = deleteNonexistentFiles(destinationDir, fileList)
		if err != nil {
			log.Fatal(err)
		}

		spinner.Success("Grabbed git repo")

	}

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

	//lint:ignore S1000 There is no issue with this.
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

func LoadConfig(file string, out *YamlConfig) error {

	cfgFile, err := os.ReadFile(file)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {

			if _, err := os.Create(file); err != nil {
				return err
			}

			cfgFile, err = os.ReadFile(file)
			if err != nil {
				return err
			}

		} else {
			return err
		}
	}

	if err := yaml.Unmarshal(cfgFile, out); err != nil {
		return err
	}

	return nil

}
