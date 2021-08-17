package fileutil

import (
	"bufio"
	log "github.com/sirupsen/logrus"
	"os"
	"strconv"
	"strings"
)

type AppConfigProperties struct {
	v map[string]string
}


func (props AppConfigProperties) FetchAsInt(label string) int {

	value, exists := props.v[label]

	if !exists {
		panic("label: "+ label + " is not defined in props")
	}

	v, err := strconv.Atoi(value)
	if err != nil {
		panic(err)
	}

	return v
}

func (props AppConfigProperties) PrintInfo() {
	log.Debugln("AppConfigProperties loaded: ", props.v)
}

func ReadPropertiesFile(filename string) AppConfigProperties {
	config := AppConfigProperties{v: make(map[string]string)}

	if len(filename) == 0 {
		return config
	}
	file, err := os.Open(filename)
	if err != nil {
		panic(err)
	}

	defer func(file *os.File) {
		_ = file.Close()
	}(file)

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {

		line := scanner.Text()

		if !strings.HasPrefix(line, "#") { // Note: if is not a comment

			if equal := strings.Index(line, "="); equal >= 0 {

				if key := strings.TrimSpace(line[:equal]); len(key) > 0 {

					value := ""
					if len(line) > equal {
						value = strings.TrimSpace(line[equal+1:])
					}
					config.v[key] = value
				}
			}

		}
	}

	if err := scanner.Err(); err != nil {
		panic(err)
	}

	return config
}