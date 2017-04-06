// Package dictionary contains a interface and wraps all io related work.
// It is used by jiebago module to read/write files.
package dictionary

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

// DictLoader is the interface that could add one token or load
// tokens from channel.
type DictLoader interface {
	Load(<-chan Token)
	AddToken(Token)
}

var dict_reg = regexp.MustCompile(`^(.+?)( [0-9]+)?( [a-z]+)?$`)

func loadDictionary(file *os.File) (<-chan Token, <-chan error) {
	tokenCh, errCh := make(chan Token), make(chan error, 1)

	go func() {
		defer close(tokenCh)
		defer close(errCh)
		scanner := bufio.NewScanner(file)
		var token Token
		var line string
		var err error
		for scanner.Scan() {
			line = scanner.Text()
			fields := dict_reg.FindStringSubmatch(strings.TrimSpace(line))
			if len(fields) == 0 {
				continue
			}
			fields = fields[1:]
			token.text = strings.TrimSpace(strings.Replace(fields[0], "\ufeff", "", 1))
			if fields[1] != "" {
				token.frequency, err = strconv.ParseFloat(strings.TrimSpace(fields[1]), 64)
				if err != nil {
					errCh <- fmt.Errorf("Error:%v, line(%s)", err, line)
					return
				}
				if fields[2] != "" {
					token.pos = strings.TrimSpace(fields[2])
				}
			}
			tokenCh <- token
		}

		if err = scanner.Err(); err != nil {
			errCh <- err
		}
	}()
	return tokenCh, errCh

}

// LoadDictionary reads the given file and passes all tokens to a DictLoader.
func LoadDictionary(dl DictLoader, fileName string) error {
	filePath, err := dictPath(fileName)
	if err != nil {
		return err
	}
	dictFile, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer dictFile.Close()
	tokenCh, errCh := loadDictionary(dictFile)
	dl.Load(tokenCh)

	return <-errCh

}

func dictPath(dictFileName string) (string, error) {
	if filepath.IsAbs(dictFileName) {
		return dictFileName, nil
	}
	var dictFilePath string
	cwd, err := os.Getwd()
	if err != nil {
		return dictFilePath, err
	}
	dictFilePath = filepath.Clean(filepath.Join(cwd, dictFileName))
	return dictFilePath, nil
}
