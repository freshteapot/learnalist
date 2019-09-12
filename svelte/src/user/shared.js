import Cookies from './js.cookie.js'

const ID_LOGGED_IN_KEY = 'loggedin';

function register(username, password) {
  console.log('// TODO send a request to the server to register a user');
  console.log('// TODO save response and log them in.');
}

function logOut(redirect) {
  if (!redirect) {
    redirect = '/';
  }
  console.log('Clear cookies');
	Cookies.remove(ID_LOGGED_IN_KEY);
	window.location = redirect
}

function logIn(token, redirect) {
  Cookies.set(ID_LOGGED_IN_KEY, token);
	if (!redirect) {
		redirect = '/welcome.html';
	}
	window.location = redirect;
}

function loggedIn() {
	let item = Cookies.get(ID_LOGGED_IN_KEY);
	if (!item) {
		return false;
	}
	// TODO do I need to check whats in the cookie?
	return true;
}

export {loggedIn, logIn, logOut, register};
