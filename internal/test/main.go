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
	defer speechRecognizer.Close()

	phrasesChan, ok := speechRecognizer.StartListening()
	if ok {
		fmt.Println("Listening ...")
	} else {
		fmt.Println("Failed to start listening, exiting program ...")
		return
	}

	fmt.Println("Say 'stop' to quit program")

	wg := &sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()

		for phrase := range phrasesChan {
			fmt.Println(phrase)
			if phrase == "stop" {
				fmt.Println("Stopped listening")
				speechRecognizer.StopListening()
			}
		}
	}()

	wg.Wait()
}
