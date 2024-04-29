# metadata-hub

## Usage

### Prepare the dataset schema file

Prepare the dataset schema file. e.g.

```yaml
name: foo
columns:
  name: foo
  type: interger
  comments: This is the comments
---
name: bar
columns:
  name: bar
  type: string
  comments: This is the comments
```

And put the schema file under a folder.

### Serve the API

```bash
metadata-hub serve -d the_path_to_the_schema_folder
```

The default port is `:8080`.

## API

| API        | Example                                                |
| ---------- | ------------------------------------------------------ |
| 搜索字段   | `/api/search-columns?q=alarm`                          |
| 搜索表     | `/api/search-tables?q=alarm`                           |
| 数据集列表 | `/api/datasets?q=alarm`                                |
| 数据表列表 | `/api/datasets/:datasetName/tables/:tableName?q=alarm` |
