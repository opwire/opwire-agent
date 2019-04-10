package services

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"github.com/opwire/opwire-agent/utils"
)

type TextFormatter struct {
	format string
	printer func (w http.ResponseWriter, label string, data interface{})
}

func NewTextFormatter(format string) *TextFormatter {
	f := new(TextFormatter)
	switch(format) {
	case "json", "yaml":
		f.format = format
	default:
		f.format = "free"
		f.printer = freeStylePrinter
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
			p.printer(w, label + " (text)", textData)
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

func freeStylePrinter(w http.ResponseWriter, label string, data interface{}) {
	header := utils.PadString("[" + label, utils.LEFT, 80, "-")
	footer := utils.PadString(label + "]", utils.RIGHT, 80, "-")
	if !utils.IsEmpty(data) {
		io.WriteString(w, fmt.Sprintf("\n%s\n%s\n%s\n", header, data, footer))
	}
}
