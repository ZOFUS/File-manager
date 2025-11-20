package fs

import (
	"encoding/json"
	"encoding/xml"
	"os"
)

// Generic map for JSON/XML to allow flexibility
type DataContainer map[string]interface{}

func ReadJSON(path string) (interface{}, error) {
	safePath, err := ResolvePath(path)
	if err != nil {
		return nil, err
	}

	fileMutex.RLock()
	defer fileMutex.RUnlock()

	file, err := os.Open(safePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var data interface{}
	decoder := json.NewDecoder(file)
	// Go's json decoder is safe from code execution, but we can add limits if needed
	err = decoder.Decode(&data)
	return data, err
}

func WriteJSON(path string, data interface{}) error {
	safePath, err := ResolvePath(path)
	if err != nil {
		return err
	}

	fileMutex.Lock()
	defer fileMutex.Unlock()

	file, err := os.Create(safePath)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	return encoder.Encode(data)
}

// Simple XML structure for demonstration
type XMLData struct {
	XMLName xml.Name `xml:"root"`
	Content string   `xml:"content"`
}

func ReadXML(path string) (*XMLData, error) {
	safePath, err := ResolvePath(path)
	if err != nil {
		return nil, err
	}

	fileMutex.RLock()
	defer fileMutex.RUnlock()

	file, err := os.Open(safePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var data XMLData
	decoder := xml.NewDecoder(file)
	err = decoder.Decode(&data)
	return &data, err
}

func WriteXML(path string, data *XMLData) error {
	safePath, err := ResolvePath(path)
	if err != nil {
		return err
	}

	fileMutex.Lock()
	defer fileMutex.Unlock()

	file, err := os.Create(safePath)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := xml.NewEncoder(file)
	encoder.Indent("", "  ")
	return encoder.Encode(data)
}
