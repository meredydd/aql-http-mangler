aql-http-mangler
================
An HTTP transparent proxy that kills the connection rather than deliver any non-HTTP-200 status code.

(c) 2012 Meredydd Luff <meredydd@senatehouse.org>


Background
----------
AQL (www.aql.com) offer some very nice SMS and telecomms services, but their delivery mechanisms for mobile-originated SMS are, to put it politely, a little kooky. If delivering an MO-SMS message to your specified URL fails, that's it - it's gone forever. So instead, you can give a list of URLs to the AQL system, and it will try them all in sequence until one of them succeeds. If they all fail, your message is still toast - but at least this way you have a chance, and you can use different hosting providers for redundancy. The idea is that if at least one of your hosts is up and running, you can spool messages and retry them yourself.

However, the AQL delivery system interprets *any* HTTP response as a successful delivery. So if your database barfed and you returned an HTTP 500, tough luck - your message is toast.

This little hack works around that, by sitting between AQL and your app. If your app returns anything except an HTTP-200 ("OK") response, it cuts the connection dead. 

I've raised a ticket or two, and AQL have indicated a willingness to modify their delivery systems (eventually to support retrying failed requests, and in the meantime at least to interpret a non-200 HTTP response as a failure). In the meantime, you can use this hack to work around it.


Installation
------------

You need to have the Go language/runtime installed (http://www.golang.org).

The following commands will install the HTTP mangler on any system that uses `upstart`, and configure it to start at boot time:

```bash
$ go build http-mangler.go
$ sudo cp http-mangler /usr/local/sbin
$ sudo chmod 755 /usr/local/sbin/http-mangler
$ sudo cp http-mangler.conf /etc/init/
$ sudo service http-mangler start
```


How it works
------------

The `http-mangler` process listens for TCP connections on port 8080. When a new connection arrives, it makes a new connection to `localhost:80` (where it is normal to find a webserver).

It forwards the incoming HTTP request to port 80, modifying only the `Connection` header (it specifies `Connection: close`, so it doesn't have to deal with keep-alive HTTP connections).

When the HTTP server on port 80 responds, `http-mangler` checks the first line of the response. If it is an HTTP 200 response, it forwards the response to the client, and all is right with the world.


Example
-------

I am running an HTTP server on port 80. It returns 200 for `http://localhost/`, but 404 for `http://localhost/doesntexist`:

```bash
$ curl -D - http://localhost/
HTTP/1.1 200 OK
Date: Thu, 13 Dec 2012 16:01:31 GMT
Server: Apache/2.2.22 (Ubuntu)
[... rest of response omitted ...]

$ curl -D - http://localhost/doesntexist
HTTP/1.1 200 OK
Date: Thu, 13 Dec 2012 16:04:17 GMT
[... rest of response omitted ...]
```

But if I go through `http-mangler` on port 8080, the error response will simply be cut off, in a way that AQL's servers will report as an error:

```bash
meredydd@hera:~/programming/aql-http-mangler$ curl -D - http://localhost:8080/
HTTP/1.1 200 OK
Date: Thu, 13 Dec 2012 16:05:51 GMT
Server: Apache/2.2.22 (Ubuntu)
[... rest of response omitted ]

meredydd@hera:~/programming/aql-http-mangler$ curl -D - http://localhost:8080/doesntexist
curl: (52) Empty reply from server
```

Tada!


Licence
-------

This software is available under the simplified BSD licence. See the `COPYING` file for more information.
