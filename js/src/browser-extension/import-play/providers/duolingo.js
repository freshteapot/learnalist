
function convert(input) {
    const data = input.detail;
    return {
        info: {
            title: data.title,
            type: "v2",
            from: input.metadata,
        },
        data: data.listData,
    };
}

export default {
    key: "duolingo",
    convert,
    url: "https://www.duolingo.com",
    domain: "duolingo.com",
}
