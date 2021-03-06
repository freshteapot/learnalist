function convert(input) {
    // setID = UUID
    const data = input.detail;
    const listData = data.listData.map((term) => {
        return { from: term.front_plain, to: term.back_plain };
    });

    return {
        info: {
            title: data.title,
            type: "v2",
            from: input.metadata,
        },
        data: listData,
    };
}

export default {
    key: "cram",
    convert: convert,
    url: "https://www.cram.com",
    domain: "cram.com",
}
