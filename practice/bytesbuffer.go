package main

import (
	"bytes"
	"fmt"
	"io"
	"log"
)

// you might have net.Conn here too
func writeTo(w io.Writer, msg []byte) error {
	_, err := w.Write(msg)
	return err
}

func main() {
	buf := new(bytes.Buffer)
	buf.Write([]byte("Foo"))
	buf.WriteString("Bar")
	fmt.Println(buf.Len())
	fmt.Println(buf.String())

	buf = new(bytes.Buffer)
	if err := writeTo(buf, []byte("Happy New year!")); err != nil {
		log.Fatal(err)
	}
	fmt.Println(buf.String())
}
