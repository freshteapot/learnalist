Learnalist - Education by one list at a time.

# Today
[vaporware](https://en.wikipedia.org/wiki/Vaporware), check [status.json](./status.json) for random updates.

# Tomorrow

A way to learn via "alist". Made by you, another human or something else.
It will be a service, which will consume the Learnalist API. Hosted via learnalist.net or privately.

# Api

| Method | Uri | Description |
| --- | --- | --- |
| POST | /alist | Save a list. |
| PATCH | /alist/{uuid} | Update one or more fields to the list. |
| PUT | /alist/{uuid} | Update all fields allowed to a list. |
| GET | /alist/{uuid} | Get a list via uuid. |
| GET | /alist/by/{uuid} | Get lists by {uuid}. Allow for both public, private lists. |
