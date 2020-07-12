const keys = {
  'lists.by.me': 'my.lists',
  'my.edited.lists': 'my.edited.lists',
  'last.screen': 'last.screen',
  'authentication.bearer': 'settings.authentication',
  'user.uuid': 'user.uuid',
  'settings.server': 'settings.server',
  'settings.install.defaults': 'settings.install.defaults',
}

function get(key, defaultResult) {
  if (!localStorage.hasOwnProperty(key)) {
    return defaultResult;
  }

  return JSON.parse(localStorage.getItem(key))
}

function save(key, data) {
  localStorage.setItem(key, JSON.stringify(data));
}

function rm(key) {
  localStorage.removeItem(key);
}

// TODO remove this
function clear() {
  localStorage.clear();
  save(keys['settings.install.defaults'], true);
  const apiServer = document.querySelector('meta[name="api.server"]');
  if (apiServer) {
    save(keys['settings.server'], apiServer.content);
  } else {
    save(keys['settings.server'], 'https://learnalist.net');
  }
  // TODO why is this not showing up?
  save(keys['my.edited.lists'], []);
  save(keys['lists.by.me'], []);
}

export default {
  get,
  save,
  rm,
  clear,
  keys
};
