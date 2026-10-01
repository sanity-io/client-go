# Sanity client in Go

> **Under development!** For developers *with an adventurous spirit only*.

This is a client for [Sanity](https://www.sanity.io) written in Go.

## Using

See the [API reference](https://godoc.org/github.com/sanity-io/client-go) for the full documentation.

```go
package main

import (
	"context"
	"log"

	sanity "github.com/sanity-io/client-go"
)

func main() {
	client, err := sanity.VersionV20210325.NewClient("zx3vzmn!", sanity.DefaultDataset,
		sanity.WithCallbacks(sanity.Callbacks{
			OnQueryResult: func(result *sanity.QueryResult) {
				log.Printf("Sanity queried in %d ms!", result.Time.Milliseconds())
			},
		}),
		sanity.WithToken("mytoken"))
	if err != nil {
		log.Fatal(err)
	}

	var project struct {
		ID    string `json:"_id"`
		Title string
	}
	result, err := client.
		Query("*[_type == 'project' && _id == $id][0]").
		Param("id", "123").
		Do(context.Background())
	if err != nil {
		log.Fatal(err)
	}

	if err := result.Unmarshal(&project); err != nil {
		log.Fatal(err)
    }

	log.Printf("Project: %+v", project)
}
```

### Listening for changes

`Listen` opens a server-sent event stream of mutations that match a GROQ filter. The caller
owns the reconnect: `Next` returns `io.EOF` after `Close` and when the server ends the stream.

With `EnableResume(true)`, the server puts an ID on every event. Give the last ID to
`LastEventID` on the next attempt, and the stream continues from that position. Resume is
best effort. The server sends a `welcomeback` event when it continues, and a `welcome`
event when it cannot, which means events in the gap are lost.

```go
stream, err := client.Listen("*[_type == 'project']").
    EnableResume(true).
    LastEventID(lastID). // empty on the first attempt
    Do(ctx)
if err != nil {
    log.Fatal(err)
}
defer stream.Close()

for {
    event, err := stream.Next()
    if err != nil {
        lastID = stream.LastEventID() // reconnect from here
        break
    }
    if event.Type == api.ListenEventMutation {
        log.Printf("mutation: %s", *event.Data)
    }
}
```

## Installation

```
go get github.com/sanity-io/client-go
```

## Requirements

Go 1.13 or later.

# License

See [`LICENSE`](https://github.com/sanity-io/client-go/blob/master/LICENSE) file.
