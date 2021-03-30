
function convert(input) {
    const aList = input.detail;
    if (aList.info.type !== "v2") {
        throw "Not v2";
    }
    aList.info.from = input.metadata;
    return aList;
}

export default {
    key: "learnalist",
    convert,
    url: "https://learnalist.net",
    domain: "learnalist.net",
}
