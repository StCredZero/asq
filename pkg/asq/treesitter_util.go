package asq

import (
	"errors"
	"github.com/go-enry/go-enry/v2"
	sitter "github.com/smacker/go-tree-sitter"
	"github.com/smacker/go-tree-sitter/golang"
)

var (
	ErrUnsupportedLang = errors.New("unsupported language")
)

// GetTSLanguageFromEnry detects the language of a file using go-enry and returns
// the corresponding tree-sitter language parser. Currently only supports Go.
func GetTSLanguageFromEnry(filename string, contents []byte) (*sitter.Language, error) {
	lang := enry.GetLanguage(filename, contents)
	if lang == "" {
		return nil, errors.New("could not detect language")
	}
	switch lang {
	case "Go":
		return golang.GetLanguage(), nil
	default:
		return nil, ErrUnsupportedLang
	}
}
