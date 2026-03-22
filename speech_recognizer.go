package stt

import (
	"encoding/binary"
	"encoding/json"
	"log"
	"sync"
	"sync/atomic"

	vosk "github.com/alphacep/vosk-api/go"
	"github.com/gordonklaus/portaudio"
	"github.com/pkg/errors"
)

// SpeechRecognizer reads audio input from microphone and translates speech into text.
type SpeechRecognizer struct {
	recognizer         *vosk.VoskRecognizer
	microphoneStream   *portaudio.Stream
	audioSamplesBuffer []int16
	stopConvertSignal  chan struct{}
	stopCaptureSignal  chan struct{}
	audioInputChan     chan []byte
	textOutputChan     chan string
	wg                 *sync.WaitGroup
	listeningEnabled   atomic.Bool
}

// GetTextOuputStream returns a channel of phrases transcribed from microphone input.
func (r *SpeechRecognizer) GetTextOuputStream() <-chan string {
	return r.textOutputChan
}

// StartListening enable microphone audio capture and transcription.
func (r *SpeechRecognizer) StartListening() {
	r.listeningEnabled.Store(true)
}

// StopListening disable microphone audio capture and transcription.
func (r *SpeechRecognizer) StopListening() {
	r.listeningEnabled.Store(false)
}

// IsListening returns true if microphone audio capture and transcription is running.
func (r *SpeechRecognizer) IsListening() bool {
	return r.listeningEnabled.Load()
}

// Close must be called when done using SpeechRecognizer.
func (r *SpeechRecognizer) Close() {
	r.stopConvertSignal <- struct{}{}
	r.stopCaptureSignal <- struct{}{}
	r.wg.Wait()

	close(r.textOutputChan)
	close(r.audioInputChan)

	deinitAudioCapture(r.microphoneStream)
	deinitVosk(r.recognizer)
}

// processAudioChunk returns ready = true if silence was reached and a phrase can be retrieved by calling getResults.
func (r *SpeechRecognizer) processAudioChunk(audioData []byte) (ready bool, err error) {
	result := r.recognizer.AcceptWaveform(audioData)

	// 1 if silence is occured and you can retrieve a new utterance with result method
	// 0 if decoding continues
	// -1 if exception occured

	switch result {
	case 1:
		return true, nil
	case 0:
		return false, nil
	case -1:
		return false, errors.New("exception occured when processing audio")
	default:
		return false, errors.New("unexpected recognizer.AcceptWaveform result")
	}
}

// getResults returns results from processing audio chunks.
func (r *SpeechRecognizer) getResults() (string, error) {
	var phrase resultPhrase
	if err := json.Unmarshal([]byte(r.recognizer.Result()), &phrase); err != nil {
		return "", errors.Wrap(err, "failed to unmarshal result phrase")
	}

	return phrase.Text, nil
}

// getFinalResults returns results from processing audio chunks (without waiting for silence).
func (r *SpeechRecognizer) getFinalResults() (string, error) {
	var phrase resultPhrase
	if err := json.Unmarshal([]byte(r.recognizer.FinalResult()), &phrase); err != nil {
		return "", errors.Wrap(err, "failed to unmarshal result phrase")
	}

	return phrase.Text, nil
}

func (r *SpeechRecognizer) convertSpeechToText() {
	defer r.wg.Done()

	for audioData := range r.audioInputChan {
		ready, err := r.processAudioChunk(audioData)
		if err != nil {
			// TODO
			panic(err)
		}

		if !ready {
			continue
		}

		text, err := r.getResults()
		if err != nil {
			// TODO
			panic(err)
		}

		select {
		case r.textOutputChan <- text:
		case <-r.stopConvertSignal:
			return
		}
	}

	text, err := r.getFinalResults()
	if err != nil {
		// TODO
		panic(err)
	}

	r.textOutputChan <- text
}

func (r *SpeechRecognizer) captureAudioFromMicrophone() {
	defer r.wg.Done()

	if startErr := r.microphoneStream.Start(); startErr != nil {
		// TODO terminate instead of log or panic ???
		log.Fatalf("failed to start microphone stream, err: %s\n", startErr.Error())
	}

loop:
	for {
		if readErr := r.microphoneStream.Read(); readErr != nil {
			// TODO terminate instead of log or panic ???
			log.Printf("failed to read from microphone stream, err: %s\n", readErr.Error())
			continue
		}

		// TODO make this better and don't do needless work of reading microphone stream if listening is disabled
		if r.listeningEnabled.Load() {
			r.audioInputChan <- int16ToBytes(r.audioSamplesBuffer)
		}

		select {
		case <-r.stopCaptureSignal:
			break loop
		default:
		}
	}

	if stopErr := r.microphoneStream.Stop(); stopErr != nil {
		// TODO terminate instead of log or panic ???
		log.Fatalf("failed to stop microphone stream, err: %s\n", stopErr.Error())
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

// NewSpeechRecognizer initializes an instance of SpeechRecognizer
// On program exit SpeechRecognizer.Close must be called to deinitialize.
func NewSpeechRecognizer(voskModelPath string) (*SpeechRecognizer, error) {
	recognizer, err := initVosk(voskModelPath)
	if err != nil {
		return nil, errors.Wrap(err, "failed to init vosk")
	}

	microphoneStream, audioSamplesBuffer, err := initAudioCapture()
	if err != nil {
		return nil, errors.Wrap(err, "failed to init audio capture")
	}

	speechRecognizer := &SpeechRecognizer{
		recognizer:         recognizer,
		microphoneStream:   microphoneStream,
		audioSamplesBuffer: audioSamplesBuffer,
		stopConvertSignal:  make(chan struct{}),
		stopCaptureSignal:  make(chan struct{}),
		audioInputChan:     make(chan []byte, 1000), // TODO what buffer size to use?
		textOutputChan:     make(chan string, 1000),
		wg:                 &sync.WaitGroup{},
		listeningEnabled:   atomic.Bool{},
	}

	speechRecognizer.wg.Add(2)

	go speechRecognizer.convertSpeechToText()
	go speechRecognizer.captureAudioFromMicrophone()

	return speechRecognizer, nil
}
