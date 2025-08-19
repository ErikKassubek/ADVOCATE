| | Syncting | nsq | terraform | caddy | zinx |
| --- | --- | --- | --- | --- | --- |
| TP  | 1 | 0 | 1 | 0 | 0 |
| Select in endless for loop | 2 | 2 | 0 | 0 | 0 |
| Main routine terminates while routine still running, but should continue to run if main routine runs longer | 1 | 0 | 0 | 0 | 2 |
| Select on context with cancel in defer | 2 | 0 | 1 | 0 | 0 |
| Range on channel without close | 0 | 0 | 1 | 0 | 0 |
| Timeout | 5 | 7 | 9 | 4 | 0 |
| Note | |  Alle TO auf dem selben Test | TP: read on nil channel, timeout durch 2 tests | All timeouts have the same cause |  |
