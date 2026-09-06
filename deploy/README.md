# Deploying on Swarm

Four VPS: three managers, one worker.

| Node    | Swarm role | Labels                | Runs                |
|---------|------------|-----------------------|---------------------|
| node-1  | manager    | `etcd=1`, `pgnode=1`  | etcd1, pg1, maybe the bot |
| node-2  | manager    | `etcd=2`, `pgnode=2`  | etcd2, pg2, maybe the bot |
| node-3  | manager    | `etcd=3`, `pgnode=3`  | etcd3, pg3, maybe the bot |
| node-4  | worker     | `pgnode=4`            | pg4, maybe the bot  |

Three stacks, deployed in order: `etcd` → `pg` → `app`.

- **The bot floats.** One replica, no placement constraints, so Swarm can
  restart it on any node when its current node goes down.
- **Postgres does not float.** Each member is pinned to one node by label. Its
  data is in a node-local volume, so moving the task would only produce an
  empty database. Availability comes from the other three members and Patroni
  promoting one of them, never from relocating a member.

## 1. Label the nodes

```sh
docker node update --label-add etcd=1 --label-add pgnode=1 node-1
docker node update --label-add etcd=2 --label-add pgnode=2 node-2
docker node update --label-add etcd=3 --label-add pgnode=3 node-3
docker node update --label-add pgnode=4 node-4
```

## 2. Create the overlay network

Two cases — pick the one that matches your setup.

### (a) Nodes talk over the public internet

```sh
docker network create \
  --driver overlay \
  --attachable \
  --opt encrypted \
  --subnet 10.20.0.0/16 \
  vpn-net
```

`--opt encrypted` is not optional here. Overlay traffic between separate VPS
crosses the public internet, and that traffic includes streaming replication
and database credentials. Swarm encrypts the *control* plane by default but
**not** the data plane.

### (b) Nodes are already on a VPN (WireGuard, Tailscale, Nebula, ...)

Drop `--opt encrypted`. The VPN already encrypts this traffic, and stacking
Swarm's IPsec on top of it costs throughput for no security you don't already
have. In exchange you have two things to get right.

**Set the MTU.** This is the one that bites.

```sh
ip link show wg0        # -> mtu 1420, say

docker network create \
  --driver overlay \
  --attachable \
  --opt com.docker.network.driver.mtu=1420 \
  --subnet 10.20.0.0/16 \
  vpn-net
```

Pass the **VPN interface's own MTU**, not a pre-adjusted number. Docker treats
this option as the underlay MTU and subtracts the 50 bytes of VXLAN overhead
itself — measured on Docker 29: `1420 -> 1370`, `1380 -> 1330`, `1280 -> 1230`,
always exactly −50. Subtracting by hand just wastes 50 bytes.

Left unset, Docker assumes a 1500-byte underlay and gives the overlay 1450,
which is ~80 bytes too large to fit inside WireGuard's 1420. The resulting
failure is genuinely nasty to diagnose: connections establish, small queries
return fine, and anything large silently stalls, because oversized frames are
black-holed rather than fragmented. The classic symptom in *this* stack is a
replica that clones to a certain point and hangs — `pg_basebackup` wedged
mid-transfer while Patroni reports "creating replica" forever.

Typical interface MTUs: WireGuard/`wg-quick` 1420, Nebula 1300, Tailscale 1280.
Read it off the interface rather than trusting the table.

**Point Swarm at the VPN.** A VPN the cluster does not actually route over buys
you nothing. Three per-node flags control this, and all three accept an
interface name as well as an IP — so pass `wg0` and every node gets an
identical command with no per-host IP bookkeeping:

| Flag | Controls |
|---|---|
| `--advertise-addr` | the address other nodes are told to reach this one on (control plane) |
| `--listen-addr` | which interface this node binds `2377` on (default `0.0.0.0`) |
| `--data-path-addr` | which interface carries VXLAN overlay traffic — i.e. **all replication traffic** |

Interface names are resolved to an address **when the command runs** — they
are not stored as names — and resolution fails if the interface carries more
than one address of a family (`interface lo has more than one IPv4 address`).
A single-stack `wg0` is fine; if yours is dual-stack, pass the literal IP.

On the first manager:

```sh
docker swarm init \
  --advertise-addr wg0 \
  --listen-addr wg0 \
  --data-path-addr wg0

docker swarm join-token -q manager   # note these down
docker swarm join-token -q worker
```

Then on each remaining node, pointing at the first manager's *WireGuard*
address:

```sh
docker swarm join \
  --token <manager-or-worker-token> \
  --advertise-addr wg0 \
  --listen-addr wg0 \
  --data-path-addr wg0 \
  10.8.0.1:2377
```

The mesh must be full: every node needs a route to every other node, not just
to a hub, since any member may stream from any other.

**Bring WireGuard up before dockerd.** If `wg0` does not exist when the daemon
starts it cannot bind the address it has on record, and Swarm comes up broken
after every reboot:

```ini
# /etc/systemd/system/docker.service.d/wireguard.conf
[Unit]
Requires=wg-quick@wg0.service
After=wg-quick@wg0.service
```

Because the address was resolved and persisted at init time, on reboot the
daemon tries to bind a literal `10.8.0.x` that exists only while the tunnel is
up. Without the ordering above, a boot that starts dockerd first fails with
`bind: cannot assign requested address`, and Swarm on that node does not come
up — its pinned Patroni member stays down with it. `docker.service` usually
carries `Restart=always`, so it may retry into a working state once WireGuard
appears, but it can equally exhaust its start limit and stay down; it is not
something to rely on.

The race is real rather than theoretical because the stock `wg-quick@.service`
is ordered `After=network-online.target` — the same target `docker.service`
waits on — so absent an explicit rule the two are free to start concurrently.
Check yours:

```sh
systemctl cat docker.service  | grep -E '^(After|Wants|Requires)='
systemctl cat wg-quick@wg0    | grep -E '^(After|Wants|Before)='
```

This matters most in exactly the situation you least want it: a provider-wide
power or maintenance event reboots several nodes at once, and the race fires on
all of them simultaneously. Lose two of three managers that way and the cluster
sits read-only until someone intervenes.

If WireGuard is configured through `systemd-networkd` (a `.netdev` + `.network`
pair) rather than `wg-quick`, the interface is created during network setup and
Docker's existing `After=network-online.target` is usually enough on its own.

Then `systemctl daemon-reload`. `Requires` means Docker refuses to start if the
tunnel fails — which is what you want when the tunnel is the only path between
nodes.

**On an existing cluster these cannot be changed in place.** `docker swarm
update` has no such flags; each node must leave and rejoin. Since the public
and WireGuard addresses coexist, this rolls without downtime: do the worker
first, then the managers one at a time — `docker node demote` → `docker swarm
leave` → `docker node rm` (from a surviving manager) → rejoin with the flags
above → `docker node promote` — never dropping below 2 of 3 managers.

If the cluster holds nothing precious yet, re-initialising is faster and less
error-prone. That does **not** destroy the database: the Postgres volumes are
node-local and survive `swarm leave`, so redeploying under the same stack name
(`pg`) reattaches `pg_pgN-data` untouched. Only the secrets need recreating.

**Verify it took:**

```sh
docker node inspect self --format '{{.Status.Addr}}'   # -> 10.8.0.x  (managers)
docker info --format '{{.Swarm.NodeAddr}}'             # -> 10.8.0.x  (any node)
sudo tcpdump -ni wg0 udp port 4789                     # overlay traffic on the tunnel
```

That last one is the real proof: if `4789/udp` shows up on `wg0` while a
replica is streaming, the data plane is on the VPN. If you see it on `eth0`
instead, `--data-path-addr` did not take.

Then close the Swarm ports on public interfaces — `2377/tcp`, `7946/tcp+udp`
and `4789/udp` reachable over `wg0` only, with just WireGuard's own
`51820/udp` exposed publicly.

### Either way: keep four ranges from colliding

The overlay subnet must not overlap your WireGuard range — and WireGuard is not
the only thing to check. Swarm creates an `ingress` network at init whether you
publish ports or not, and it defaults to **`10.0.0.0/24`**, which is exactly
where WireGuard meshes tend to live.

| Range | Used by | Where you set it |
|---|---|---|
| `10.8.0.0/24` | WireGuard mesh | your `wg0` config |
| `10.20.0.0/16` | the `vpn-net` overlay | `--subnet` on `network create` |
| `10.30.0.0/16` | Swarm's own networks, incl. `ingress` | `--default-addr-pool` at `swarm init` |
| `172.17.0.0/16`, `172.18.0.0/16` | `docker0` and `docker_gwbridge` | Linux defaults; `bip` in `daemon.json` |

Pin Swarm's auto-allocation pool at init, since `ingress` otherwise lands on
`10.0.0.0/24` on its own:

```sh
docker swarm init \
  --advertise-addr wg0 --listen-addr wg0 --data-path-addr wg0 \
  --default-addr-pool 10.30.0.0/16 \
  --default-addr-pool-mask-length 24
```

This is an init-time flag. Changing it afterwards means deleting and
recreating the `ingress` network, which interrupts every published port on the
cluster — so it is much cheaper to get right the first time.

**Why an overlap actually breaks things here.** The VXLAN tunnel endpoints
*are* the WireGuard addresses: overlay packets get encapsulated and sent to a
peer's `10.8.0.x`. Inside a container, the overlay subnet is claimed by `eth0`.
So if the two ranges overlap, traffic aimed at a WireGuard address matches the
container's own on-link route and is never handed to the tunnel at all.

For this project the sharp edge is `internal/executor/docker.go`, which dials
`tcp://<server>:<port>` to reach the managed VPN servers. If those servers sit
on the WireGuard mesh and the overlay overlaps it, the bot will talk to
Postgres perfectly well while every config-generation call fails — the dial
never leaves the container namespace. That asymmetry makes it a genuinely
confusing failure to chase.

Finally, `patroni/patroni.yml` scopes its `pg_hba` rules to the overlay subnet
rather than `0.0.0.0/0`. If you change the overlay range, change those lines to
match — and if it overlapped the WireGuard range, those rules would be
authorising a good deal more than intended.

## 3. Create the secrets

Hex passwords, so nothing needs URL-encoding inside the connection string:

```sh
PGAPP=$(openssl rand -hex 24)

printf '%s' "$(openssl rand -hex 24)" | docker secret create pg_superuser_password -
printf '%s' "$(openssl rand -hex 24)" | docker secret create pg_replication_password -
printf '%s' "$PGAPP"                  | docker secret create pg_app_password -

printf '%s' "postgres://vpn:${PGAPP}@pg1:5432,pg2:5432,pg3:5432,pg4:5432/vpn?target_session_attrs=read-write&connect_timeout=5&sslmode=disable" \
  | docker secret create database_url -

printf '%s' '<telegram-bot-token>' | docker secret create tg_token -
```

That multi-host DSN is the whole failover story on the client side. pgx tries
each host in turn on every new connection and `target_session_attrs=read-write`
makes it keep only the one accepting writes — so the bot finds the new leader
by itself, with no pgbouncer or HAProxy in the path. Nothing needs redeploying
after a failover.

## 4. Build and push the images

Both images must match your VPS architecture. On an Apple Silicon machine that
means building explicitly for amd64:

```sh
docker buildx build --platform linux/amd64 \
  -t ghcr.io/liminalvoid/patroni-pg17:latest --push deploy/patroni

docker buildx build --platform linux/amd64 \
  -t ghcr.io/liminalvoid/vpn-manager:latest --push .
```

## 5. Deploy

```sh
docker stack deploy -c deploy/etcd-stack.yml etcd
# wait until all three report healthy before continuing
docker service ls | grep etcd

docker stack deploy -c deploy/pg-stack.yml pg
# first boot: one member bootstraps, the other three clone from it
docker exec $(docker ps -qf name=pg_pg1) \
  patronictl -c /etc/patroni/patroni.yml list

docker stack deploy -c app-stack.yml app
```

Whichever member wins the initial race runs `post_init.sh`, which creates the
`vpn` role and database and loads `schema.sql`. The other three get all of it
through replication — do not run it on them by hand.

Expected steady state:

```
+ Cluster: vpn -------+---------+-----------+----+-----------+
| Member | Host | Role    | State     | TL | Lag in MB |
+--------+------+---------+-----------+----+-----------+
| pg1    | pg1  | Leader  | running   |  1 |           |
| pg2    | pg2  | Replica | streaming |  1 |         0 |
| pg3    | pg3  | Replica | streaming |  1 |         0 |
| pg4    | pg4  | Replica | streaming |  1 |         0 |
+--------+------+---------+-----------+----+-----------+
```

## What happens when a node dies

| Failure | Result |
|---|---|
| Node hosting the bot dies | Swarm marks the node down and starts the single replica elsewhere. Roughly 15–30s. |
| Node hosting a Postgres *replica* dies | Nothing. Three members remain; the member returns and catches up on its own when the node does. |
| Node hosting the Postgres *leader* dies | Patroni promotes a replica once the leader lock expires (`ttl: 30`). Measured at ~18s in testing. The bot's pool reconnects and lands on the new leader. |
| Two of the three managers die | Everything freezes. See below. |

**The two-manager limit is the real ceiling of a 4-node cluster.** Swarm raft
needs 2 of 3 managers for quorum, and so does etcd. Lose two and Swarm can no
longer schedule anything *and* Patroni, unable to renew the leader lock, will
demote the leader to read-only rather than risk split-brain. The database stays
readable, the bot stops being able to write, and nothing recovers until a
manager comes back. Adding a 5th node as a manager would let you survive two
losses; with four nodes this is the trade you are making.

## Operations

```sh
# cluster state
docker exec $(docker ps -qf name=pg_pg1) patronictl -c /etc/patroni/patroni.yml list

# planned failover, e.g. before rebooting the leader's node
docker exec -it $(docker ps -qf name=pg_pg1) patronictl -c /etc/patroni/patroni.yml switchover

# change Postgres settings cluster-wide (the bootstrap block is ignored after
# the first boot — the live config lives in etcd)
docker exec -it $(docker ps -qf name=pg_pg1) patronictl -c /etc/patroni/patroni.yml edit-config
```

Swarm configs are immutable. To edit `patroni.yml`, `bootstrap.sql` or
`schema.sql`, bump the `_v1` suffix in `pg-stack.yml`'s `configs:` block and
redeploy; Docker will not update a config in place.

## Two things this setup does not give you

**Replicas are not backups.** They are synchronous with your mistakes: a
`DROP TABLE` or a bad migration reaches all four members in milliseconds. If
the data matters, add WAL archiving to off-cluster storage (pgBackRest or
`wal-g`) — that is what gives you point-in-time recovery.

**Failover can lose the last few transactions.** Replication is asynchronous by
default, so a leader that dies mid-write may not have shipped its last commits.
For a bot recording payments that may matter. `patronictl edit-config` and set
`synchronous_mode: true` to trade a little write latency for not losing
acknowledged commits.
