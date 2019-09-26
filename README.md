# Sanity client in Go

> **Under development!** For developers *with an adventurous spirit only*.

This is a client for [Sanity](https://www.sanity.io) written in Go.

## Example use

```go
package main

import (
  "log"
  "context"

  sanity "github.com/sanity-io/golang-go"
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
  if err = client.Query(context.Background(), "*[_type == 'project' && _id == $id][0]", &project,
    sanity.Param("id", "123")); err != nil {
    log.Fatal(err)
  }

  log.Printf("Project: %+v", project)
}
```

# License

See [`LICENSE`](https://github.com/sanity-io/client-go/blob/master/LICENSE) file.
