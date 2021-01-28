# postgres-storage

### Routes

```
POST upload/:tenantID
```

```
POST files/:tenantID
req {page:1, rowsPerPage:10}
res {success:true, data:[], total:0, error:null}
```

```
GET file/:id
```

```
DELETE file/:id
```

### TODO

[x] pagination support
[ ] sort by field
[ ] file size limit env var
[ ] dockerize
[ ] env vars
