package fs

import (
	"encoding/json"
	"encoding/xml"
	"os"
)

// DataContainer — универсальный контейнер для JSON/XML данных
type DataContainer map[string]interface{}

// ReadJSON читает и десериализует JSON файл
// Go's json decoder безопасен от выполнения произвольного кода
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
	err = decoder.Decode(&data)
	return data, err
}

// WriteJSON сериализует данные и записывает в JSON файл
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
	encoder.SetIndent("", "  ") // Красивое форматирование с отступами
	return encoder.Encode(data)
}

// XMLData — простая структура для демонстрации работы с XML
type XMLData struct {
	XMLName xml.Name `xml:"root"`
	Content string   `xml:"content"`
}

// ReadXML читает и десериализует XML файл
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

// WriteXML сериализует данные и записывает в XML файл
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
	encoder.Indent("", "  ") // Красивое форматирование с отступами
	return encoder.Encode(data)
}
