# speech-to-text-go

## Prerequisites

### 1. Set up VOSK

Official set up guide for Go here: https://github.com/alphacep/vosk-api/tree/master/go/example

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

3. Create a directory in the root of this project and name it `vosk`.

4. Put vosk & vosk model into the newly created directory.

### Run program arguments

```
VOSK_PATH=/home/beefllama/dev/speech-to-text-go/vosk/vosk-linux-x86_64-0.3.45 LD_LIBRARY_PATH=$VOSK_PATH CGO_CPPFLAGS="-I $VOSK_PATH" CGO_LDFLAGS="-L $VOSK_PATH -ldl" go run ./internal/test/main.go
```