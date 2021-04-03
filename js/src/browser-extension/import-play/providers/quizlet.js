function convert(input) {
    const data = input.detail;
    const listData = Object.values(
        data.listData.setPageData.termIdToTermsMap
    ).map((term) => {
        return { from: term.word, to: term.definition };
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
    key: "quizlet",
    convert,
    url: "https://quizlet.com",
    domain: "quizlet.com",
}
