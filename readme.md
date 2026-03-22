# speech-to-text-go

A package that transcribes speech from microphone audio input.

Table of Contents:
- [Quick Start Example](#quick-start-example)
- [Prerequisites](#prerequisites)
- [Running a program](#running-a-program)

## Quick Start Example

```golang
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
```

## Prerequisites

This package uses 2 dependencies:
- vosk
- portaudio

Dependency installation is required.

### 1. Install VOSK

Official set up guide for Go is here: https://github.com/alphacep/vosk-api/tree/master/go/example

1. Download VOSK

```
wget https://github.com/alphacep/vosk-api/releases/download/v0.3.45/vosk-linux-x86_64-0.3.45.zip

unzip vosk-linux-x86_64-0.3.45.zip
```

2. Download VOSK model

```
wget https://alphacephei.com/vosk/models/vosk-model-small-en-us-0.15.zip

unzip vosk-model-small-en-us-0.15.zip
```

### 2. Install portaudio

You will need to download portaudio.

Check out this guide: https://files.portaudio.com/docs/v19-doxydocs/tutorial_start.html

On Ubuntu, you can install portaudio with the following command:
```
sudo apt install portaudio19-dev
```

## Running a program

To run golang program that uses this package, you must provide environment variables for vosk.

VOSK_PATH must be the absolute path to vosk library downloaded [here](#1-install-vosk).

For linux:
```
VOSK_PATH=/home/beefllama/dev/speech-to-text-go/vosk/vosk-linux-x86_64-0.3.45 LD_LIBRARY_PATH=$VOSK_PATH CGO_CPPFLAGS="-I $VOSK_PATH" CGO_LDFLAGS="-L $VOSK_PATH -ldl" go run ./internal/test/main.go
```

For other platforms consult vosk documentation: https://github.com/alphacep/vosk-api/tree/master/go/example