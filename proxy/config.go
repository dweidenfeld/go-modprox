package proxy
import (
	"io/ioutil"
	"encoding/json"
)

type Config struct {
	Port          int               `json:"port"`
	Modifications []Modification    `json:"modifications"`
}

type Modification struct {
	URLMatch  string    `json:"urlMatch"`
	Selector  string    `json:"selector"`
	Index     int       `json:"index"`
	Attribute string    `json:"attribute"`
	Wrapper   string    `json:"wrapper"`
	AppendTo  string    `json:"appendTo"`
	Trim      bool      `json:"trim"`
}

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