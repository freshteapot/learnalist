# Api

| Method | Uri | Description | Status Codes |
| --- | --- | --- | --- |
| post | /alist | add a new list | 201,400,500 |
| get | /alist/by/me | Get lists by me | 200 |
| delete | /alist/{uuid} | Delete a list | 200,403,404,500 |
| get | /alist/{uuid} | Get a list | 200,403,404,500 |
| put | /alist/{uuid} | Update a list | 200,403,404,500 |
| post | /spaced-repetition/ | Add entry for spaced based learning | 200,201,400,500 |
| get | /spaced-repetition/all | Get all entries for spaced repetition learning | 200,500 |
| get | /spaced-repetition/next | Get next entry for spaced based learning | 200,204,404,500 |
| post | /spaced-repetition/viewed | Update spaced entry with feedback from the user | 200,404,500 |
| delete | /spaced-repetition/{uuid} | Deletes a single entry based on the UUID | 204,400,500 |
| post | /user/login | Login with username and password. The token can be used in future api requests via bearerAuth | 200,400,403,500 |
| post | /user/register | Register a new user with username and password | 200,201,400,500 |
| delete | /user/{uuid} | Deletes a user and there lists | 200,403,500 |
| get | /version | Get information about the server, linked to the git repo | 200 |

# Auto generated via
```
make generate-docs-api-overview
```
