# Markdown Architecture Decision Records

This is mostly built on: https://github.com/adr/madr

To start a MADR see: [000-template.md](./decisions/000-template.md).

## Front matter

Newer MADRs start with YAML front matter (`title`, `status`, `date`, `tags`, `summary`, `related`) — see [000-template.md](./decisions/000-template.md).
It exists so the MADR set can be scanned as a map without opening every file. When both front matter and the `* Status:` bullet are present, front matter wins.

`outdated` marks a MADR whose decision still holds but whose content mentions things a release removed, with one line per release saying what no longer matches the code.
It is independent of `status`: an `accepted` MADR can be `outdated`. Older MADRs may carry front matter with only this field.

## Listing MADRs

Use the `list.sh` script to list and filter MADRs. Output is `<file> [<status>] <summary> #tag #tag`, with `!outdated` appended when the MADR has an `outdated` field:

```bash
./docs/madr/list.sh                          # list all MADRs
./docs/madr/list.sh | grep '\[accepted\]'    # list only accepted MADRs
./docs/madr/list.sh | grep -v '\[accepted\]' # list only not accepted MADRs
./docs/madr/list.sh | grep '#zone-egress'    # list MADRs tagged zone-egress
./docs/madr/list.sh | grep '!outdated'       # list MADRs partly outdated by a release
```
