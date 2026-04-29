package stt

type VoskClient interface {
	// ProcessAudioChunk returns ready = true if silence was reached and a phrase can be retrieved by calling GetResults.
	ProcessAudioChunk(audioData []byte) (ready bool, err error)
	// GetResults returns results from processing audio chunks.
	GetResults() (string, error)
	// GetFinalResults returns results from processing audio chunks (without waiting for silence).
	GetFinalResults() (string, error)
	// Close closes vosk client. Must be called in SpeechRecognizer.Close
	Close()
}
