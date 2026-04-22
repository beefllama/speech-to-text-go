# speech-to-text-go

A package for speech-to-text from microphone audio input.

```
go get github.com/beefllama/speech-to-text-go
```

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
	fmt.Println("speech-to-text-go!")

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

You can check out other vosk models here: https://alphacephei.com/vosk/models

### 2. Install portaudio

You will need to download portaudio.

Check out this guide: https://files.portaudio.com/docs/v19-doxydocs/tutorial_start.html

**Ubuntu**:
```
sudo apt install portaudio19-dev
```

**Arch**:
```
sudo pacman -Syu portaudio
```

**MacOS**:
```
brew install portaudio
```

## Running a program

To run golang program that uses this package, you must provide environment variables for vosk.

VOSK_PATH must be the absolute path to vosk library downloaded [here](#1-install-vosk).

For linux (example):
```
VOSK_PATH=/home/beefllama/dev/speech-to-text-go/vosk/vosk-linux-x86_64-0.3.45 LD_LIBRARY_PATH=$VOSK_PATH CGO_CPPFLAGS="-I $VOSK_PATH" CGO_LDFLAGS="-L $VOSK_PATH -ldl" go run ./internal/test/main.go
```

For other platforms consult vosk documentation: https://github.com/alphacep/vosk-api/tree/master/go/example