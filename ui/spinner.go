package ui

import (
	"fmt"
	"log"
	"time"

	"github.com/gosuri/uilive"
)

type SpinnerData struct {
	Title  string
	Writer *uilive.Writer
	Chan   chan bool
}

func Spinner(title string) SpinnerData {

	charSeq := []string{"▀ ", " ▀", " ▄", "▄ "}
	currChar := 0

	spinnerChan := make(chan bool)

	writer := uilive.New()

	writer.Start()

	go func() {

		for {
			select {
			case <-spinnerChan:
				return
			default:
				if currChar+1 > len(charSeq)-1 {
					currChar = 0

					//Heh. Nice. (I am a literal child.)
					spinner := fmt.Sprintf("\033[38;5;69m%s\033[0m", charSeq[currChar])
					if _, err := fmt.Fprintf(writer, "%s %s\n", spinner, title); err != nil {
						log.Fatalln(err)
					}

				} else {
					currChar++

					//Heh. Nice. (I am a literal child.)
					spinner := fmt.Sprintf("\033[38;5;69m%s\033[0m", charSeq[currChar])
					if _, err := fmt.Fprintf(writer, "%s %s\n", spinner, title); err != nil {
						log.Fatalln(err)
					}
				}

				time.Sleep(500 * time.Millisecond)
			}
		}

	}()

	return SpinnerData{
		Title:  title,
		Writer: writer,
		Chan:   spinnerChan,
	}

}

func (s SpinnerData) Success(text string) {

	if _, err := fmt.Fprintf(s.Writer, "\033[1m\033[38;5;36m%s\033[0m\n", text); err != nil {
		log.Fatalln(err)
	}

	s.Writer.Stop()

	s.Chan <- true

}

func (s SpinnerData) Error(text string) {

	if _, err := fmt.Fprintf(s.Writer, "\033[1m\033[38;5;210m%s\033[0m\n", text); err != nil {
		log.Fatalln(err)
	}

	s.Writer.Stop()

	s.Chan <- true

}

func (s SpinnerData) Warn(text string) {

	if _, err := fmt.Fprintf(s.Writer, "\033[1m\033[38;5;214m%s\033[0m\n", text); err != nil {
		log.Fatalln(err)
	}

	s.Writer.Stop()

	s.Chan <- true

}
