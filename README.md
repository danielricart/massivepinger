Continuous Ping Exporter
---
This exporter continously send ICMP packets to specified targets and collect metrics about it. 

when diagnosing network trasient erorrs and sporadic sudden packet drops, low-frequency ICMPs easily hide those issues as one single packet lost, or small latency spike. 

When running at a higher frequency, like the following examples at 0.1s intervals, we see the evolution of the latency spike and its recovery. 

```
64 bytes from 192.168.1.3: icmp_seq=52394 ttl=64 time=6.731 ms
64 bytes from 192.168.1.3: icmp_seq=52395 ttl=64 time=7.060 ms
64 bytes from 192.168.1.3: icmp_seq=52396 ttl=64 time=7.114 ms
64 bytes from 192.168.1.3: icmp_seq=52397 ttl=64 time=5.497 ms
64 bytes from 192.168.1.3: icmp_seq=52398 ttl=64 time=6.465 ms
64 bytes from 192.168.1.3: icmp_seq=52399 ttl=64 time=7.001 ms
64 bytes from 192.168.1.3: icmp_seq=52400 ttl=64 time=5.778 ms
64 bytes from 192.168.1.3: icmp_seq=52401 ttl=64 time=6.787 ms
64 bytes from 192.168.1.3: icmp_seq=52402 ttl=64 time=13.102 ms
64 bytes from 192.168.1.3: icmp_seq=52403 ttl=64 time=5.777 ms
64 bytes from 192.168.1.3: icmp_seq=52404 ttl=64 time=7.378 ms
64 bytes from 192.168.1.3: icmp_seq=52405 ttl=64 time=6.840 ms
64 bytes from 192.168.1.3: icmp_seq=52406 ttl=64 time=29.658 ms
64 bytes from 192.168.1.3: icmp_seq=52407 ttl=64 time=51.513 ms
64 bytes from 192.168.1.3: icmp_seq=52408 ttl=64 time=6.637 ms
64 bytes from 192.168.1.3: icmp_seq=52409 ttl=64 time=6.702 ms
64 bytes from 192.168.1.3: icmp_seq=52410 ttl=64 time=5.689 ms
64 bytes from 192.168.1.3: icmp_seq=52411 ttl=64 time=1047.632 ms
64 bytes from 192.168.1.3: icmp_seq=52412 ttl=64 time=942.658 ms
64 bytes from 192.168.1.3: icmp_seq=52413 ttl=64 time=836.959 ms
64 bytes from 192.168.1.3: icmp_seq=52414 ttl=64 time=733.305 ms
```

Intermitently spiky, with a recovery for one second:
```
64 bytes from 192.168.1.3: icmp_seq=53484 ttl=64 time=375.051 ms
64 bytes from 192.168.1.3: icmp_seq=53485 ttl=64 time=271.246 ms
64 bytes from 192.168.1.3: icmp_seq=53486 ttl=64 time=166.617 ms
64 bytes from 192.168.1.3: icmp_seq=53487 ttl=64 time=64.078 ms
64 bytes from 192.168.1.3: icmp_seq=53488 ttl=64 time=6.309 ms
64 bytes from 192.168.1.3: icmp_seq=53489 ttl=64 time=5.647 ms
64 bytes from 192.168.1.3: icmp_seq=53490 ttl=64 time=7.126 ms
64 bytes from 192.168.1.3: icmp_seq=53491 ttl=64 time=6.892 ms
64 bytes from 192.168.1.3: icmp_seq=53492 ttl=64 time=7.557 ms
64 bytes from 192.168.1.3: icmp_seq=53493 ttl=64 time=7.360 ms
64 bytes from 192.168.1.3: icmp_seq=53494 ttl=64 time=7.079 ms
64 bytes from 192.168.1.3: icmp_seq=53495 ttl=64 time=6.277 ms
64 bytes from 192.168.1.3: icmp_seq=53496 ttl=64 time=7.276 ms
64 bytes from 192.168.1.3: icmp_seq=53497 ttl=64 time=5.900 ms
64 bytes from 192.168.1.3: icmp_seq=53498 ttl=64 time=138.667 ms
64 bytes from 192.168.1.3: icmp_seq=53499 ttl=64 time=87.273 ms
64 bytes from 192.168.1.3: icmp_seq=53500 ttl=64 time=571.991 ms
64 bytes from 192.168.1.3: icmp_seq=53501 ttl=64 time=467.219 ms
64 bytes from 192.168.1.3: icmp_seq=53502 ttl=64 time=392.340 ms
64 bytes from 192.168.1.3: icmp_seq=53503 ttl=64 time=288.836 ms
```

or even worse:
```
64 bytes from 192.168.1.3: icmp_seq=52448 ttl=64 time=550.193 ms
64 bytes from 192.168.1.3: icmp_seq=52449 ttl=64 time=447.231 ms
64 bytes from 192.168.1.3: icmp_seq=52450 ttl=64 time=2962.455 ms
64 bytes from 192.168.1.3: icmp_seq=52451 ttl=64 time=2857.421 ms
64 bytes from 192.168.1.3: icmp_seq=52452 ttl=64 time=2752.061 ms
64 bytes from 192.168.1.3: icmp_seq=52453 ttl=64 time=2648.516 ms
64 bytes from 192.168.1.3: icmp_seq=52454 ttl=64 time=2543.411 ms
64 bytes from 192.168.1.3: icmp_seq=52455 ttl=64 time=2438.741 ms
64 bytes from 192.168.1.3: icmp_seq=52456 ttl=64 time=2333.805 ms
64 bytes from 192.168.1.3: icmp_seq=52457 ttl=64 time=2228.549 ms
64 bytes from 192.168.1.3: icmp_seq=52458 ttl=64 time=2124.848 ms
64 bytes from 192.168.1.3: icmp_seq=52459 ttl=64 time=2021.359 ms
64 bytes from 192.168.1.3: icmp_seq=52460 ttl=64 time=1912.618 ms
64 bytes from 192.168.1.3: icmp_seq=52461 ttl=64 time=1809.720 ms
```

or completely dropping and with very high latency:
```
Request timeout for icmp_seq 52817
Request timeout for icmp_seq 52818
Request timeout for icmp_seq 52819
Request timeout for icmp_seq 52820
64 bytes from 192.168.1.3: icmp_seq=52737 ttl=64 time=8821.256 ms
64 bytes from 192.168.1.3: icmp_seq=52738 ttl=64 time=8847.471 ms
64 bytes from 192.168.1.3: icmp_seq=52739 ttl=64 time=8742.829 ms
64 bytes from 192.168.1.3: icmp_seq=52740 ttl=64 time=8637.343 ms
64 bytes from 192.168.1.3: icmp_seq=52741 ttl=64 time=8522.618 ms
64 bytes from 192.168.1.3: icmp_seq=52742 ttl=64 time=8801.734 ms
64 bytes from 192.168.1.3: icmp_seq=52743 ttl=64 time=8697.493 ms
Request timeout for icmp_seq 52828
Request timeout for icmp_seq 52829
64 bytes from 192.168.1.3: icmp_seq=52786 ttl=64 time=4656.306 ms
Request timeout for icmp_seq 52831
Request timeout for icmp_seq 52832
Request timeout for icmp_seq 52833
Request timeout for icmp_seq 52834
64 bytes from 192.168.1.3: icmp_seq=52787 ttl=64 time=5019.544 ms
64 bytes from 192.168.1.3: icmp_seq=52788 ttl=64 time=4924.896 ms
Request timeout for icmp_seq 52837
Request timeout for icmp_seq 52838
64 bytes from 192.168.1.3: icmp_seq=52789 ttl=64 time=5282.936 ms
64 bytes from 192.168.1.3: icmp_seq=52790 ttl=64 time=5226.743 ms
64 bytes from 192.168.1.3: icmp_seq=52791 ttl=64 time=5311.990 ms
Request timeout for icmp_seq 52842
Request timeout for icmp_seq 52843
64 bytes from 192.168.1.3: icmp_seq=52792 ttl=64 time=5466.005 ms
Request timeout for icmp_seq 52845
Request timeout for icmp_seq 52846
Request timeout for icmp_seq 52847
Request timeout for icmp_seq 52848
```

until a final recovery:
```
ping: sendto: No route to host
ping: sendto: No route to host
ping: sendto: No route to host
64 bytes from 192.168.1.3: icmp_seq=56762 ttl=64 time=14.552 ms
64 bytes from 192.168.1.3: icmp_seq=56763 ttl=64 time=56.441 ms
64 bytes from 192.168.1.3: icmp_seq=56764 ttl=64 time=7.548 ms
64 bytes from 192.168.1.3: icmp_seq=56765 ttl=64 time=24.078 ms
64 bytes from 192.168.1.3: icmp_seq=56766 ttl=64 time=173.136 ms
64 bytes from 192.168.1.3: icmp_seq=56767 ttl=64 time=77.860 ms
64 bytes from 192.168.1.3: icmp_seq=56768 ttl=64 time=10.407 ms
64 bytes from 192.168.1.3: icmp_seq=56769 ttl=64 time=7.153 ms
64 bytes from 192.168.1.3: icmp_seq=56770 ttl=64 time=7.035 ms
64 bytes from 192.168.1.3: icmp_seq=56771 ttl=64 time=12.843 ms
64 bytes from 192.168.1.3: icmp_seq=56772 ttl=64 time=13.468 ms
64 bytes from 192.168.1.3: icmp_seq=56773 ttl=64 time=8.430 ms
64 bytes from 192.168.1.3: icmp_seq=56774 ttl=64 time=7.581 ms
64 bytes from 192.168.1.3: icmp_seq=56775 ttl=64 time=11.032 ms
64 bytes from 192.168.1.3: icmp_seq=56776 ttl=64 time=74.007 ms
64 bytes from 192.168.1.3: icmp_seq=56777 ttl=64 time=5.864 ms
64 bytes from 192.168.1.3: icmp_seq=56778 ttl=64 time=44.798 ms
64 bytes from 192.168.1.3: icmp_seq=56779 ttl=64 time=8.116 ms

```

## configuration
this exporter supports multiple monitored destinations.. each is configured in the specified `config.yaml` file with the following syntax:
```yaml
- target: 192.168.1.1
  interval: 0.1s
  timeout: infinite 
- target: 10.7.2.14
  interval: 0.2s
  timeout: 1s
```

All values must be specified. Incomplete targets are ignored and logged as warnings. 

`timeout` can be `infinite` meaning the packet will be waited forever. 

## CLI Arguments
| argument                             | default                      | dscription                                                                        | 
|--------------------------------------|------------------------------|-----------------------------------------------------------------------------------|
| `--config /some/path/file_name.yaml` | `--config config.yaml`       | configuration file path                                                           | 
| `--port 1234`                        | `--port 9123`                | metrics port                                                                      |
| `--identifier myorigin`              | `--identifier massivepinger` | an stable ID for this running instance (or set of instances, there's no blocking) |

## metrics exposed
Given the high frequency of each collected datapoint (each ping generated) it's not viable to expose the latency for each ICMP packet processed. 

This exporter exposes histogram timeseries per each target to visualize the different possible latencies given the amount of data. 

It exposes the latest processed ping value at the time of scrape. 


### general and information metrics
| metric name           | labels                 | description                                                    | 
|-----------------------|------------------------|----------------------------------------------------------------|
| icmp_duration_seconds | `target`, `identifier` | latest ping for each labelled target at the moment of scraping |
| icmp_interval_seconds | `target`, `identifier`               | configured ICMP interval defined per target, in seconds        | 
| icmp_timeout_seconds  | `target`, `identifier`               | configured ICMP timeout defined per target, in seconds         | 

### histogram metrics
| metric name | labels | description | 
| --- | --- | --- |
| icmp_sent_count       | `target`, `identifier`  | total icmp packets sent. the count of events that have been observed. |
| icmp_received_count | `target`, `identifier` | total icmp packets received back. should be close to icmp_sent_count |
| icmp_sent_bucket | `target`, `identifier` `le="<upper inclusive bound>"` | cumulative counters for the observation buckets | 
| icmp_received_bucket | `target`, `identifier` `le="<upper inclusive bound>"` | cumulative counters for the observation buckets | 
| icmp_sent_sum | `target`, `identifier` | the total sum of all observed values | 
| icmp_received_sum | `target`, `identifier` | the total sum of all observed values | 


# Technical details
The behaviour is coded after the Mac/BSD ping man page sections:

## ICMP PACKET DETAILS
An IP header without options is 20 bytes. An ICMP ECHO_REQUEST packet contains an additional 8 bytes worth of ICMP header followed by an arbitrary amount of data. When a packetsize is given, this indicated the
size of this extra piece of data (the default is 56). Thus the amount of data received inside of an IP packet of type ICMP ECHO_REPLY will always be 8 bytes more than the requested data space (the ICMP header).

If the data space is at least eight bytes large, ping uses the first eight bytes of this space to include a timestamp which it uses in the computation of round trip times. If less than eight bytes of pad are
specified, no round trip times are given.

## DUPLICATE AND DAMAGED PACKETS
The ping utility will report duplicate and damaged packets. Duplicate packets should never occur when pinging a unicast address, and seem to be caused by inappropriate link-level retransmissions. Duplicates may
occur in many situations and are rarely (if ever) a good sign, although the presence of low levels of duplicates may not always be cause for alarm. Duplicates are expected when pinging a broadcast or multicast
address, since they are not really duplicates but replies from different hosts to the same request.

Damaged packets are obviously serious cause for alarm and often indicate broken hardware somewhere in the ping packet's path (in the network or in the hosts).

