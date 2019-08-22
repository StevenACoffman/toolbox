## What The Issue: Find the Jira issue from the topic

This utility will pull out the summary and description for a given JIRA ticket aka "topic".

### Usage
`JIRA_API_URI` environment variable is optional

```
$ export JIRA_LOGIN=login
$ export JIRA_PASSWORD=password
$ export JIRA_BASE_URL=https://jira.example.com
$ export JIRA_API_URI=/rest/api/2/issue/
$ wti CORE-5339
CORE-5339 - Logbuffer should recover automatically from kafka transients

When kafka servers restart we often see log messages begin to queue on disk. The logbuffer process is not able to sufficiently handle kafka interruptions. Recovery currently requires manual intervention to restart affected app instances. This action should be automated.

SC:
# Logbuffer process is automatically restarted if message queuing exceeds a threshold
```

If you pass the optional `-resolves` flag, it will inject a link back to the jira issue:
```
$ wti CORE-5339 -resolves
CORE-5339 - Logbuffer should recover automatically from kafka transients

Resolves [CORE-5339|https://jira.jstor.org/browse/CORE-5339]

When kafka servers restart we often see log messages begin to queue on disk. The logbuffer process is not able to sufficiently handle kafka interruptions. Recovery currently requires manual intervention to restart affected app instances. This action should be automated.

SC:
# Logbuffer process is automatically restarted if message queuing exceeds a threshold
``` 
