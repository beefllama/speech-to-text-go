package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"

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

	phrasesChan := speechRecognizer.GetTextOuputStream()

	speechRecognizer.StartListening()

	log.Println("listening...")

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

loop:
	for {
		select {
		case phrase := <-phrasesChan:
			fmt.Println(phrase)
		case <-ctx.Done():
			break loop
		default:
		}
	}
}
