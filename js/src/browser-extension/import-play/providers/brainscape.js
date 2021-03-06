
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
    key: "brainscape",
    convert,
    url: "https://www.brainscape.com",
    domain: "brainscape.com",
}
