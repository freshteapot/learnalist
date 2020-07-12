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

export {
  getServer,
  getHeaders,
  postList,
};
