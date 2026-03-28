package stt

import (
	"encoding/binary"
	"encoding/json"
	"sync"
	"sync/atomic"

	vosk "github.com/alphacep/vosk-api/go"
	"github.com/gordonklaus/portaudio"
	"github.com/pkg/errors"
)

const chanBufferSize = 1000

// SpeechRecognizer reads audio input from microphone and translates speech into text.
type SpeechRecognizer struct {
	recognizer         *vosk.VoskRecognizer
	microphoneStream   *portaudio.Stream
	audioSamplesBuffer []int16
	stopCaptureSignal  chan struct{}
	wg                 *sync.WaitGroup
	listeningEnabled   atomic.Bool
}

// StartListening enables microphone audio capture and transcription.
// Returns a channel of phrases transcribed from microphone input and true if listening is enabled, nil and false otherwise.
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

// StopListening disable microphone audio capture and transcription.
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
	deinitVosk(r.recognizer)
}

func (r *SpeechRecognizer) convertSpeechToText(audioInputChan <-chan []byte, textOutputChan chan<- string) {
	defer r.wg.Done()

	for audioData := range audioInputChan {
		ready, err := r.processAudioChunk(audioData)
		if err != nil {
			panic(errors.Wrap(err, "failed to process audio chunk"))
		}

		if !ready {
			continue
		}

		text, err := r.getResults()
		if err != nil {
			panic(errors.Wrap(err, "failed to get results from vosk recognizer"))
		}

		textOutputChan <- text
	}

	text, err := r.getFinalResults()
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
		stopCaptureSignal:  make(chan struct{}),
		wg:                 &sync.WaitGroup{},
		listeningEnabled:   atomic.Bool{},
	}

	return speechRecognizer, nil
}
