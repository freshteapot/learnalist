# Api

| Method | Uri | Description | Status Codes |
| --- | --- | --- | --- |
| GET | /version | Get information about the server, linked to the git repo | 200 |
| POST | /alist | add a new list | 201,400,422,500 |
| DELETE | /alist/{uuid} | Delete a list | 200,403,404,500 |
| PUT | /alist/{uuid} | Update a list | 200,403,404,422,500 |
| GET | /alist/{uuid} | Get a list | 200,403,404,500 |
| GET | /alist/by/me | Get lists by me | 200 |
| PUT | /app-settings/remind_v1 | Enable or disable push notifications for spaced repetition in remindV1 | 200,422,500 |
| POST | /assets/upload | Upload asset and link it to the user logged in | 201,400,500 |
| GET | /assets/{uuid} |  | 200 |
| DELETE | /assets/{uuid} | Deletes a single asset based on the UUID | 204,400,403,404,500 |
| PUT | /assets/share | Set asset for public or private access | 200,400,403,500 |
| GET | /challenge/{uuid} | Get all challenge info, users and records | 200,403,404,500 |
| PUT | /challenge/{uuid}/join | Join a challenge | 200,400,404,500 |
| PUT | /challenge/{uuid}/leave | Leave a challenge | 200,400,403,404,500 |
| GET | /challenges/{userUUID} | Get all challenges for a given user | 200,403,500 |
| POST | /challenge/ | Create a new challenge | 201,422,500 |
| DELETE | /challenge/{uuid} | Delete a challenge, forever | 200,403,404,500 |
| POST | /mobile/register-device | Register the user and the token, to be able to send push notifications | 200,422,500 |
| GET | /plank/history | Get all planks for a given user | 200,500 |
| DELETE | /plank/{uuid} | Delete a single entry based on the UUID | 204,400,404,500 |
| POST | /plank/ | Add plank stats | 200,201,500 |
| GET | /remind/daily/{app_identifier} |  | 200,404,422 |
| DELETE | /remind/daily/{app_identifier} |  | 200,404,500 |
| PUT | /remind/daily/ | Set remind settings for app_identifier | 200,422,500 |
| POST | /spaced-repetition/viewed | Update spaced entry with feedback from the user | 200,404,422,500 |
| DELETE | /spaced-repetition/overtime | Remove list from dripfeed. | 200,403,500 |
| POST | /spaced-repetition/overtime | Add for dripfeed (Slowly add this list for spaced repetition learning). | 200,403,404,422,500 |
| GET | /spaced-repetition/overtime/active/{uuid} | Ugly light url to check if list active for this user. | 200,404 |
| DELETE | /spaced-repetition/{uuid} | Deletes a single entry based on the UUID | 204,400,404,500 |
| GET | /spaced-repetition/next | Get next entry for spaced based learning | 200,204,404,500 |
| GET | /spaced-repetition/all | Get all entries for spaced repetition learning | 200,500 |
| POST | /spaced-repetition/ | Add entry for spaced based learning | 200,201,422,500 |
| DELETE | /user/{uuid} | Deletes a user and there lists | 200,403,500 |
| POST | /user/register | Register a new user with username and password | 200,201,400,500 |
| POST | /user/login/idp | Login with idToken, mostly to support mobile devices. | 200,400,403,422,500 |
| POST | /user/login | Login with username and password. The token can be used in future api requests via bearerAuth | 200,400,403,500 |

# Auto generated via
```
make generate-docs-api-overview
```
