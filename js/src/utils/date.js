
// http://jsfiddle.net/j6qJp/1/
// https://stackoverflow.com/users/872981/tomexx
// https://stackoverflow.com/Questions/3605214/javascript-add-leading-zeroes-to-date
const dateYearMonthDay = (when) => {
    const MyDate = new Date(when);
    return MyDate.getFullYear() + '-' + ('0' + (MyDate.getMonth() + 1)).slice(-2) + '-' + ('0' + MyDate.getDate()).slice(-2);
}

export {
    dateYearMonthDay
}
