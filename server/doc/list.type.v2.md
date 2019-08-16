# From To list (v2)

Data, provides an array of objects with attributes from and to.
```json
{
  "from": "January",
  "to": "Januar"
}
```

To create a list of type "v2", set type in the info object payload.

# Full example
```json
{
  "info": {
    "title": "Days of the Week",
    "type": "v2",
    "labels": []
  },
  "data":[
    {
      "from": "January",
      "to": "Januar"
    }
  ]
}
```

# Post it
This is an example of from and to, "from" English months "to" Norwegian.
```sh
curl -XPOST 'http://localhost:1234/api/v1/alist' -u'iamchris:test123' -d'
{
  "data": [
    {
      "from": "January",
      "to": "Januar"
    },
    {
      "from": "February",
      "to": "Februar"
    },
    {
      "from": "March",
      "to": "Mars"
    },
    {
      "from": "April",
      "to": "April"
    },
    {
      "from": "May",
      "to": "Mai"
    },
    {
      "from": "June",
      "to": "Juni"
    },
    {
      "from": "July",
      "to": "Juli"
    },
    {
      "from": "August",
      "to": "August"
    },
    {
      "from": "September",
      "to": "September"
    },
    {
      "from": "October",
      "to": "Oktober"
    },
    {
      "from": "November",
      "to": "November"
    },
    {
      "from": "December",
      "to": "Desember"
    }
  ],
  "info": {
    "title": "Months from English to Norwegian",
    "type": "v2",
    "labels": [
      "english",
      "norwegian"
    ]
  }
}
'
```

# Get your lists by filtering on FromToList / v2
We add pretty to the query string, to return the json a little easier to read.
```sh
curl 'http://localhost:1234/api/v1/alist/by/me?list_type=v2&pretty'  -u'iamchris:test123'
```
or
```sh
curl 'http://localhost:1234/api/v1/alist/by/me?list_type=v2'  -u'iamchris:test123'
```

# How this data was made:

```sh
paste -d"\t" dataset/months.en.txt dataset/months.no.txt > dataset/months.en.no.txt
cat dataset/months.en.no.txt | awk -F'\t' '{print "from:"$1"::to:"$2}' | go run integrations/convert/main.go -type=v2 | jq . > dataset/months.en.no.json
```
