---
title: "Login"
type: "user"
url: /login.html
---
<user-menu></user-menu>
<user-login></user-login>
<script>
	const el = document.querySelector('user-login');
	el.redirectOnLogin = "/welcome.html";
</script>
