package main

import (
	"fmt"
	"log"
	"sync"

	stt "github.com/beefllama/speech-to-text-go"
)

const (
	voskModelPath = "vosk/vosk-model-small-en-us-0.15"
)

func main() {
	fmt.Println("speech-to-text-go !")

	defer func() {
		fmt.Println("test deinit is done")
	}()

	speechRecognizer, err := stt.NewSpeechRecognizer(voskModelPath)
	if err != nil {
		log.Printf("failed to init speech recognizer, err: %s\n", err.Error())
		return
	}

	phrasesChan := speechRecognizer.GetTextOuputStream()

	speechRecognizer.StartListening()

	fmt.Println("Listening...")
	fmt.Println("Press Enter to quit program")

	wg := &sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()

		for phrase := range phrasesChan {
			fmt.Println(phrase)
		}
	}()

	fmt.Scanln()
	speechRecognizer.Close()
	wg.Wait()
}
