import {
  getConfiguration,
  KeyUserAuthentication,
  KeySettingsServer
} from './configuration.js';

// TODO remove this file
function getAuth() {
  const token = getConfiguration(KeyUserAuthentication, null)
  if (token === null) {
    throw new Error('login.required');
  }
  return `Bearer ${token}`;
}

function getServer() {
  const server = getConfiguration(KeySettingsServer, null)
  if (server === null) {
    throw new Error('settings.server.missing');
  }
  return server;
}



function getHeaders() {
  return {
    "Content-Type": "application/json",
    Authorization: getAuth()
  };
}

async function getVersion() {
  const url = getServer() + "/api/v1/version";
  const res = await fetch(url);
  const data = await res.json();

  if (res.ok) {
    return data;
  }
  throw new Error("Failed to get learnalist server version information");
}

async function getListsByMe() {
  const url = getServer() + "/api/v1/alist/by/me";
  const res = await fetch(url, {
    headers: getHeaders()
  });

  let manyLists = await res.json();
  if (res.ok) {
    return manyLists;
  }
  throw new Error("Failed to get lists by me");
}


async function putList(aList) {
  const response = {
    status: 400,
    data: {}
  }

  const url = getServer() + '/api/v1/alist/' + aList.uuid;
  const res = await fetch(url, {
    method: 'PUT',
    headers: getHeaders(),
    body: JSON.stringify(aList)
  });

  const data = await res.json();
  switch (res.status) {
    case 200:
    case 403:
    case 400:
      response.status = res.status
      response.data = data
      return response;
      break;
  }
  throw new Error('Unexpected response from the server');
}


// Look at https://github.com/freshteapot/learnalist-api/blob/master/docs/api.user.login.md
async function postLogin(username, password) {
  const response = {
    status: 400,
    body: {}
  }

  const input = {
    username: username,
    password: password
  }

  const url = getServer() + "/api/v1/user/login";
  const res = await fetch(url, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
    },
    body: JSON.stringify(input)
  });

  const data = await res.json();
  switch (res.status) {
    case 200:
    case 403:
    case 400:
      response.status = res.status
      response.body = data
      return response;
      break;
  }
  throw new Error('Unexpected response from the server');
}


// postList title: string listType: string
async function postList(title, listType) {
  const input = {
    data: [],
    info: {
      title: title,
      type: listType,
      labels: []
    }
  }

  const url = getServer() + '/api/v1/alist';
  const res = await fetch(url, {
    method: "POST",
    headers: getHeaders(),
    body: JSON.stringify(input)
  });

  const data = await res.json();
  // TODO double check we handle the codes we send from the server.
  if (res.status === 400 || res.status === 201) {
    return {
      status: res.status,
      body: data
    };
  }
  throw new Error('Unexpected response from the server when posting a list');
}

// deleteList uuid: string
async function deleteList(uuid) {
  const url = getServer() + "/api/v1/alist/" + uuid;
  const res = await fetch(url, {
    method: "DELETE",
    headers: getHeaders()
  });

  const data = await res.json();
  // TODO double check we handle the codes we send from the server.
  if (res.status === 400 || res.status === 200 || res.status === 404) {
    return {
      status: res.status,
      body: data
    };
  }
  throw new Error('Unexpected response from the server when deleting a list');
}


export {
  getServer,
  getHeaders,
  getListsByMe,
  getVersion,
  postLogin,
  putList,
  postList,
  deleteList,
};
