# Sanity client in Go

> **Under development!** For developers *with an adventurous spirit only*.

This is a client for [Sanity](https://www.sanity.io) written in Go.

## Using

See the [API reference](https://godoc.org/github.com/sanity-io/client-go) for the full documentation.


```go
package main

import (
  "log"
  "context"

  sanity "github.com/sanity-io/client-go"
)

func main() {
  client, err := sanity.New("zx3vzmn!",
    sanity.WithCallbacks(sanity.Callbacks{
      OnQueryResult: func(result *sanity.QueryResult) {
        log.Printf("Sanity queried in %d ms!", result.Ms)
      },
    }),
    sanity.WithToken("mytoken"),
    sanity.WithDataset("production"))
  if err != nil {
    log.Fatal(err)
  }

  var project struct {
    ID    string `json:"_id"`
    Title string
  }
  queryBuilder = client.Query(context.Background(), "*[_type == 'project' && _id == $id][0]")
  queryBuilder.param("id", 123)

  queryResult, err := queryBuilder.Do(); 
  queryResult.Unmarshal(&project)

  if err != nil  {
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
