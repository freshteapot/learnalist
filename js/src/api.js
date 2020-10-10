import {
  getConfiguration,
  KeyUserAuthentication,
  KeySettingsServer
} from './configuration.js';
import {
  Configuration,
  DefaultApi,
  UserApi,
  AListApi,
  PlankApi,
  SpacedRepetitionApi,
  HttpUserLoginRequestFromJSON,
  AlistInputFromJSON,
  AlistFromJSON,
  SpacedRepetitionEntryViewedFromJSON,
  PlankFromJSON
} from "./openapi";

const Services = {
  Default: DefaultApi,
  User: UserApi,
  Alist: AListApi,
  SpacedRepetition: SpacedRepetitionApi,
  Plank: PlankApi
}

function getServer() {
  const server = getConfiguration(KeySettingsServer, null)
  if (server === null) {
    throw new Error('settings.server.missing');
  }
  return server;
}

// getApi service = One of the services based on Services
function getApi(service) {
  var config = new Configuration({
    basePath: `${getServer()}/api/v1`,
    accessToken: getConfiguration(KeyUserAuthentication, undefined),
  });

  return new service(config);
}


async function postLogin(username, password) {
  const api = getApi(Services.User);
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
}

async function getListsByMe(filter) {
  const api = getApi(Services.Alist);
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

async function addList(aList) {
  try {
    const api = getApi(Services.Alist);

    const input = {
      alistInput: AlistInputFromJSON(aList)
    }
    return await api.addList(input);
  } catch (error) {
    console.error(error);
    throw new Error("Failed to save list");
  }
}

async function updateList(aList) {
  try {
    const api = getApi(Services.Alist);

    const input = {
      uuid: aList.uuid,
      alist: AlistFromJSON(aList)
    }
    return await api.updateListByUuid(input);
  } catch (error) {
    console.error(error);
    throw new Error("Failed to update list");
  }
}

async function deleteList(uuid) {
  try {
    const api = getApi(Services.Alist);
    const input = {
      uuid: uuid,
    }
    return await api.deleteListByUuid(input);
  } catch (error) {
    console.error(error);
    throw new Error("Failed to delete list");
  }
}

async function getServerVersion() {
  const api = getApi(Services.SpacedRepetition);
  return await api.getServerVersion();
}

async function getSpacedRepetitionEntries() {
  const api = getApi(Services.SpacedRepetition);
  try {
    return await api.getSpacedRepetitionEntries();
  } catch (error) {
    throw error;
  }
}
async function getSpacedRepetitionNext() {
  const api = getApi(Services.SpacedRepetition);

  const response = {
    status: 404,
    body: {}
  }

  try {
    const res = await api.getNextSpacedRepetitionEntryRaw();
    response.status = res.raw.status;
    if (response.status === 200) {
      response.body = await res.value();
    }

  } catch (responseError) {
    response.status = responseError.status;
  }
  return response;
}


async function addSpacedRepetitionEntry(entry) {
  const response = {
    status: 500,
    body: {}
  }

  try {
    const api = getApi(Services.SpacedRepetition);
    const input = {
      body: entry,
    }
    const res = await api.addSpacedRepetitionEntryRaw(input);
    response.status = res.raw.status;
    response.body = await res.value();
    return response;
  } catch (error) {
    response.status = error.status;
    response.body = await error.json();
    return response;
  }
}

async function updateSpacedRepetitionEntry(entry) {
  try {
    const api = getApi(Services.SpacedRepetition);

    const input = {
      spacedRepetitionEntryViewed: SpacedRepetitionEntryViewedFromJSON(entry)
    }
    return api.updateSpacedRepetitionEntry(input);
  } catch (error) {
    console.log("error", error);
    throw new Error("Failed to get lists by me");
  }
}


async function getPlankHistoryByUser() {
  const api = getApi(Services.Plank);
  try {
    return await api.getPlankHistoryByUser();
  } catch (error) {
    throw new Error({
      message: "Failed to get planks",
      error: error
    });
  }
}

async function addPlankEntry(entry) {
  try {
    const api = getApi(Services.Plank);
    const input = {
      plank: PlankFromJSON(entry),
    }
    return api.addPlankEntry(input);
  } catch (error) {
    console.log("error", error);
    throw new Error("Failed to save plank");
  }
}

async function deletePlankEntry(uuid) {
  try {
    const api = getApi(Services.Plank);
    const input = {
      uuid: uuid,
    }
    return api.deletePlankEntry(input);
  } catch (error) {
    console.log("error", error);
    throw new Error("Failed to delete plank");
  }
}

export {
  getServer,
  postLogin,
  getListsByMe,
  addList,
  updateList,
  deleteList,
  getPlankHistoryByUser,
  addPlankEntry,
  deletePlankEntry,
  getServerVersion,
  getSpacedRepetitionEntries,
  getSpacedRepetitionNext,
  addSpacedRepetitionEntry,
  updateSpacedRepetitionEntry
};
