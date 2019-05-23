package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/freshteapot/learnalist-api/api/alist"
)

/*
cat dataset/capitals.txt | awk -F'\t' '{print "from:"$1"::to:"$2}' | go run integrations/convert/main.go -type=v2 | python -mjson.tool > dataset/countries.json

cat dataset/months.en.no.txt | awk -F'\t' '{print "from:"$1"::to:"$2}' | go run integrations/convert/main.go -type=v2 | python -mjson.tool > dataset/months.json
*/
func main() {
	listType := flag.String("type", "", "Which list type?")
	flag.Parse()

	data, err := ioutil.ReadAll(os.Stdin)
	check(err)
	rawStrings := string(data)
	rows := strings.Split(rawStrings, "\n")
	var aList *alist.Alist
	if *listType == alist.FromToList {
		aList = parseTypeV2(rows)
	}
	jsonBytes, _ := aList.MarshalJSON()
	fmt.Println(string(jsonBytes))
}

func parseTypeV2(rows []string) *alist.Alist {
	aList := alist.NewTypeV2()
	data := aList.Data.(alist.TypeV2)

	for _, row := range rows {
		row = strings.TrimSpace(row)
		if row == "" {
			continue
		}
		if !strings.HasPrefix(row, "from:") {
			continue
		}

		if !strings.Contains(row, "::to:") {
			continue
		}

		parts := strings.Split(row, "::to:")
		if len(parts) != 2 {
			continue
		}

		from := strings.TrimPrefix(parts[0], "from:")
		to := parts[1]

		from = strings.Trim(from, " ")
		to = strings.Trim(to, " ")
		item := alist.TypeV2Item{
			From: from,
			To:   to,
		}
		data = append(data, item)
	}
	aList.Data = data
	return aList
}

func check(e error) {
	if e != nil {
		panic(e)
	}
}
