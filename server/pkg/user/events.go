package user

import (
	"encoding/json"
	"fmt"

	"github.com/freshteapot/learnalist-api/server/pkg/event"
)

func (m management) eventUserRegister(entry event.Eventlog) {
	if entry.Kind != event.ApiUserRegister {
		return
	}

	b, _ := json.Marshal(entry.Data)

	var moment event.EventUser
	json.Unmarshal(b, &moment)

	todo := `
- [ ] Set display Name based on username or based on something from the IDP (based on .kind)
	`

	fmt.Println(string(b))
	fmt.Println(todo)
	fmt.Println(moment)
}
