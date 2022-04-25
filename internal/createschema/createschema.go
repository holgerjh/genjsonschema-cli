package createschema

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"

	"github.com/holgerjh/genjsonschema"
	"github.com/holgerjh/genjsonschema-cli/internal/merge"
	"gopkg.in/yaml.v2"
)

type CreateSchemaApp struct {
	Arguments *Arguments
}

type Arguments struct {
	SchemaConfig genjsonschema.SchemaConfig
	OutputFile   string
	InputFiles   []string
	MergeOnly    bool
}

func (c *CreateSchemaApp) Run() error {
	inputHandles, err := openAllFiles(c.Arguments.InputFiles)
	if err != nil {
		return fmt.Errorf("failed to open input file(s): %s", err)
	}
	defer closeAllFiles(inputHandles)

	var outputHandle *os.File
	if c.Arguments.OutputFile == "" {
		outputHandle = os.Stdout
	} else {
		outputHandle, err = os.Create(c.Arguments.OutputFile)
		if err != nil {
			return fmt.Errorf("failed to create output file: %s", err)
		}
		defer outputHandle.Close()
	}

	inputReaders := make([]io.Reader, 0)
	for _, v := range inputHandles {
		inputReaders = append(inputReaders, v)
	}

	result, err := CreateSchemaFromFiles(&c.Arguments.SchemaConfig, inputReaders, c.Arguments.MergeOnly)
	if err != nil {
		return fmt.Errorf("failed to create schema: %s", err)
	}
	_, err = outputHandle.Write(result)
	if err != nil {
		return fmt.Errorf("failed to write result: %s", err)
	}
	return nil

}

func openAllFiles(files []string) ([]*os.File, error) {
	handles := make([]*os.File, 0)
	for _, v := range files {
		var handle *os.File
		var err error = nil
		if v == "-" {
			handle = os.Stdin
		} else {
			handle, err = os.Open(v)
			if err != nil {
				closeAllFiles(handles)
				return nil, err
			}
		}
		handles = append(handles, handle)
	}
	return handles, nil
}

func closeAllFiles(handles []*os.File) error {
	var lastErr error
	for _, v := range handles {
		lastErr = v.Close()
	}
	return lastErr
}

func CreateSchemaFromFiles(cfg *genjsonschema.SchemaConfig, files []io.Reader, onlyMerge bool) ([]byte, error) {
	merged, err := loadAndMergeFiles(files)
	if err != nil {
		return nil, err
	}
	b, err := yaml.Marshal(merged)
	if err != nil {
		return nil, err
	}
	if onlyMerge {
		return b, nil
	}
	return genjsonschema.GenerateFromYAML(b, cfg)
}

func loadAndMergeFiles(files []io.Reader) (interface{}, error) {
	loadedFiles, err := loadAllFiles(files)
	if err != nil {
		return nil, err
	}
	merged, err := merge.MergeAllYAML(loadedFiles...)
	if err != nil {
		return nil, err
	}
	return merged, nil
}

func loadAllFiles(files []io.Reader) ([][]byte, error) {
	loadedFiles := make([][]byte, len(files))
	for i, v := range files {
		var err error = nil
		loadedFiles[i], err = ioutil.ReadAll(v)
		if err != nil {
			return nil, err
		}
	}
	return loadedFiles, nil
}
