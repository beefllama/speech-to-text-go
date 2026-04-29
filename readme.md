# speech-to-text-go

A package for speech-to-text from microphone audio input.

This package captures audio input from the default microphone on the system and transcribes user's speech.

Speech recognition is enabled by vosk spech recognition toolkit and is done locally (offline).

There are 2 ways to use this package:
- vosk library + speech recognition model installed locally on the system
- docker container with vosk server

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

## Prerequisites

This package uses 2 dependencies:
- portaudio (audio capture from microphone)
- vosk (speech recognition)

Dependency installation is required.

### 1. Install portaudio

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

### 2. Install VOSK

[Setup guide for vosk server docker container](docker-container-example.md)

[Setup guide for vosk local model](local-model-example.md)