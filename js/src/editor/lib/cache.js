const keys = {
  'lists.by.me': 'my.lists',
  'my.edited.lists': 'my.edited.lists',
  'last.screen': 'last.screen',
  'authentication.bearer': 'auth.bearer',
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

function clear() {
  localStorage.clear();
  save(keys['settings.install.defaults'], true);
  save(keys['settings.server'], 'https://learnalist.net');
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
