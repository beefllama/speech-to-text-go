package stt

import (
	"encoding/binary"
	"sync"
	"sync/atomic"

	"github.com/gordonklaus/portaudio"
	"github.com/pkg/errors"
)

const chanBufferSize = 1000

// SpeechRecognizer reads audio input from microphone and translates speech into text.
type SpeechRecognizer struct {
	voskClient         voskClient
	microphoneStream   *portaudio.Stream
	audioSamplesBuffer []int16
	stopCaptureSignal  chan struct{}
	wg                 *sync.WaitGroup
	listeningEnabled   atomic.Bool
}

// StartListening enables microphone audio capture and transcription.
// Returns a channel of phrases transcribed from microphone input and true if listening has not been enabled previously.
// If listeting has been started already, this function will return nil and false.
func (r *SpeechRecognizer) StartListening() (<-chan string, bool) {
	if r.IsListening() {
		return nil, false
	}

	audioInputChan := make(chan []byte, chanBufferSize)
	textOutputChan := make(chan string, chanBufferSize)

	r.wg.Add(2)

	go r.convertSpeechToText(audioInputChan, textOutputChan)
	go r.captureAudioFromMicrophone(audioInputChan)

	r.listeningEnabled.Store(true)

	return textOutputChan, true
}

// StopListening disables microphone audio capture and transcription.
func (r *SpeechRecognizer) StopListening() {
	if !r.IsListening() {
		return
	}

	r.stopCaptureSignal <- struct{}{}
	r.wg.Wait()

	r.listeningEnabled.Store(false)
}

// IsListening returns true if microphone audio capture and transcription is running.
func (r *SpeechRecognizer) IsListening() bool {
	return r.listeningEnabled.Load()
}

// Close must be called when done using SpeechRecognizer.
func (r *SpeechRecognizer) Close() {
	r.StopListening()

	close(r.stopCaptureSignal)

	deinitAudioCapture(r.microphoneStream)
	r.voskClient.Close()
}

func (r *SpeechRecognizer) convertSpeechToText(audioInputChan <-chan []byte, textOutputChan chan<- string) {
	defer r.wg.Done()

	for audioData := range audioInputChan {
		ready, err := r.voskClient.ProcessAudioChunk(audioData)
		if err != nil {
			panic(errors.Wrap(err, "failed to process audio chunk"))
		}

		if !ready {
			continue
		}

		text, err := r.voskClient.GetResults()
		if err != nil {
			panic(errors.Wrap(err, "failed to get results from vosk recognizer"))
		}

		if len(text) == 0 {
			continue
		}

		textOutputChan <- text
	}

	text, err := r.voskClient.GetFinalResults()
	if err != nil {
		panic(errors.Wrap(err, "failed to get final results from vosk recognizer"))
	}

	textOutputChan <- text
	close(textOutputChan)
}

func (r *SpeechRecognizer) captureAudioFromMicrophone(audioInputChan chan<- []byte) {
	defer r.wg.Done()

	if startErr := r.microphoneStream.Start(); startErr != nil {
		panic(errors.Wrap(startErr, "failed to start microphone stream"))
	}

loop:
	for {
		if readErr := r.microphoneStream.Read(); readErr != nil {
			panic(errors.Wrap(readErr, "failed to read from microphone stream"))
		}

		select {
		case audioInputChan <- int16ToBytes(r.audioSamplesBuffer):
		case <-r.stopCaptureSignal:
			break loop
		}
	}

	close(audioInputChan)

	if stopErr := r.microphoneStream.Stop(); stopErr != nil {
		panic(errors.Wrap(stopErr, "failed to stop microphone stream"))
	}
}

func int16ToBytes(samples []int16) []byte {
	// TODO make this faster
	buf := make([]byte, len(samples)*2) // 2 bytes per int16
	for i, v := range samples {
		binary.LittleEndian.PutUint16(buf[i*2:], uint16(v))
	}
	return buf
}

// NewSpeechRecognizer initializes an instance of SpeechRecognizer that uses local vosk model.
// At some point before program exit, SpeechRecognizer.Close must be called to deinitialize.
func NewSpeechRecognizerWithLocalVoskModel(voskModelPath string) (*SpeechRecognizer, error) {
	voskClient, err := newLocalModelVoskClient(voskModelPath)
	if err != nil {
		return nil, errors.Wrap(err, "failed to init vosk client")
	}

	return newSpeechRecognizer(voskClient)
}

// NewSpeechRecognizer initializes an instance of SpeechRecognizer that uses vosk server.
// At some point before program exit, SpeechRecognizer.Close must be called to deinitialize.
func NewSpeechRecognizerWithVoskServer(host, port string) (*SpeechRecognizer, error) {
	voskClient, err := newVoskServerClient(host, port)
	if err != nil {
		return nil, errors.Wrap(err, "failed to init vosk client")
	}

	return newSpeechRecognizer(voskClient)
}

func newSpeechRecognizer(voskClient voskClient) (*SpeechRecognizer, error) {
	microphoneStream, audioSamplesBuffer, err := initAudioCapture()
	if err != nil {
		return nil, errors.Wrap(err, "failed to init audio capture")
	}

	speechRecognizer := &SpeechRecognizer{
		voskClient:         voskClient,
		microphoneStream:   microphoneStream,
		audioSamplesBuffer: audioSamplesBuffer,
		stopCaptureSignal:  make(chan struct{}),
		wg:                 &sync.WaitGroup{},
		listeningEnabled:   atomic.Bool{},
	}

	return speechRecognizer, nil
}
