// TODO visualise with colour the daily, weekly, monthly
// TODO visualise days in a row doing a plank
import dayjs from "dayjs";
import isBetween from "dayjs/plugin/isBetween";
import isToday from "dayjs/plugin/isToday";
import en from "dayjs/locale/en";

dayjs.locale({
    ...en,
    weekStart: 1
});
dayjs.extend(isBetween);
dayjs.extend(isToday);


export function totals(entries) {
    return entries.reduce((a, b) => a + (b["timerNow"] || 0), 0);
}

export function todayTotals(entries) {
    return entries.reduce((a, b) => {
        if (!dayjs(b.beginningTime).isToday()) {
            return a;
        }
        return a + (b["timerNow"] || 0);
    }, 0);
}

export function weekTotals(entries) {
    const startOf = dayjs().startOf("week");
    const endOf = dayjs().endOf("week");

    return entries.reduce((a, b) => {
        const now = dayjs(b.beginningTime);
        if (!now.isBetween(startOf, endOf)) {
            return a;
        }
        return a + (b["timerNow"] || 0);
    }, 0);
}

export function monthTotals(entries) {
    const startOf = dayjs().startOf("month");
    const endOf = dayjs().endOf("month");

    return entries.reduce((a, b) => {
        const now = dayjs(b.beginningTime);
        if (!now.isBetween(startOf, endOf)) {
            return a;
        }
        return a + (b["timerNow"] || 0);
    }, 0);
}
