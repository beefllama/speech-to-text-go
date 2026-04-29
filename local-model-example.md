# Local Model Example

## 1. Install vosk

Official set up guide for Go is here: https://github.com/alphacep/vosk-api/tree/master/go/example

### 1.1. Download VOSK

```
wget https://github.com/alphacep/vosk-api/releases/download/v0.3.45/vosk-linux-x86_64-0.3.45.zip

unzip vosk-linux-x86_64-0.3.45.zip
```

### 1.2. Download VOSK model

```
wget https://alphacephei.com/vosk/models/vosk-model-small-en-us-0.15.zip

unzip vosk-model-small-en-us-0.15.zip
```

You can check out other vosk models here: https://alphacephei.com/vosk/models

## 2. Run sample program (from the root of the project)

To run golang program that uses this package, you must provide environment variables for vosk.

VOSK_PATH must be the absolute path to vosk library downloaded [here](#11-download-vosk).

For linux (example):
```
VOSK_PATH=/home/beefllama/dev/speech-to-text-go/vosk/vosk-linux-x86_64-0.3.45 LD_LIBRARY_PATH=$VOSK_PATH CGO_CPPFLAGS="-I $VOSK_PATH" CGO_LDFLAGS="-L $VOSK_PATH -ldl" go run ./internal/test/local-model/main.go
```

For other platforms consult vosk documentation: https://github.com/alphacep/vosk-api/tree/master/go/example

## Code Sample

```golang
package main

import (
	"fmt"
	"log"
	"sync"

	stt_local_model "github.com/beefllama/speech-to-text-go/local-model"
)

const (
	voskModelPath = "vosk/vosk-model-small-en-us-0.15"
)

func main() {
	fmt.Println("speech-to-text-go!")

	speechRecognizer, err := stt_local_model.NewSpeechRecognizerWithLocalVoskModel(voskModelPath)
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