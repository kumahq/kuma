# Markdown Architecture Decision Records

This is mostly built on: https://github.com/adr/madr

To start a MADR see: [000-template.md](./decisions/000-template.md).

## Front matter

Newer MADRs start with YAML front matter (`title`, `status`, `date`, `tags`, `summary`, `related`) — see [000-template.md](./decisions/000-template.md).
It exists so the MADR set can be scanned as a map without opening every file. When both front matter and the `* Status:` bullet are present, front matter wins.

## Listing MADRs

Use the `list.sh` script to list and filter MADRs. Output is `<file> [<status>] <summary> #tag #tag`:

```bash
./docs/madr/list.sh                      # list all MADRs
./docs/madr/list.sh | grep '\[accepted\]' # list only accepted MADRs
./docs/madr/list.sh | grep -v accepted   # list only not accepted MADRs
./docs/madr/list.sh | grep '#zone-egress' # list MADRs tagged zone-egress
```
