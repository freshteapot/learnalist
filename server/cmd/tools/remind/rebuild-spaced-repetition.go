package remind

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/freshteapot/learnalist-api/server/api/database"
	"github.com/freshteapot/learnalist-api/server/pkg/logging"
	"github.com/freshteapot/learnalist-api/server/pkg/remind"
	"github.com/freshteapot/learnalist-api/server/pkg/spaced_repetition"
	userStorage "github.com/freshteapot/learnalist-api/server/pkg/user/sqlite"
)

var rebuildSpacedRepetitionCMD = &cobra.Command{
	Use:   "rebuild-spaced-repetition",
	Short: "Rebuild reminders, based on spaced_repetition table",
	Long: `
Rebuild reminder table for spaced repetitions

- Assumes spaced_repetition has been populated.
- Deletes all current entries in spaced_repetition_reminder (lastActive will be fake).
- Rebuilds spaced_repetition_reminder for each user found in spaced_repetition table with lastActive based on now.

sqlite3 other.db "DELETE FROM spaced_repetition"
sqlite3 server.db .dump | grep 'INSERT INTO spaced_repetition' | sqlite3 other.db
	`,
	Run: func(cmd *cobra.Command, args []string) {
		logger := logging.GetLogger()

		viper.SetDefault("db", "")
		viper.BindEnv("db", "DB")

		databaseName := viper.GetString("db")
		db := database.NewDB(databaseName)

		userInfoRepo := userStorage.NewUserInfo(db)
		spacedRepetitionRepo := spaced_repetition.NewSqliteRepository(db)
		remindSpacedRepetitionRepo := remind.NewRemindSpacedRepetitionSqliteRepository(db)

		manager := remind.NewSpacedRepetition(
			userInfoRepo,
			spacedRepetitionRepo,
			remindSpacedRepetitionRepo,
			logger.WithField("context", "spaced-repetition-reminder"))

		logContext := logger.WithField("context", "spaced-repetition-reminder-rebuild")
		//lastActive := time.Now().UTC().Add(-6 * time.Minute)
		lastActive := time.Now().UTC()

		var users []string
		err := db.Select(&users, `SELECT DISTINCT(user_uuid) FROM spaced_repetition`)
		if err != nil {
			fmt.Println(err)
			panic("...")
		}

		db.MustExec("DELETE FROM spaced_repetition_reminder")

		for _, userUUID := range users {
			manager.CheckForNextEntryAndSetReminder(logContext, userUUID, lastActive)
		}
	},
}
