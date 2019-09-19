# Mint web framework

it is simple lightweight web framework, it helps to keep the core simple but extensible. Mint uses gorilla mux router.
Mint does not include a database abstraction layer or body validation or anything else

## Installation

To use Mint package, you need to install Go first in your system and set its workspace

```sh
$ go get -u github.com/5anthosh/mint
```

## A Simple Example

In example.go file

```go
package main

import "github.com/5anthosh/mint"

func main() {
  r := mint.New()

	r.GET("/{message}", func(c *mint.Context) {
		c.JSON(200, mint.JSON{
			"message": c.DefaultParam("message", "Hello World !"),
		})
  })

	r.Run(":8080")
}

```

To run the program

```
$ go run example.go
```

## Example

- [nottu](https://github.com/5anthosh/nottu)
