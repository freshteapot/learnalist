package user

import (
	"encoding/json"
	"fmt"

	"github.com/freshteapot/learnalist-api/server/api/database"
	userSqlite "github.com/freshteapot/learnalist-api/server/pkg/user/sqlite"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var scratchCMD = &cobra.Command{
	Use:   "scratch",
	Short: "Temp code",
	Long: `

	go run --tags="json1"  main.go --config=../config/dev.config.yaml tools user scratch
	`,
	Run: func(cmd *cobra.Command, args []string) {
		dsn := viper.GetString("server.sqlite.database")

		db := database.NewDB(dsn)
		storage := userSqlite.NewSqliteManagementStorage(db)

		search := "iamtest1"
		r, _ := storage.FindUserUUID(search)
		if len(r) == 0 {
			fmt.Println("not found")
			return
		}
		userUUID := r[0]
		b, err := storage.GetInfo(userUUID)
		fmt.Println(string(b), err)

		rawJSON := `
{
	"app_settings": {
		"plank:v1": {
			"showIntervals": true,
			"intervalTime": 15
		},
		"remind:v1": {
			"time_of_day": "20:00:00",
			"offset": "+2:00"
		}
	},
	"daily_notifications": {
		"plank:v1": {
			"time_of_day": "08:00:00",
			"offset": "+2:00"
		},
		"remind:v1": {
			"time_of_day": "20:00:00",
			"offset": "+2:00"
		}
	}
}`
		storage.SaveInfo(userUUID, []byte(rawJSON))

		b, _ = storage.GetInfo(userUUID)
		fmt.Println(string(b))

		// How to remove remind.v1
		// storage.RemoveInfo(userUUID, `remind:v1`)
		// storage.RemoveInfo(userUUID, `remind`)
		storage.RemoveInfo(userUUID, `apps`)
		b, _ = storage.GetInfo(userUUID)
		fmt.Println(string(b))

		var obj UserInfoExtra
		json.Unmarshal(b, &obj)
		fmt.Println(obj.Apps.RemindV1.TimeOfDay)
		fmt.Println(obj.Apps.PlankV1.IntervalTime)

		fmt.Println(obj.DailyNotifications.RemindV1.TimeOfDay)
		fmt.Println(obj.DailyNotifications.PlankV1.TimeOfDay)
		// notification/daily/{uuid}
		// uuid = "user:remind:v1"
		// 1) Save to user info
		// 2) Save to daily_reminder table
		// 3) Remove from daily_reminder table
		// 4) Remove from user info

		// 5) Not apps but daily
		// 6) Add UI in website to delete notifications
		// 7) Build engine to consume daily_notification settings to send notifications
		// 8) Send message
		// 9) Refactor: If no activity, send message. If activity send well done, keep at it
		// Can I store activity in user_info, having it work is better than optimal
		// if trying, use a different table
		// user_activity: user_uuid, kind, ext_id, when
		// Or query the source to last created
		// could build last_active from the log (as its daily)
	},
}

// RemindV1 is linked to Spaced learning mobile app
// Could add timesone
type UserInfoExtra struct {
	Apps struct {
		RemindV1 RemindV1 `json:"remind:v1"` // Not needed yet
		PlankV1  PlankV1  `json:"plank:v1"`  // Only nice to sync between app and web, not needed yet
	} `json:"app_settings"` // TODO good to know, but lets not run with it yet
	DailyNotifications struct {
		RemindV1 RemindV1 `json:"remind:v1"` // Needed first :D
		PlankV1  RemindV1 `json:"plank:v1"`
	} `json:"daily_notifications"`
	LastActive struct {
		Plank            string `json:"plank"`             // UTC int64? or string time.RFC3339Nano
		SpacedRepetition string `json:"spaced_repetition"` // UTC int64? or string time.RFC3339Nano
	} `json:"last_active"`
}

type RemindV1 struct {
	TimeOfDay string `json:"time_of_day"`
	Offset    string `json:"offset"`
	// push, email (might be overkill for now)
	Medium []string `json:"medium"` // start with push
	// Not sure if I want this, or need to verify the email
	Email         string `json:"email,omitempty"`          // Maybe have a different workflow here or let them pick, first version different workflow
	EmailVerified string `json:"email_verified,omitempty"` // 0 or 1
	Token         string `json:"token,omitempty"`          // Device token, might not be needed as it should be in the register table
}

// Taken from // ChallengePlankRecord struct for ChallengePlankRecord
type PlankV1 struct {
	ShowIntervals bool  `json:"showIntervals"`
	IntervalTime  int32 `json:"intervalTime"`
}
