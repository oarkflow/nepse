package utils

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"io"
	"log"
	"os"
	"strconv"

	"github.com/klauspost/compress/gzip"
)

func ToCompressedString(data []byte) string {
	var compressed bytes.Buffer
	w := gzip.NewWriter(&compressed)
	_, _ = w.Write(data)
	_ = w.Close()
	return base64.StdEncoding.EncodeToString(compressed.Bytes())
}

func FromCompressedString(data string) ([]byte, error) {
	decodedBytes, err := base64.StdEncoding.DecodeString(data)
	if err != nil {
		return nil, err
	}
	zipReader, err := gzip.NewReader(bytes.NewReader(decodedBytes))
	if err != nil {
		return nil, err
	}
	rawBytes, err := io.ReadAll(zipReader)
	if err != nil {
		return nil, err
	}
	return rawBytes, nil
}

// GenerateGoFile generates a Go file with the CSV data.
func GenerateGoFile(filename, packageName, varName string, dataMap map[string][]map[string]any) error {
	var buf bytes.Buffer
	buf.WriteString(fmt.Sprintf("package %s\n\n", packageName))
	buf.WriteString(fmt.Sprintf("var %s = map[string][]map[string]any{\n", varName))
	for date, dataList := range dataMap {
		buf.WriteString(fmt.Sprintf("    \"%s\": {\n", date))
		for _, entry := range dataList {
			buf.WriteString("        {\n")
			for key, value := range entry {
				if strValue, ok := value.(string); ok {
					compressedValue := ToCompressedString([]byte(strValue))
					buf.WriteString(fmt.Sprintf("            \"%s\": \"%s\",\n", key, compressedValue))
				} else {
					buf.WriteString(fmt.Sprintf("            \"%s\": %v,\n", key, value))
				}
			}
			buf.WriteString("        },\n")
		}
		buf.WriteString("    },\n")
	}
	buf.WriteString("}\n")

	return os.WriteFile(filename, buf.Bytes(), 0644)
}

func GenerateBinaryContent(packageName, varName string, data []byte, fileName ...string) []byte {
	encoded := ToCompressedString(data)
	output := &bytes.Buffer{}
	output.WriteString("package " + packageName + "\n\nvar " + varName + " = " + strconv.Quote(encoded) + "\n")
	bt := output.Bytes()
	if len(fileName) > 0 {
		writeFile(fileName[0], bt)
	}
	return bt
}

func DecodeBinaryString(data string) ([]byte, error) {
	return FromCompressedString(data)
}

func writeFile(filePath string, data []byte) {
	err := os.WriteFile(filePath, data, os.FileMode(0o664))
	if err != nil {
		log.Fatalf("Error writing '%s': %s", filePath, err)
	}
}
