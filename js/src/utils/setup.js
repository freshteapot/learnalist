function getApiServer() {
    const apiServer = document.querySelector('meta[name="api.server"]');
    return apiServer ? apiServer.content : "https://learnalist.net";
}

export {
    getApiServer
}
