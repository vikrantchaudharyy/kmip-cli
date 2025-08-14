package main

import (
	"bufio"
	"crypto/tls"
	"crypto/x509"
	"encoding/hex"
	"encoding/xml"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"strings"

	"github.com/gemalto/kmip-go"
	"github.com/gemalto/kmip-go/kmip14"
	ttlv2 "github.com/gemalto/kmip-go/ttlv"
)

func main() {
	// Define flags
	serverAddr := flag.String("server", "", "KMIP server address and port (e.g., localhost:5696) (required)")
	keyFile := flag.String("key", "", "Client private key file in PEM format (required)")
	certFile := flag.String("cert", "", "Client certificate file in PEM format (required)")
	caFile := flag.String("cacert", "", "CA certificate file in PEM format for server verification (optional)")
	inputFile := flag.String("in", "", "Input file for the KMIP request. If not specified, reads from stdin (optional)")
	inputFormat := flag.String("input-format", "hex", "The format of the input request. Can be 'hex' or 'xml' (optional)")
	outputFormat := flag.String("output-format", "hex", "The format for printing the response. Can be 'hex' or 'xml' (optional)")
	help := flag.Bool("help", false, "Show help message (optional)")

	flag.Usage = printCustomHelp
	flag.Parse()

	if *help || *serverAddr == "" || *keyFile == "" || *certFile == "" {
		printCustomHelp()
		return
	}

	// setup logger
	log.SetFlags(0)

	// Create client connection
	conn, err := setupConnection(*serverAddr, *keyFile, *certFile, *caFile)
	if err != nil {
		log.Fatalf("Failed to establish connection: %v", err)
	}
	defer conn.Close()

	// Read request
	requestBytes, err := readRequest(*inputFile, *inputFormat)
	if err != nil {
		log.Fatalf("Failed to read request: %v", err)
	}

	// Send request and get response
	responseTTLV, _, err := sendRequest(conn, requestBytes)
	if err != nil {
		log.Fatalf("Failed to send request: %v", err)
	}

	// Print response
	if err := printResponse(responseTTLV, *outputFormat); err != nil {
		log.Fatalf("Failed to print response: %v", err)
	}
}

func setupConnection(serverAddr, keyFile, certFile, caFile string) (*tls.Conn, error) {
	cer, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		return nil, fmt.Errorf("failed to load client key pair: %w", err)
	}

	conf := &tls.Config{
		Certificates: []tls.Certificate{cer},
		MinVersion:   tls.VersionTLS12,
	}

	if caFile != "" {
		caCert, err := ioutil.ReadFile(caFile)
		if err != nil {
			return nil, fmt.Errorf("failed to read CA certificate: %w", err)
		}
		caCertPool := x509.NewCertPool()
		caCertPool.AppendCertsFromPEM(caCert)
		conf.RootCAs = caCertPool
	} else {
		// for dev/testing only, if no CA is provided
		conf.InsecureSkipVerify = true
	}

	conn, err := tls.Dial("tcp", serverAddr, conf)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to server: %w", err)
	}
	return conn, nil
}

func readRequest(inputFile, inputFormat string) ([]byte, error) {
	var reader io.Reader
	if inputFile != "" {
		f, err := os.Open(inputFile)
		if err != nil {
			return nil, fmt.Errorf("failed to open input file %q: %w", inputFile, err)
		}
		defer f.Close()
		reader = f
	} else {
		fmt.Println("Enter KMIP request, then press Ctrl+D (or Ctrl+Z on Windows) to send:")
		reader = os.Stdin
	}

	requestBytes, err := ioutil.ReadAll(reader)
	if err != nil {
		return nil, fmt.Errorf("failed to read request: %w", err)
	}

	requestStr := strings.TrimSpace(string(requestBytes))

	if inputFormat == "xml" {
		hexRequest, err := xmlToHex(strings.NewReader(requestStr))
		if err != nil {
			return nil, fmt.Errorf("failed to convert XML request to hex: %w", err)
		}
		return hex.DecodeString(hexRequest)
	}

	return hex.DecodeString(requestStr)
}

func sendRequest(conn io.ReadWriter, request []byte) (ttlv2.TTLV, string, error) {
	_, err := conn.Write(request)
	if err != nil {
		return nil, "Unable to write request", err
	}

	decoder := ttlv2.NewDecoder(bufio.NewReader(conn))
	resp, err := decoder.NextTTLV()
	if err != nil {
		return nil, "Unable to decode response TTLV", err
	}

	var respMsg kmip.ResponseMessage
	err = decoder.DecodeValue(&respMsg, resp)
	if err != nil {
		// still return the raw response even if decoding fails
		return resp, "", fmt.Errorf("failed to decode response message: %w", err)
	}

	// // Check header result status
	// if respMsg.ResponseHeader.ResultStatus != kmip14.ResultStatusSuccess {
	// 	return resp, "", fmt.Errorf("KMIP batch operation failed: %s - %s",
	// 		respMsg.ResponseHeader.ResultStatus.String(),
	// 		respMsg.ResponseHeader.ResultMessage)
	// }

	// Check each batch item's result status
	for i, item := range respMsg.BatchItem {
		if item.ResultStatus != kmip14.ResultStatusSuccess {
			return resp, "", fmt.Errorf("KMIP operation in batch item %d failed: %s - %s",
				i,
				item.ResultStatus.String(),
				item.ResultMessage)
		}
	}

	hexResponse := fmt.Sprintf("%x", []byte(resp))
	return resp, hexResponse, nil
}

func printResponse(response ttlv2.TTLV, format string) error {
	switch format {
	case "xml":
		s, err := xml.MarshalIndent(response, "", "  ")
		if err != nil {
			return fmt.Errorf("error printing XML: %w", err)
		}
		fmt.Println(string(s))
	case "hex":
		fmt.Println(hex.EncodeToString(response))
	default:
		return fmt.Errorf("unknown output format: %s", format)
	}
	return nil
}

func xmlToHex(r io.Reader) (string, error) {
	var raw ttlv2.TTLV
	decoder := xml.NewDecoder(r)
	if err := decoder.Decode(&raw); err != nil {
		if err == io.EOF {
			return "", nil
		}
		return "", err
	}
	return hex.EncodeToString(raw), nil
}

// Print custom help message
func printCustomHelp() {
	fmt.Println("A command-line tool for sending KMIP requests to a server.")
	fmt.Println("\nUsage:")
	fmt.Println("  kmip-cli -server <ip:port> -key <keyfile> -cert <certfile> [options]")
	fmt.Println("\nOptions:")
	flag.PrintDefaults()
	fmt.Println("\nExample:")
	fmt.Println(`  echo "42007801..." | kmip-cli -server localhost:5696 -key client.key -cert client.pem`)
}
