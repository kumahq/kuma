# Markdown Architecture Decision Records

This is mostly built on: https://github.com/adr/madr

To start an MADR see: [000-template.md](./decisions/000-template.md).

## Listing MADRs

Use the `list.sh` script to list and filter MADRs:

```bash
./docs/madr/list.sh                    # list all MADRs
./docs/madr/list.sh | grep accepted    # list only accepted MADRs 
./docs/madr/list.sh | grep -v accepted # list only non accepted MADRs
```
