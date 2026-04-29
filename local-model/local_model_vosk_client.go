package local_model

import (
	"encoding/json"

	vosk "github.com/alphacep/vosk-api/go"
	stt "github.com/beefllama/speech-to-text-go"
	"github.com/pkg/errors"
)

// localModelVoskClient implements voskClient interface.
type localModelVoskClient struct {
	recognizer *vosk.VoskRecognizer
}

func (c *localModelVoskClient) ProcessAudioChunk(audioData []byte) (ready bool, err error) {
	result := c.recognizer.AcceptWaveform(audioData)

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

func (c *localModelVoskClient) GetResults() (string, error) {
	var phrase resultPhrase
	if err := json.Unmarshal([]byte(c.recognizer.Result()), &phrase); err != nil {
		return "", errors.Wrap(err, "failed to unmarshal result phrase")
	}

	return phrase.Text, nil
}

func (c *localModelVoskClient) GetFinalResults() (string, error) {
	var phrase resultPhrase
	if err := json.Unmarshal([]byte(c.recognizer.FinalResult()), &phrase); err != nil {
		return "", errors.Wrap(err, "failed to unmarshal result phrase")
	}

	return phrase.Text, nil
}

func (c *localModelVoskClient) Close() {
	deinitVosk(c.recognizer)
}

func newLocalModelVoskClient(localModelPath string) (stt.VoskClient, error) {
	recognizer, err := initVosk(localModelPath)
	if err != nil {
		return nil, errors.Wrap(err, "failed to init vosk")
	}

	return &localModelVoskClient{recognizer: recognizer}, nil
}

// NewSpeechRecognizer initializes an instance of SpeechRecognizer that uses local vosk model.
// At some point before program exit, SpeechRecognizer.Close must be called to deinitialize.
func NewSpeechRecognizerWithLocalVoskModel(voskModelPath string) (*stt.SpeechRecognizer, error) {
	voskClient, err := newLocalModelVoskClient(voskModelPath)
	if err != nil {
		return nil, errors.Wrap(err, "failed to init vosk client")
	}

	return stt.NewSpeechRecognizer(voskClient)
}
