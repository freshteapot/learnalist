let basicAuth = "";
let renderScreen = "options";
let server = "https://learnalist.net/api";
let config = {
  server: server,
  headers: {
    Authorization: "Basic " + basicAuth
  },
  renderScreen: renderScreen
};

function loadState() {
  let basicAuth = localStorage.getItem("basicAuth");
  if (basicAuth) {
    config.headers.Authorization = "Basic " + basicAuth;
  } else {
    renderScreen = "options";
    return;
  }

  let server = localStorage.getItem("server");
  if (server) {
    config.server = server;
  }

  let screen = localStorage.getItem("renderScreen");
  if (screen) {
    config.renderScreen = screen;
  }
}

loadState();

function getAuth() {
  if (localStorage.hasOwnProperty('bearerAuth')) {
    return "Bearer " + localStorage.getItem("bearerAuth");
  }

  if (localStorage.hasOwnProperty('basicAuth')) {
    return "Basic " + localStorage.getItem("basicAuth");
  }

  return "";
}

function getServer() {
  return localStorage.getItem("server");
}

export { config, getAuth, getServer };
