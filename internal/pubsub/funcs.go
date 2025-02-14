package pubsub

import (
	"bytes"
	"encoding/gob"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/routing"
)

func UnmarshallGameLog(data []byte) (routing.GameLog, error) {
	var log routing.GameLog

	var buf bytes.Buffer
	buf.Write(data)

	dec := gob.NewDecoder(&buf)
	err := dec.Decode(&log)

	if err != nil {
		return routing.GameLog{}, err
	}

	return log, nil
}
