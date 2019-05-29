# Api

| Method | Uri | Description | Status Codes |
| --- | --- | --- | --- |
| GET | /v1/ | Replies with a simple message. | 200 |
| GET | /v1/version | Version informationn about the server. | 200 |
| POST | /v1/alist | Save a list. | 201, 400, 403 |
| DELETE | /v1/alist/{uuid} | Delete a list via uuid. | 200, 404, 403, 500 |
| PUT | /v1/alist/{uuid} | Update all fields allowed to a list. | 200, 400, 403 |
| GET | /v1/alist/{uuid} | Get a list via uuid. | 200, 404, 403 |
| GET | /v1/alist/by/me(?labels=,list_type={v1,v2}) | Get lists by the currently logged in user. | 200 |
| POST | /v1/labels | Save a new label. | 200, 201, 400 |
| GET | /v1/labels/by/me | Get labels by the currently logged in user. | 200, 500 |
| DELETE | /v1/labels/{uuid} | Delete a label via uuid. | 200, 500 |
| POST | /v1/share/alist | Share a list with another user. | 200, 404, 400 |
