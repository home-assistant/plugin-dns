# fallback

## Name

Plugin *Fallback* is able to selectively forward the query to another upstream server, depending the error result provided by the initial resolver

## Description

The *fallback* plugin allows an fallback set of upstreams be specified which will be used
if the plugin chain returns specific error messages. The *fallback* plugin utilizes the *forward* plugin (<https://coredns.io/plugins/forward>) to query the specified upstreams.

> The *fallback* plugin supports only DNS protocol and random policy w/o additional *forward* parameters, so following directives will fail:

```
. {
    forward . 8.8.8.8
    fallback NXDOMAIN . tls://192.168.1.1:853 {
        policy sequential
    }
}
```

As the name suggests, the purpose of the *fallback* is to allow a fallback when, for example,
the desired upstreams became unavailable.

## Syntax

```
{
    fallback [original] RCODE_1[,RCODE_2,RCODE_3...] . DNS_RESOLVERS
}
```

* **original** is optional flag. If it is set then fallback uses original request instead of potentially changed by other plugins
* **RCODE** is the string representation of the error response code. The complete list of valid rcode strings are defined as `RcodeToString` in <https://github.com/miekg/dns/blob/master/msg.go>, examples of which are `SERVFAIL`, `NXDOMAIN` and `REFUSED`. At least one rcode is required, but multiple rcodes may be specified, delimited by commas.
* **DNS_RESOLVERS** accepts dns resolvers list.

## Examples

### Fallback to local DNS server

The following specifies that all requests are forwarded to 8.8.8.8. If the response is `NXDOMAIN`, *fallback* will forward the request to 192.168.1.1:53, and reply to client accordingly.

```
. {
	forward . 8.8.8.8
	fallback NXDOMAIN . 192.168.1.1:53
	log
}

```
### Fallback with original request used

The following specify that `original` query will be forwarded to 192.168.1.1:53 if 8.8.8.8 response is `NXDOMAIN`. `original` means no changes from next plugins on request. With no `original` flag fallback will forward request with EDNS0 option (set by rewrite).

```
. {
	forward . 8.8.8.8
	rewrite edns0 local set 0xffee 0x61626364
	fallback original NXDOMAIN . 192.168.1.1:53
	log
}

```

### Multiple alternates

Multiple alternates can be specified, as long as they serve unique error responses.

```
. {
    forward . 8.8.8.8
    fallback NXDOMAIN . 192.168.1.1:53
    fallback original SERVFAIL,REFUSED . 192.168.100.1:53
    log
}

```
