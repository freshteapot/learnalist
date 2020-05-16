package fix

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/jmoiron/sqlx"
)

type interactV1 struct {
	db *sqlx.DB
}

func NewInteractV1(db *sqlx.DB) interactV1 {
	return interactV1{db: db}
}

func (f interactV1) GetListsToChange() []string {
	db := f.db
	lists := make([]string, 0)
	query := `
	SELECT
		body
	FROM
		alist_kv
	WHERE
		list_type="v1"
	AND (
		json_type(body, '$.info.interact.slideshow')="text"
	OR
		json_type(body, '$.info.interact.totalrecall')="text"
	);`

	err := db.Select(&lists, query)
	if err != nil {
		fmt.Println(err)
		panic("...")
	}
	return lists
}

func (f interactV1) ChangeFromStringToInt() {
	db := f.db
	lists := f.GetListsToChange()

	type listInfoInteract struct {
		Slideshow   string `json:"slideshow"`
		TotalRecall string `json:"totalrecall"`
	}

	type newListInfoInteract struct {
		Slideshow   int `json:"slideshow"`
		TotalRecall int `json:"totalrecall"`
	}

	type listInfo struct {
		Interact listInfoInteract `json:"interact"`
		ListType listInfoInteract `json:"type"`
	}

	type tempList struct {
		Info listInfo `json:"info"`
		UUID string   `json:"uuid"`
	}

	fmt.Printf("Will modify %d lists\n", len(lists))
	for _, body := range lists {
		var temp tempList
		json.Unmarshal([]byte(body), &temp)

		if temp.Info.Interact.Slideshow == "" && temp.Info.Interact.TotalRecall == "" {
			fmt.Println("skipping")
			continue
		}

		slideshow, err := strconv.Atoi(temp.Info.Interact.Slideshow)
		if err != nil {
			slideshow = 0
		}

		totalrecall, err := strconv.Atoi(temp.Info.Interact.TotalRecall)
		if err != nil {
			totalrecall = 0
		}

		interact := newListInfoInteract{
			Slideshow:   slideshow,
			TotalRecall: totalrecall,
		}

		newInteract, _ := json.Marshal(interact)
		newInteractJSON := string(newInteract)

		oldInteract, _ := json.Marshal(temp.Info.Interact)
		oldInteractJSON := string(oldInteract)

		fmt.Println("oldInteractJSON", oldInteractJSON)
		fmt.Println("newInteractJSON", newInteractJSON)
		fmt.Println("")
		// TODO maybe print and not update

		updateQuery := `
UPDATE
alist_kv
SET
body=json_replace(body, '$.info.interact', json(?))
WHERE
uuid=?`

		db.MustExec(updateQuery, newInteractJSON, temp.UUID)
	}

	// Todo need to rebuild hugo
	fmt.Println("Dont forget to rebuild static site")
	fmt.Println("HUGO_EXTERNAL=false  /app/bin/learnalist-cli --config=/etc/learnalist/config.yaml tools rebuild-static-site")
}

func (f interactV1) RemoveInteractFromNonV1() {
	// TODO I should probably have a record of what has been ran
	db := f.db
	query := `
UPDATE
	alist_kv
SET
	body=json_remove(body, '$.info.interact')
WHERE
	list_type != "v1"
`

	db.MustExec(query)
}
