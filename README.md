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

## Installation

```
go get github.com/sanity-io/client-go
```

## Requirements

Go 1.13 or later.

# License

See [`LICENSE`](https://github.com/sanity-io/client-go/blob/master/LICENSE) file.
