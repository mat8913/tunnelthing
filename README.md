# Tunnelthing
Tunnel a TCP connection through a Syncthing relay.

## How to build
Make sure you have a go compiler installed (Tunnelthing is written in golang),
and run `./build.sh`. It will output binaries to `bin/`. You can then install by
copying those binaries to `/usr/local/bin/`, or any other directory on your
`$PATH`.

## Example usage
This is an example of how to use Tunnelthing to tunnel an ssh server over a
Syncthing relay.

On the server, create a new directory and move to it. This is where the
server’s Tunnelthing certificate and private key will be stored. Then, run
`tt-gencert` to generate the certificate and private key.

```
$ mkdir ~/tunnelthing_workdir
$ cd ~/tunnelthing_workdir
$ tt-gencert
```

The certificate and private key only needs to be generated once. Re-use them to
keep the same device ID.

Then run `tt-listen` to start the Tunnelthing server. It will print its device
ID on startup

```
$ cd ~/tunnelthing_workdir
$ tt-listen tcp 127.0.0.1:22
2024/10/27 16:16:02 Server ID: QBZF2LI-W6DYH66-LFP4XPY-FQTUIT6-5SQZULY-XGN5UH5-4G5C4N3-HJXCGQC
```

Note that it can take around 5 minutes to find a relay, connect to it, and
announce its address to the discovery server. If you get a 404 Not Found error
when trying to connect, wait a few minutes for the server to announce its
address. The server will print a message like this when ready:

```
2024/10/27 16:21:07.364055 global.go:251: DEBUG: global@https://discovery-v6.syncthing.net/v2/ Announcement: {[relay://85.143.216.93:22067/?id=KCV3GDC-NYP3C7K-4Q35Y5S-VCERYFM-HQCX67E-UWXHPSB-3LXFXR4-6J2DOQO]}
```

On the client side, to connect to the ssh server, you will need to add an entry
to your ssh config to make it use `tt-connect`. For example (note: replace the
device ID with your own server's device ID):

```
Host tttest
ProxyCommand tt-connect QBZF2LI-W6DYH66-LFP4XPY-FQTUIT6-5SQZULY-XGN5UH5-4G5C4N3-HJXCGQC
ProxyUseFdpass yes
ServerAliveInterval 60
```

> [!NOTE]
> Some notes about this config.
>
> `ServerAliveInterval 60` is necessary to keep
> idle ssh sessions alive (Syncthing relays automatically disconnect
> connections where nothing is sent).
>
> `ProxyUseFdpass yes` is required due to the fact that `tt-connect` is
> designed to pass its connection to the ssh client instead of running in the
> background (see
> https://www.gabriel.urdhr.fr/2016/08/07/openssh-proxyusefdpass/ for more info
> on how that works).

Then use normal ssh commands to connect:

```
$ ssh tttest
2024/10/27 16:23:06 looking up QBZF2LI-W6DYH66-LFP4XPY-FQTUIT6-5SQZULY-XGN5UH5-4G5C4N3-HJXCGQC
2024/10/27 16:23:08 found QBZF2LI-W6DYH66-LFP4XPY-FQTUIT6-5SQZULY-XGN5UH5-4G5C4N3-HJXCGQC at relay://85.143.216.93:22067/?id=KCV3GDC-NYP3C7K-4Q35Y5S-VCERYFM-HQCX67E-UWXHPSB-3LXFXR4-6J2DOQO
2024/10/27 16:23:08 connecting to relay://85.143.216.93:22067/?id=KCV3GDC-NYP3C7K-4Q35Y5S-VCERYFM-HQCX67E-UWXHPSB-3LXFXR4-6J2DOQO : QBZF2LI-W6DYH66-LFP4XPY-FQTUIT6-5SQZULY-XGN5UH5-4G5C4N3-HJXCGQC
2024/10/27 16:23:09 joining session QBZF2LI-W6DYH66-LFP4XPY-FQTUIT6-5SQZULY-XGN5UH5-4G5C4N3-HJXCGQC@85.143.216.93:22067
2024/10/27 16:23:09 performing tls handshake for QBZF2LI-W6DYH66-LFP4XPY-FQTUIT6-5SQZULY-XGN5UH5-4G5C4N3-HJXCGQC@85.143.216.93:22067
$
```

## Security considerations
Note that the connection you choose to tunnel is essentially exposed to the
internet. Treat it as if you had opened a port in your firewall and take
necessary precautions.

Note that Tunnelthing does not provide encryption or authentication. It is
assumed that the protocol you are tunneling provides its own encryption and
authentication (like ssh for example).

Note that Tunnelthing may not be bug free. You may wish to use the systemd
service file in the `example/` directory or take other measures to limit the
potential damage in case a bug is found and exploited.
