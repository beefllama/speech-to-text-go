package stt

import (
	"github.com/gordonklaus/portaudio"
	"github.com/pkg/errors"
)

func initAudioCapture() (*portaudio.Stream, []int16, error) {
	if initErr := portaudio.Initialize(); initErr != nil {
		return nil, nil, errors.Wrap(initErr, "failed to init portaudio")
	}

	samplesBuffer := make([]int16, 64)
	microphoneStream, err := portaudio.OpenDefaultStream(1, 0, 16000, len(samplesBuffer), samplesBuffer)
	if err != nil {
		return nil, nil, errors.Wrap(err, "failed to open default portaudio stream")
	}

	return microphoneStream, samplesBuffer, nil
}

func deinitAudioCapture(microphoneStream *portaudio.Stream) {
	microphoneStream.Close()
	portaudio.Terminate()
}
