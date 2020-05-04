import { wrap } from 'svelte-spa-router'
import Home from './routes/home.svelte'
import Login from './routes/login.svelte'
import Logout from './routes/logout.svelte'
import NotFound from './routes/not_found.svelte'
import CreateList from './routes/create_list.svelte'
import CreateLabel from './routes/create_label.svelte'
import Create from './routes/create.svelte'
import ListFind from './routes/list_find.svelte'
import ListView from './routes/list_view.svelte'
import ListEdit from './routes/list_edit.svelte'
import ListDeleted from './routes/list_deleted.svelte'
import SettingsServerInformation from './routes/settings_server_information.svelte'
import { loginHelper } from './lib/helper.js';

// Outside of svelte, auto subscribing doesnt work.
let lh;
const unsubscribe = loginHelper.subscribe(value => {
  lh = value;
});

function logout() {
  loginHelper.logout();
  return true;
}

function checkIfLoggedIn(detail) {
  if (!lh.loggedIn) {
    loginHelper.redirectURLAfterLogin(detail.location);
    return false;
  }
  return true;
}

let routes = {
  '/': Home,
  '/login': Login,
  '/logout': wrap(
    Logout,
    logout),
  '/create': wrap(
    Create,
    checkIfLoggedIn),
  '/create/list': wrap(
    CreateList,
    checkIfLoggedIn),
  '/create/label': wrap(
    CreateLabel,
    checkIfLoggedIn),
  '/list/edit/:uuid': wrap(
    ListEdit,
    checkIfLoggedIn),
  '/list/view/:uuid': wrap(
    ListView,
    checkIfLoggedIn),
  '/list/deleted': wrap(
    ListDeleted,
    checkIfLoggedIn),
  '/lists/by/me': wrap(
    ListFind,
    checkIfLoggedIn),
  '/settings/server_information': SettingsServerInformation,
  // Catch-all, must be last
  '*': NotFound,
}
export default routes
