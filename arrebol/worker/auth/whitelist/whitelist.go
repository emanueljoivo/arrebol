package whitelist

import (
	"bufio"
	"log"
	"os"
)

const (
	WhiteListPath = "WHITE_LIST_PATH"
)

type WhiteList struct {
	list []string
}

func NewWhiteList() WhiteList {
	list, err := loadSourceFile()
	if err != nil {
		log.Fatal(err)
	}
	return WhiteList{list: list}
}

func (l *WhiteList) Contains(workerId string) bool {
	for _, current := range l.list {
		if current == workerId {
			return true
		}
	}
	return false
}

func loadSourceFile() ([]string, error) {
	file, err := os.Open(os.Getenv(WhiteListPath))
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var ids []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		ids = append(ids, scanner.Text())
	}
	return ids, scanner.Err()
}

