package proxy
import (
	"io/ioutil"
	"encoding/json"
)

// Config is the configuration base structure
type Config struct {
	Port          int               `json:"port"`
	SSLRewrite    []string          `json:"sslRewrite"`
	Modifications []Modification    `json:"modifications"`
}

// Modifications describes each modification to be executed
type Modification struct {
	URLMatch  string    `json:"urlMatch"`
	Selector  string    `json:"selector"`
	Index     int       `json:"index"`
	Attribute string    `json:"attribute"`
	Wrapper   string    `json:"wrapper"`
	AppendTo  string    `json:"appendTo"`
	Replace   string    `json:"replace"`
	Trim      bool      `json:"trim"`
}

// Replace or Append options
const (
	APPEND = iota
	REPLACE = iota
)

// ReadConfigFromFile reads the configuration from a file
func ReadConfigFromFile(file string) (Config, error) {
	config := Config{
		Port: 8080,
	}
	b, err := ioutil.ReadFile(file)
	if nil != err {
		return config, err
	}
	json.Unmarshal(b, &config)
	return config, nil
}