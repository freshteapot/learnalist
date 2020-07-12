import { wrap } from 'svelte-spa-router'
import Home from './routes/home.svelte'
import NotFound from './routes/not_found.svelte'
import CreateList from './routes/create_list.svelte'
import CreateLabel from './routes/create_label.svelte'
import Create from './routes/create.svelte'
import ListFind from './routes/list_find.svelte'
import ListView from './routes/list_view.svelte'
import ListEdit from './routes/list_edit.svelte'
import ListDeleted from './routes/list_deleted.svelte'
import SettingsServerInformation from './routes/settings_server_information.svelte'
import { loginHelper } from '../utils/login_helper.js';
import { loggedIn } from "../store.js";

// Outside of svelte, auto subscribing doesnt work.
let lh;
const unsubscribe = loginHelper.subscribe(value => {
  lh = value;
});

function checkIfLoggedIn(detail) {
  loggedIn()
  console.log(loggedIn());
  console.log(lh.loggedIn);
  if (!lh.loggedIn) {
    loginHelper.redirectURLAfterLogin(detail.location);
    return false;
  }
  return true;
}

let routes = {
  '/': Home,
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
