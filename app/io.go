package app

import (
	"io"
	"io/ioutil"
	"os"
)

type nopCloser struct {
	io.Writer
}

func (nopCloser) Close() error { return nil }

func newNopCloser(w io.Writer) io.WriteCloser {
	return nopCloser{w}
}

func createIO(stdin io.Reader, stdout io.Writer, inputFile, outputFile string) (io.ReadCloser, io.WriteCloser, error) {
	input := ioutil.NopCloser(stdin)
	if inputFile != "" {
		f, err := os.Open(inputFile)
		if err != nil {
			return nil, nil, err
		}
		input = f
	}

	output := newNopCloser(stdout)
	if outputFile != "" {
		f, err := os.OpenFile(outputFile, os.O_WRONLY|os.O_CREATE, 0644)
		if err != nil {
			input.Close()
			return nil, nil, err
		}
		output = f
	}

	return input, output, nil
}
