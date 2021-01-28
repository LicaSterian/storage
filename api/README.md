API

### Routes

```
POST /file/upload
form-file named document
```

```
POST /files
req {
  "id": 1, // optional
  "page": 1,
  "perPage": 10,
  "filterFields": ["name"],
  "filterValues": [["$like", "foo"]],
  "sortBy": "created_at",
  "sortAsc": false,
  "fields": []
}
res {
  "request_id": 1, // only if passed in request
  "success": true,
  "data": {
    "rows": []
  },
  "error": "" // only if there was an error
}
```

```
GET file/:id
download the file
```

```
DELETE file/:id
delete the file from /files/uuid path and DB
```

### TODO

[x] pagination support
[z] sort by field
[ ] file size limit env var
[ ] encrypt local files
