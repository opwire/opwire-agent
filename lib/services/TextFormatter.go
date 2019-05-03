package services

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/opwire/opwire-agent/lib/utils"
)

type TextFormatter struct {
	format  string
	width   int
	printer func(w http.ResponseWriter, label string, data interface{})
}

type TextFormatterOptions interface {
	GetFormat() string
}

func NewTextFormatter(options TextFormatterOptions) *TextFormatter {
	f := new(TextFormatter)
	format := ""
	if options != nil {
		format = options.GetFormat()
	}
	switch format {
	case "yaml":
		f.format = format
	default:
		f.format = "free"
		f.printer = f.freeStylePrinter
	}
	return f
}

func (p *TextFormatter) PrintTextgraph(w http.ResponseWriter, label string, data interface{}) {
	p.printer(w, label, data)
}

func (p *TextFormatter) PrintCollection(w http.ResponseWriter, label string, listData []string) {
	if len(listData) > 0 {
		lines := utils.Map(listData, func(s string, i int) string {
			return fmt.Sprintf("%d) %s", (i + 1), s)
		})
		section := strings.Join(lines, "\n")
		p.printer(w, label, section)
	}
}

func (p *TextFormatter) PrintJsonString(w http.ResponseWriter, label string, textData string) error {
	if len(textData) > 0 {
		dataMap := make(map[string]interface{}, 0)
		err := json.Unmarshal([]byte(textData), &dataMap)
		if err == nil {
			return p.PrintJsonObject(w, label, dataMap)
		} else {
			p.printer(w, label+" (text)", textData)
		}
	}
	return nil
}

func (p *TextFormatter) PrintJsonObject(w http.ResponseWriter, label string, hashData map[string]interface{}) error {
	if len(hashData) > 0 {
		dataJson, err := json.MarshalIndent(hashData, "", "  ")
		if err != nil {
			return err
		}
		if len(dataJson) == 0 {
			return fmt.Errorf("Marshalling failed, data is empty")
		}
		p.printer(w, label, dataJson)
	}
	return nil
}

func (p *TextFormatter) freeStylePrinter(w http.ResponseWriter, label string, data interface{}) {
	width := p.width
	if width == 0 {
		width = 80
	}
	header := utils.PadString("["+label, utils.LEFT, width, "-")
	footer := utils.PadString(label+"]", utils.RIGHT, width, "-")
	if !utils.IsEmpty(data) {
		io.WriteString(w, fmt.Sprintf("\n%s\n%s\n%s\n", header, data, footer))
	}
}
