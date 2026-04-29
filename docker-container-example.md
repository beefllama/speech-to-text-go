# Docker Container Example

## 1. Start up vosk server docker container

Official guide: https://alphacephei.com/vosk/server

Run container (This will download docker image on the first run):

```
docker run -d -p 2700:2700 alphacep/kaldi-en:latest
```

## 2. Run sample program (from the root of the project)

```
go run internal/test/server/main.go
```

## Code sample

```golang
package main

import (
	"fmt"
	"log"
	"sync"

	stt "github.com/beefllama/speech-to-text-go"
)

const (
	voskServerHost = "localhost"
	voskServerPort = "2700"
)

func main() {
	fmt.Println("speech-to-text-go!")

	defer func() {
		fmt.Println("test deinit is done")
	}()

	speechRecognizer, err := stt.NewSpeechRecognizerWithVoskServer(voskServerHost, voskServerPort)
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

```