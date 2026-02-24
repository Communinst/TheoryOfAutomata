package main

import (
	"context"
)

func main() {

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	eventCh := make(chan Event)

}
