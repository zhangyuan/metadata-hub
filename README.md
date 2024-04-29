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
