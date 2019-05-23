# Api

| Method | Uri | Description |
| --- | --- | --- |
| GET | /v1/ | Replies with a simple message. |
| GET | /v1/version | Version informationn about the server. |
| POST | /v1/alist | Save a list. |
| DELETE | /v1/alist/{uuid} | Delete a list via uuid. |
| PUT | /v1/alist/{uuid} | Update all fields allowed to a list. |
| GET | /v1/alist/{uuid} | Get a list via uuid. |
| GET | /v1/alist/by/me(?labels=,list_type={v1,v2}) | Get lists by the currently logged in user. |
| POST | /v1/labels | Save a new label. |
| GET | /v1/labels/by/me | Get labels by the currently logged in user. |
| DELETE | /v1/labels/{uuid} | Delete a label via uuid. |
| POST | /v1/share/alist | Share a list with another user. |
