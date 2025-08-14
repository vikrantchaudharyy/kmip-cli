# KMIP Client CLI

A command-line interface (CLI) tool for sending Key Management Interoperability Protocol (KMIP) requests to a KMIP server.

This tool allows you to send raw KMIP TTLV requests in hex format, or in a more human-readable XML format, which it will convert to TTLV before sending.

## Features

*   Connect to a KMIP server over TLS.
*   Send KMIP requests from a file or standard input.
*   Supports requests in hex-encoded TTLV format or XML format.
*   Display responses in hex-encoded TTLV or pretty-printed XML.

## Installation

```bash
go install github.com/vikrantchaudharyy/kmip-cli
```

## Usage

```
kmip-cli -server <ip:port> -key <keyfile> -cert <certfile> [options]
```

### Options

| Flag            | Description                                                                                                 | Default |
|-----------------|-------------------------------------------------------------------------------------------------------------|---------|
| `-server`       | KMIP server address and port (e.g., `localhost:5696`). (Required)                                           |         |
| `-key`          | Path to the client's private key file in PEM format. (Required)                                             |         |
| `-cert`         | Path to the client's certificate file in PEM format. (Required)                                             |         |
| `-cacert`       | Path to the CA certificate file in PEM format for server verification. (Optional)                           |         |
| `-in`           | Input file for the KMIP request. If not specified, reads from standard input. (Optional)                    | stdin   |
| `-input-format` | The format of the input request. Can be `hex` or `xml`. (Optional)                                          | `hex`   |
| `-output-format`| The format for printing the response. Can be `hex` or `xml`. (Optional)                                     | `hex`   |
| `-help`         | Show the help message.                                                                                      |         |

### Examples

**1. Send a hex-encoded TTLV request from stdin and get a hex response:**

```bash
echo "42007801..." | kmip-cli -server localhost:5696 -key client.key -cert client.pem
```

**2. Send an XML-formatted request from a file and get an XML response:**

```bash
kmip-cli -server localhost:5696 \
               -key client.key \
               -cert client.pem \
               -in request.xml \
               -input-format xml \
               -output-format xml
```

## For Developers

This tool uses the [gemalto/kmip-go](https://github.com/gemalto/kmip-go) library.

You can use the `ppkmip` tool from that library to convert between hex and XML formats for creating request files.

```bash
go get github.com/gemalto/kmip-go/cmd/ppkmip
```

```
