package remind

func (m *dailyManager) templatePlankV1(remind RemindMe) (title string, template string) {
	title = "Daily Reminder"
	template = "Today is a fine day for a plank"
	if remind.Activity {
		template = "You planked, nice work!"
	}
	return title, template
}

func (m *dailyManager) templateRemindV1(remind RemindMe) (title string, template string) {
	title = "Daily Reminder"
	template = "What shall we learn today"
	if remind.Activity {
		template = "Nice work!"
	}
	return title, template
}
