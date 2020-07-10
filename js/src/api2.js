import { get as cacheGet, KeyUserAuthentication, KeySettingsServer } from './cache.js';
import { Configuration, DefaultApi, HttpUserLoginRequestFromJSON } from "./openapi";

function getApi() {
  const server = cacheGet(KeySettingsServer, null)
  if (server === null) {
    throw new Error('settings.server.missing');
  }

  var config = new Configuration({
    basePath: `${server}/api/v1`,
    accessToken: cacheGet(KeyUserAuthentication, undefined),
  });

  return new DefaultApi(config);
}


async function postLogin(username, password) {
  const api = getApi();
  const response = {
    status: 400,
    body: {}
  }

  try {
    const input = {
      httpUserLoginRequest: HttpUserLoginRequestFromJSON({ username, password })
    }

    const res = await api.loginWithUsernameAndPasswordRaw(input);
    response.status = res.raw.status;
    response.body = await res.value();
    return response;
  } catch (error) {
    response.status = error.status;
    response.body = await error.json();
    return response;
  }

  /*
  return api.loginWithUsernameAndPasswordRaw(input).then(async data => {
    return {
      status: data.raw.status,
      body: await data.value()
    }
  },
    async error => {
      return {
        status: error.status,
        body: await error.json()
      }
    }
  ).then(response => {
    return response;
  });
  */
}

async function getListsByMe(filter) {
  const api = getApi();
  if (!filter) {
    filter = {};
  }

  try {
    return await api.getListsByMe(filter);
  } catch (error) {
    console.log("error", error);
    throw new Error("Failed to get lists by me");
  }
}

async function getPlanks() {
  const api = getApi();

  try {
    return await api.getListsByMe({ labels: "plank", listType: "v1" });
  } catch (error) {
    console.log("error", error);
    throw new Error("Failed to get lists by me");
  }
}


export {
  postLogin,
  getListsByMe,
  getPlanks,
};
