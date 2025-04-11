package colorize

import (
	"bufio"
	"io"
	"strings"
)

func ColorizeOutput(in io.Reader, out io.Writer) {
	rdr := bufio.NewReader(in)
	for {
		line, err := rdr.ReadString('\n')
		if err != nil {
			panic(err)
		}

		switch {
		case strings.Contains(line, "FAIL"):
			out.Write([]byte("\033[31m"))
		case strings.Contains(line, "PASS"):
			out.Write([]byte("\033[32m"))
		}

		out.Write([]byte(line))
		out.Write([]byte("\033[0m"))
	}
}
