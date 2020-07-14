function copyObject(item) {
    return JSON.parse(JSON.stringify(item))
}

function isObjectEmpty(obj) {
    return Object.keys(obj).length === 0 && obj.constructor === Object
}

export {
    copyObject,
    isObjectEmpty
}
