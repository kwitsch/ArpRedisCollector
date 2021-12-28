# ArpRedisCollector

Periodically fetches arp entries and stores them in a redis database.  
All settings are configured through environment variables.

---

## Settings

!!! CAP_NET_ADMIN is needed for ARP requests !!!
<br><br>
### Required

* ARC_REDIS_ADDRESS = redis server address
* ARC_ARP_SUBNET_1 = Network filter  
<br>
### Overview

| Name                | Type    | Default | Description |
|---------------------|----------|-------|--------------|
| ARC_VERBOSE         | bool     | False | Enables extensive logging |
| ARC_REDIS_ADDRESS   | string   |       | Redis connection ip:port |
| ARC_REDIS_USERNAME  | string   |       | Redis username |
| ARC_REDIS_PASSWORD  | string   |       | Redis password |
| ARC_REDIS_DATABASE  | int      | 0     | Redis database |
| ARC_REDIS_ATTEMPTS  | int      | 3     | Redis connection attempts |
| ARC_REDIS_COOLDOWN  | duration | 1s    | Coooldown between redis connection attempts |
| ARC_ARP_INTERVALL   | duration | 5m    | Pause between ARP poll |
| ARC_ARP_TIMEOUT     | duration | 200ms | Timeout for ARP requests |
| ARC_ARP_STATICTABLE | bool     | False | Disables redis TTLs |
| ARC_ARP_SUBNET_n    | string   |       | CIDR list as network filter (n = number of list entry)|
<br>
---

## Docker stack example

```yaml
version: "3.8"

services:
  test1:
    image: ghcr.io/kwitsch/arprediscollector
    environment:
      - ARC_REDIS_ADDRESS=192.168.0.100:6379
      - ARC_ARP_SUBNET_1=192.168.0.0/24
    cap_add:
      - CAP_NET_ADMIN
    networks:
      - host

networks:
  host:
    name: host
    external: true

```
