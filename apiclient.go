package common

import (
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"
)

const DEFAULT_IDLE_CONN_TIMEOUT = 5
const DEFAULT_DISABLE_KEEPALIVES = false

type APIClient interface {
	Close()
	Get(path string, response interface{}) (string, error)
	Post(path string, request, response interface{}, headers *map[string]string) (string, error)
	Put(path string, request, response interface{}, headers *map[string]string) (string, error)
	Delete(path string, response interface{}) (string, error)
}

type client struct {
	c       *http.Client
	URL     string
	Headers map[string]string
	verbose bool
	debug   bool
}

func NewAPIClient(prefix, url, certFile, keyFile, caFile string, headers *map[string]string) (APIClient, error) {

	api := client{
		URL:     url,
		Headers: make(map[string]string),
		verbose: ViperGetBool(prefix + "verbose"),
		debug:   ViperGetBool(prefix + "debug"),
	}

	if headers != nil {
		for k, v := range *headers {
			api.Headers[k] = v
		}
	}

	ViperSetDefault(prefix+"api_client.idle_conn_timeout", DEFAULT_IDLE_CONN_TIMEOUT)
	ViperSetDefault(prefix+"api_client.disable_keepalives", DEFAULT_DISABLE_KEEPALIVES)

	transport := http.Transport{
		IdleConnTimeout:   time.Duration(ViperGetInt64(prefix+"api_client.idle_conn_timeout")) * time.Second,
		DisableKeepAlives: ViperGetBool(prefix + "api_client.disable_keepalives"),
	}

	if certFile != "" || keyFile != "" || caFile != "" {
		tlsConfig := tls.Config{}
		if certFile == "" || keyFile == "" || caFile == "" {
			return nil, Fatalf("incomplete TLS config: cert=%s key=%s ca=%s\n", certFile, keyFile, caFile)
		}

		if api.debug {
			log.Printf("cert: %s\n", certFile)
			log.Printf("key: %s\n", keyFile)
		}
		cert, err := tls.LoadX509KeyPair(os.ExpandEnv(certFile), os.ExpandEnv(keyFile))
		if err != nil {
			return nil, Fatalf("error loading client certificate pair: %v", err)
		}

		if api.debug {
			log.Printf("CA: %s\n", caFile)
		}
		caCert, err := ioutil.ReadFile(os.ExpandEnv(caFile))
		if err != nil {
			return nil, Fatalf("error loading certificate authority file: %v", err)
		}

		caCertPool, err := x509.SystemCertPool()
		if err != nil {
			return nil, Fatalf("error opening system certificate pool: %v", err)
		}
		caCertPool.AppendCertsFromPEM(caCert)
		tlsConfig.Certificates = []tls.Certificate{cert}
		tlsConfig.RootCAs = caCertPool
		transport.TLSClientConfig = &tlsConfig
	}

	api.c = &http.Client{Transport: &transport}

	return &api, nil
}

func (c *client) Close() {
	c.c.CloseIdleConnections()
	c.c = nil
}

func (c *client) Get(path string, response interface{}) (string, error) {
	return c.request("GET", path, nil, response, nil)
}

func (c *client) Post(path string, request, response interface{}, headers *map[string]string) (string, error) {
	return c.request("POST", path, request, response, headers)
}

func (c *client) Put(path string, request, response interface{}, headers *map[string]string) (string, error) {
	return c.request("PUT", path, request, response, headers)
}

func (c *client) Delete(path string, response interface{}) (string, error) {
	return c.request("DELETE", path, nil, response, nil)
}

func (c *client) request(method, path string, requestData, responseData interface{}, headers *map[string]string) (string, error) {
	var requestBytes []byte
	var err error
	switch requestData.(type) {
	case nil:
	case *[]byte:
		requestBytes = *(requestData.(*[]byte))
	default:
		requestBytes, err = json.Marshal(requestData)
		if err != nil {
			return "", Fatalf("failed marshalling JSON body for %s request: %v", method, err)
		}
	}

	request, err := http.NewRequest(method, c.URL+path, bytes.NewBuffer(requestBytes))
	if err != nil {
		return "", Fatalf("failed creating %s request: %v", method, err)
	}

	// add the headers set up at instance init
	for key, value := range c.Headers {
		request.Header.Add(key, value)
	}

	if headers != nil {
		// add the headers passed in to this request
		for key, value := range *headers {
			request.Header.Add(key, value)
		}
	}

	if c.verbose {
		log.Printf("<-- %s %s (%d bytes)", method, c.URL+path, len(requestBytes))
		if c.debug {
			log.Println("BEGIN-REQUEST-HEADER")
			for key, value := range request.Header {
				log.Printf("%s: %s\n", key, value)
			}
			log.Println("END-REQUEST-HEADER")
			log.Println("BEGIN-REQUEST-BODY")
			log.Println(string(requestBytes))
			log.Println("END-REQUEST-BODY")
		}
	}

	response, err := c.c.Do(request)
	if err != nil {
		return "", Fatalf("request failed: %v", err)
	}
	defer response.Body.Close()
	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return "", Fatalf("failure reading response body: %v", err)
	}
	if c.verbose {
		log.Printf("--> '%s' (%d bytes)\n", response.Status, len(body))
		if c.debug {
			log.Println("BEGIN-RESPONSE-BODY")
			log.Println(string(body))
			log.Println("END-RESPONSE-BODY")
		}
	}

	var text string
	if len(body) > 0 {
		err = json.Unmarshal(body, responseData)
		if err != nil {
			return "", Fatalf("failed decoding JSON response: %v", err)
		}
		t, err := json.MarshalIndent(responseData, "", "  ")
		if err != nil {
			return "", Fatal(err)
		}
		text = string(t)
	}

	if response.StatusCode < 200 || response.StatusCode >= 300 {
		var detail string
		if len(body) > 0 {
			detail = "\n" + string(body)
		}
		return "", Fatalf("%s%s", response.Status, detail)
	}

	return text, nil
}
