# Questions and Answers

## Via the api

## How can I get pretty print of the response?
Add the query string "pretty"
```sh
curl 'http://localhost:1234/v1/alist/by/me?pretty'
```

## Getting the response "Please refer to the documentation on list type v3"?
Check your data for the following rules

- Distance should not be empty.
- Stroke per minute (spm) should be between the range 10 and 50.
- Time should not be empty.
- Time is not valid format (xx:xx.xx).
- When should be YYYY-MM-DD.
- Per 500 (p500) should not be empty.
- Per 500 (p500) is not valid format (xx:xx.xx).
