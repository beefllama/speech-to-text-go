package stt

import (
	"encoding/json"
	"net/url"

	"github.com/gorilla/websocket"
	"github.com/pkg/errors"
)

// voskServerClient implements voskClient interface.
type voskServerClient struct {
	conn *websocket.Conn
}

func (c *voskServerClient) ProcessAudioChunk(audioData []byte) (ready bool, err error) {
	err = c.conn.WriteMessage(websocket.BinaryMessage, audioData)
	if err != nil {
		return false, errors.Wrap(err, "failed to write audio data to the server")
	}

	return true, nil
}

func (c *voskServerClient) GetResults() (string, error) {
	_, msg, err := c.conn.ReadMessage()
	if err != nil {
		return "", errors.Wrap(err, "failed to read message from the server")
	}

	var phrase resultPhrase
	if err := json.Unmarshal(msg, &phrase); err != nil {
		return "", errors.Wrap(err, "failed to unmarshal result phrase")
	}

	return phrase.Text, nil
}

func (c *voskServerClient) GetFinalResults() (string, error) {
	err := c.conn.WriteMessage(websocket.TextMessage, []byte(`{"eof" : 1}`))
	if err != nil {
		return "", errors.Wrap(err, "failed to write EOF message to the server")
	}

	return c.GetResults()
}

func (c *voskServerClient) Close() {
	// TODO right now the errors are ignored, maybe log it

	c.conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
	c.conn.Close() // ingore error
}

func newVoskServerClient(host, port string) (voskClient, error) {
	serverURL := url.URL{Scheme: "ws", Host: host + ":" + port, Path: ""}

	// 1. open connection to the server
	conn, _, err := websocket.DefaultDialer.Dial(serverURL.String(), nil)
	if err != nil {
		return nil, errors.Wrap(err, "failed to dial vosk server")
	}

	// 2. pass in config for audio to the server for proper transcription (sample rate)
	config := `{"config": {"sample_rate": 16000}}`
	err = conn.WriteMessage(websocket.TextMessage, []byte(config))
	if err != nil {
		return nil, errors.Wrap(err, "failed to send config to server")
	}

	return &voskServerClient{conn: conn}, nil
}
