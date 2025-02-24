package lfs

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"os"
	"sync"
)

// https://github.com/git-lfs/git-lfs/blob/main/docs/custom-transfers.md

type Event any

type Init struct {
	Operation           string `json:"operation"`
	Remote              string `json:"remote"`
	Concurrent          bool   `json:"concurrent"`
	ConcurrentTransfers int    `json:"concurrenttransfers"`
}

func InitOK() any {
	return json.RawMessage(`{}`)
}

func InitErr(err error) any {
	return map[string]any{"error": map[string]any{"code": 1, "message": err.Error()}}
}

type Terminate struct{}

type Upload struct {
	OID    string `json:"oid"`
	Size   int64  `json:"size"`
	Path   string `json:"path"`
	Action *struct {
		Href   string            `json:"href"`
		Header map[string]string `json:"header"`
	} `json:"action"`
}

func UploadComplete(oid string) any {
	return map[string]any{
		"event": "complete",
		"oid":   oid,
	}
}

type Download struct {
	OID    string `json:"oid"`
	Size   int64  `json:"size"`
	Action *struct {
		Href   string            `json:"href"`
		Header map[string]string `json:"header"`
	} `json:"action"`
}

func DownloadComplete(oid, path string) any {
	return map[string]any{
		"event": "complete",
		"oid":   oid,
		"path":  path,
	}
}

func TransferError(oid string, err error) any {
	return map[string]any{
		"event": "complete",
		"oid":   oid,
		"error": map[string]any{"code": 1, "message": err.Error()},
	}
}

func TransferProgress(oid string, bytesSoFar, bytesSinceLast int64) any {
	return map[string]any{
		"event":          "progress",
		"oid":            oid,
		"bytesSoFar":     bytesSoFar,
		"bytesSinceLast": bytesSinceLast,
	}
}

func events(r io.Reader) <-chan Event {
	br := bufio.NewReader(r)
	out := make(chan Event)

	go func() {
		defer close(out)
		for {
			line, _, err := br.ReadLine()
			if err == io.EOF {
				return
			} else if err != nil {
				slog.Error(err.Error())
				continue
			}

			fmt.Fprintf(os.Stderr, "<- %s\n", line)

			var probe struct {
				Event string `json:"event"`
			}
			if err := json.Unmarshal(line, &probe); err != nil {
				slog.Error(err.Error())
				continue
			}

			var event any
			switch probe.Event {
			case "init":
				event = &Init{}
			case "terminate":
				event = &Terminate{}
			case "upload":
				event = &Upload{}
			default:
				slog.Error("unknown lfs event", "event", probe.Event)
				continue
			}

			if err := json.Unmarshal(line, event); err != nil {
				slog.Error(err.Error())
				continue
			}

			out <- event
		}
	}()

	return out
}

type Response any

func responder(w io.Writer) func(Response) {
	var mu sync.Mutex

	return func(resp Response) {
		mu.Lock()
		defer mu.Unlock()

		b, _ := json.Marshal(resp)
		fmt.Fprintf(os.Stderr, "-> %s\n", b)

		w.Write(b)
		fmt.Fprintf(w, "\n")
	}
}

func Begin() (<-chan Event, func(Response)) {
	return events(os.Stdin), responder(os.Stdout)
}
