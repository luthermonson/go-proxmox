# Recorder seed contract

Cassettes under `tests/recorder/testdata/` are recorded against a freshly
installed nested Proxmox VE instance whose state is **deterministic by
contract**. This file is that contract: anything that can affect a
recorded byte must be pinned here, otherwise re-records produce noisy
diffs and tests drift.

The driver in `mage/record` provisions and seeds the nested host; this
document describes what it produces. If you change the recorded surface,
change the driver and bump every cassette.

## Identity

| Property | Value |
|---|---|
| Hostname | `recorder` |
| FQDN | `recorder.example.test` |
| Node name | `recorder` |
| Static IP | `192.0.2.10/24` (RFC 5737 TEST-NET-1) |
| Gateway | `192.0.2.1` |
| Synthetic public host (after scrub) | `pve.example.test:8006` |

The recorded cassettes never contain the real lab IP. The
`tests/recorder` BeforeSaveHook chain rewrites the actual nested-VM
host name (`192.0.2.10:8006`) to `pve.example.test:8006` and any other
IPv4 / MAC into the corresponding TEST-NET-1 / `BC:24:11` ranges before
the cassette hits disk.

## Authentication

| Field | Value |
|---|---|
| Realm | `pve` |
| User | `recorder@pve` |
| Token name | `cassettes` |
| Token value | redacted at scrub time; cassettes never contain it |

The user has the `Administrator` role on `/`. Token privilege separation
is **disabled** so the token has the same effective rights as the user.
This is fine inside a throwaway nested PVE.

## Storages

| Name | Type | Path |
|---|---|---|
| `local` | Directory | default Proxmox layout |
| `local-lvm` | LVM-thin | default Proxmox layout |

The installer handles both at install time; the seed does not create
additional storages.

## Networks

| Property | Value |
|---|---|
| Bridge | `vmbr0` |
| Bridge IP | `192.0.2.1/24` (NAT to outside is fine for record-time API access) |

## Pinned IDs

VM and CT IDs are deliberately well above the conventional 100-200
range so they cannot collide with anything an outer-PVE operator has
running, and so the cassettes' VMIDs are visually distinct from
production.

| Resource | ID | Name | State | Purpose |
|---|---|---|---|---|
| VM | `9001` | `tpl-ubuntu` | template, stopped | Cassettes for "VM as template" (issue #198 case) |
| VM | `9002` | `vm-running` | running | Cassettes for "VM with non-null PID, blockstats, etc." |
| VM | `9003` | `vm-stopped` | stopped, has config | Cassettes for "VM with all device types configured" |
| LXC | `9101` | `ct-running` | running | Cassettes for "LXC with mountpoints, devs" |
| LXC | `9102` | `ct-stopped` | stopped | Cassettes for "LXC plain config" |

`vm-stopped` (9003) is the wide-config sample. Its config sets indexed
devices specifically chosen to exercise the routing in
`VirtualMachineConfig.UnmarshalJSON` — including indices `>9` that
issue #211 surfaced:

| Field | Value |
|---|---|
| `scsihw` | `virtio-scsi-pci` (prefix-collision case) |
| `numa` | `1` (bare scalar, prefix-collision case) |
| `scsi0` | `local-lvm:0,size=8G` |
| `scsi30` | `local-lvm:0,size=8G` |
| `net0` | `virtio,bridge=vmbr0` |
| `net15` | `virtio,bridge=vmbr0` |
| `net31` | `virtio,bridge=vmbr0` |
| `unused255` | `local-lvm:vm-9003-unused-255` |
| `hostpci15` | `0000:0f:00.0` |
| `ipconfig20` | `ip=192.0.2.20/24` |
| `numa0` | `cpus=0-1,memory=512` |

The nested host doesn't actually need to *run* this VM — Proxmox accepts
config keys for hardware that doesn't exist on the host (the VM just
won't start). The recorded cassettes only need the config to come back
from the API.

`ct-running` (9101) similarly has indexed mountpoints / devs:

| Field | Value |
|---|---|
| `mp0` | `/srv/data,mp=/data` |
| `mp42` | `/srv/forty-two,mp=/forty-two` |
| `mp255` | `/srv/last,mp=/last` |
| `dev15` | `/dev/sdb15` |
| `unused100` | `local-lvm:subvol-9101-unused-100` |
| `net0` | `name=eth0,bridge=vmbr0` |
| `net20` | `name=eth20,bridge=vmbr0` |

## Tags

`vm-running` carries `production;webserver` so cassettes for the tag
parsing path have a non-trivial input.

## What the seed script does

The `mage record` pipeline embeds the seed as a first-boot step in the
auto-installer's `answer.toml`. The script runs once on first boot of
the nested PVE and:

1. Creates the `recorder@pve` user and the `cassettes` API token
2. Grants `Administrator` on `/` to that user
3. Creates VM `9001` (template), VM `9002` (running), VM `9003` (stopped, wide config)
4. Creates LXC `9101` (running), LXC `9102` (stopped)
5. Sets the indexed-device config on VM `9003` and LXC `9101` per the tables above
6. Tags VM `9002` with `production;webserver`

After the seed completes, the nested PVE is ready for `mage record:pveN`
to drive the cassette generation against `https://192.0.2.10:8006/api2/json`
using the recorder token.

## Versions covered

| PVE major | Cassette directory | Source |
|---|---|---|
| 9.x | `tests/recorder/testdata/pve9/` | latest stable PVE 9 ISO |
| 8.x | `tests/recorder/testdata/pve8/` | latest stable PVE 8 ISO |

PVE 7 is **not** covered by this pipeline — `proxmox-auto-install-assistant`
requires PVE 8.2 or newer to prepare the unattended ISO, and PVE 7 hit
EOL in July 2024. Existing `tests/mocks/pve7x/*.go` gock fixtures stay as
legacy until the next major bump of go-proxmox drops them.

## Refreshing cassettes

```bash
# One time, on the workstation:
export PROXMOX_URL=https://outer-pve.example.test:8006/api2/json
export PROXMOX_TOKENID=automation@pve!recorder
export PROXMOX_SECRET=...
export PROXMOX_NODE_NAME=outer-node
export PROXMOX_NODE_STORAGE=local
export PROXMOX_RECORDER_SSH_HOST=outer-pve.example.test
export PROXMOX_RECORDER_SSH_USER=root

# Refresh both versions:
mage record:all

# Or one at a time:
mage record:pve9
mage record:pve8
```

The driver is idempotent against partial failures: it always destroys
the nested VM on exit, and the upstream ISO download is cached in the
outer PVE's `local` storage so re-runs skip the download.
