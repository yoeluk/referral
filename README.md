## Referral plugin for CoreDNS

The referral plugin uses the `forward` plugin, and therefore it should be linked before the `forward` plugin. This 
plugin acts pretty much like the forward plugin by inspect the response from upstream and when there are no answers 
but an authoritive is returned it attempts to resolve the query against the authority provided in the first response.
It doesn't recurse. It does this one-level referral lookup only.

## Configuration

A typical configuration 

```go
. {
    health :8080
	referral
    forward . 10.100.100.2:53
    debug
    log
}
```