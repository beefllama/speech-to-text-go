package stt

import (
	vosk "github.com/alphacep/vosk-api/go"
	"github.com/pkg/errors"
)

func initVosk(voskModelPath string) (*vosk.VoskRecognizer, error) {
	// TODO set env variables for vosk programmaticaly ??

	vosk.GPUInit()

	model, err := vosk.NewModel(voskModelPath)
	if err != nil {
		return nil, errors.Wrap(err, "failed to init model")
	}

	// TODO make sample rate configurable
	sampleRate := 16000.0

	rec, err := vosk.NewRecognizer(model, sampleRate)
	if err != nil {
		return nil, errors.Wrap(err, "failed to init recognizer")
	}
	rec.SetWords(1)

	return rec, nil
}

func deinitVosk(rec *vosk.VoskRecognizer) {
	rec.Free()
}
