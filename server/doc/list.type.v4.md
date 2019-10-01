# ContentAndUrl (v4)
A list to link some text with a url.

The data within, is an object with the attributes "content" and "url".

```json
{
  "content": "Design is the art of arranging code to work today, and be changeable forever. – Sandi Metz",
  "url":  "https://dave.cheney.net/paste/clear-is-better-than-clever.pdf"
}
```

To create a list of type "v4", set type in the info object payload.

# Full example
```json
{
  "info": {
      "title": "A list of fine quotes.",
      "type": "v4"
  },
  "data": [
    {
      "content": "Im tough, Im ambitious, and I know exactly what I want. If that makes me a bitch, okay. ― Madonna",
      "url":  "https://www.goodreads.com/quotes/54377-i-m-tough-i-m-ambitious-and-i-know-exactly-what-i"
    },
    {
      "content": "Design is the art of arranging code to work today, and be changeable forever. – Sandi Metz",
      "url":  "https://dave.cheney.net/paste/clear-is-better-than-clever.pdf"
    }
  ]
}
```

# Post it
This is an example of from and to, "from" English months "to" Norwegian.
```sh
curl -XPOST 'http://localhost:1234/api/v1/alist' -u'iamchris:test123' -d'
{
  "info": {
      "title": "A list of fine quotes.",
      "type": "v4"
  },
  "data": [
    {
      "content": "Im tough, Im ambitious, and I know exactly what I want. If that makes me a bitch, okay. ― Madonna",
      "url":  "https://www.goodreads.com/quotes/54377-i-m-tough-i-m-ambitious-and-i-know-exactly-what-i"
    },
    {
      "content": "Design is the art of arranging code to work today, and be changeable forever. – Sandi Metz",
      "url":  "https://dave.cheney.net/paste/clear-is-better-than-clever.pdf"
    }
  ]
}
'
```

# Get your lists by filtering on ContentAndUrl / v4
We add pretty to the query string, to return the json a little easier to read.
```sh
curl 'http://localhost:1234/api/v1/alist/by/me?list_type=v4&pretty'  -u'iamchris:test123'
```
or
```sh
curl 'http://localhost:1234/api/v1/alist/by/me?list_type=v4'  -u'iamchris:test123'
```
