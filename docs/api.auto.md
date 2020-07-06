# Api

| Method | Uri | Description | Status Codes |
| --- | --- | --- | --- |
| post | /spaced-repetition/ | Add entry for spaced based learning | 200,201,400,500 |
| get | /spaced-repetition/all | Get all entries for spaced repetition learning | 200,500 |
| get | /spaced-repetition/next | Get next entry for spaced based learning | 200,204,404,500 |
| post | /spaced-repetition/viewed | Update spaced entry with feedback from the user | 200,404,500 |
| delete | /spaced-repetition/{uuid} | Deletes a single entry based on the UUID | 204,400,500 |
| post | /user/register | Register a new user with username and password | 200,201,400,500 |
| delete | /user/{uuid} | Deletes a user and there lists | 200,403,500 |
| get | /version | Get information about the server, linked to the git repo | 200 |

# Auto generated via
```
make generate-docs-api-overview
```
