
## taskd

```
go install ./...
echo 'export TASKD_DB="~/.taskd.db"' >> ~/.bashrc
```

```bash
taskd new <label> <description>
taskd done <id>
taskd close <id>
taskd open
taskd list
taskd old
taskd details <id>
```
