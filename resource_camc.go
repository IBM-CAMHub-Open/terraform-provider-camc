//
// Copyright : IBM Corporation 2016, 2016
//

package main

import (
	"bytes"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"github.com/hashicorp/terraform/helper/schema"
	"io/ioutil"
	"log"
	"net/http"
)

func resourceCAMC() *schema.Resource {
	return &schema.Resource{
		Create: resourceCAMCCreate,
		Read:   resourceCAMCRead,
		Update: resourceCAMCUpdate,
		Delete: resourceCAMCDelete,

		Schema: map[string]*schema.Schema{
			"name": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},

			"url": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},

			"delete_url": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				Default:  "",
				ForceNew: true,
			},

			"update_url": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				Default:  "",
				ForceNew: true,
			},

			"read_url": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				Default:  "",
				ForceNew: true,
			},

			"method": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				Default:  "GET",
				ForceNew: true,
			},

			"payload": &schema.Schema{
				Type:      schema.TypeString,
				Optional:  true,
				StateFunc: jsonStateFunc,
				Default:   "null",
				ForceNew:  true,
			},

			"username": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				Default:  "",
				ForceNew: true,
			},

			"password": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				Default:  "",
				ForceNew: true,
			},

			"skip_ssl_verify": &schema.Schema{
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
				ForceNew: true,
			},

			"cert_file": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				Default:  "",
				ForceNew: true,
			},

			"key_file": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				Default:  "",
				ForceNew: true,
			},

			"ca_file": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				Default:  "",
				ForceNew: true,
			},

			"trace": &schema.Schema{
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
				ForceNew: true,
			},
		},
	}
}

func traceMessage(d *schema.ResourceData, msg string) {
	if d.Get("trace").(bool) == true {
		log.SetFlags(0)
		log.Print(msg)
	}
}

func traceMessagef(d *schema.ResourceData, fmt string, msg string) {
	if d.Get("trace").(bool) == true {
		log.SetFlags(0)
		log.Printf(fmt, msg)
	}
}

func makeCreateRequest(d *schema.ResourceData, m interface{}, url string) (string, error) {
	//get all possible inputs
	method := d.Get("method").(string)
	payload := d.Get("payload").(string)
	username := d.Get("username").(string)
	password := d.Get("password").(string)
	certFile := d.Get("cert_file").(string)
	keyFile := d.Get("key_file").(string)
	caFile := d.Get("ca_file").(string)
	skip_ssl_verify := d.Get("skip_ssl_verify").(bool)

	//initialize a tlsConfig structure
	tlsConfig := &tls.Config{
		InsecureSkipVerify: skip_ssl_verify,
	}

	//if client cert information provided use it to setup the tlsConfig structure
	if certFile != "" && keyFile != "" && caFile != "" {
		traceMessage(d, "**********  start using client cert connectivity")
		cert, err := tls.LoadX509KeyPair(certFile, keyFile)
		if err != nil {
			log.Fatal(err)
		}

		// Load CA cert
		caCert, err := ioutil.ReadFile(caFile)
		if err != nil {
			log.Fatal(err)
		}
		caCertPool := x509.NewCertPool()
		caCertPool.AppendCertsFromPEM(caCert)

		// Setup HTTPS client
		tlsConfig.Certificates = []tls.Certificate{cert}
		tlsConfig.RootCAs = caCertPool
		tlsConfig.BuildNameToCertificate()
	} else {
		traceMessage(d, "**********  skip using client cert connectivity, certfile, keyfile and caFile not passed in as args")
	}

	//initialize the http transport
	tr := &http.Transport{
		TLSClientConfig: tlsConfig,
	}

	//initialize the http client with the defined transport
	client := &http.Client{
		Transport: tr,
	}

	//process the input payload
	var x map[string]interface{}

	if payload != "null" {
		json.Unmarshal([]byte(payload), &x)
	}

	b := new(bytes.Buffer)
	if len(x) != 0 {
		json.NewEncoder(b).Encode(x)
	}

	//setup the http request
	req, _ := http.NewRequest(method, url, b)
	req.Close = true
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Accept", "application/json")

	if username != "" && password != "" {
		req.SetBasicAuth(username, password)
	}

	//make the call
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}

	//access the response body
	defer resp.Body.Close()
	respBody, _ := ioutil.ReadAll(resp.Body)

	rb := string(respBody)

	// don't save the payload, may have secrets
	//traceMessage(d, "[trace] setting payload to 'null' so not stored in state file")
	//d.Set("payload", "null")

	if (resp.StatusCode == 200 || resp.StatusCode == 201) {
		traceMessage(d, fmt.Sprintf("**********  good response StatusCode --> %v, body --> %s", resp.StatusCode, rb))
		return rb, nil
	} else {
		//return all errors
		traceMessage(d, fmt.Sprintf("**********  bad response StatusCode --> %v, body --> %s", resp.StatusCode, rb))
		err := fmt.Errorf("\n%s", rb)
		return "", err
	}
}

func makeRequest(d *schema.ResourceData, m interface{}, url string) (string, error) {
	//get all possible inputs
	name := d.Get("name").(string)
	method := d.Get("method").(string)
	username := d.Get("username").(string)
	password := d.Get("password").(string)
	certFile := d.Get("cert_file").(string)
	keyFile := d.Get("key_file").(string)
	caFile := d.Get("ca_file").(string)
	skip_ssl_verify := d.Get("skip_ssl_verify").(bool)

	//initialize a tlsConfig structure
	tlsConfig := &tls.Config{
		InsecureSkipVerify: skip_ssl_verify,
	}

	//if client cert information provided use it to setup the tlsConfig structure
	if certFile != "" && keyFile != "" && caFile != "" {
		traceMessage(d, "**********  start using client cert connectivity")
		cert, err := tls.LoadX509KeyPair(certFile, keyFile)
		if err != nil {
			log.Fatal(err)
		}

		// Load CA cert
		caCert, err := ioutil.ReadFile(caFile)
		if err != nil {
			log.Fatal(err)
		}
		caCertPool := x509.NewCertPool()
		caCertPool.AppendCertsFromPEM(caCert)

		// Setup HTTPS client
		tlsConfig.Certificates = []tls.Certificate{cert}
		tlsConfig.RootCAs = caCertPool
		tlsConfig.BuildNameToCertificate()
	} else {
		traceMessage(d, "**********  skip using client cert connectivity, certfile, keyfile and caFile not passed in as args")
	}

	//initialize the http transport
	tr := &http.Transport{
		TLSClientConfig: tlsConfig,
	}

	//initialize the http client with the defined transport
	client := &http.Client{
		Transport: tr,
	}

	traceMessagef(d, "**********  using name attribute --> %s", name)
	resource := fmt.Sprintf("{\"resourceID\":\"%s\"}", name)

	b := new(bytes.Buffer)
	if len(resource) != 0 {
		json.NewEncoder(b).Encode(resource)
	}

	//setup the http request
	req, _ := http.NewRequest(method, url, b)
	req.Close = true
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Accept", "application/json")

	if username != "" && password != "" {
		req.SetBasicAuth(username, password)
	}

	//make the call
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}

	//access the response body
	defer resp.Body.Close()
	respBody, _ := ioutil.ReadAll(resp.Body)

	rb := string(respBody)

	// don't save the payload, may have secrets
	//d.Set("payload", "null")

	if (resp.StatusCode == 200 || resp.StatusCode == 201) {
		traceMessage(d, fmt.Sprintf("**********  good response StatusCode --> %v, body --> %s", resp.StatusCode, rb))
		return rb, nil
	} else {
		//return all errors
		traceMessage(d, fmt.Sprintf("**********  bad response StatusCode --> %v, body --> %s", resp.StatusCode, rb))
		err := fmt.Errorf("\n%s", rb)
		return "", err
	}
}

func resourceCAMCCreate(d *schema.ResourceData, m interface{}) error {
	//get create url
	url := d.Get("url").(string)
	if url != "" {
		_, err := makeCreateRequest(d, m, url)
		if err == nil {
			b := make([]byte, 16)
			_, err := rand.Read(b)
			if err != nil {
				fmt.Println("Error:", err)
				return nil
			}
			uuid := fmt.Sprintf("%X-%X-%X-%X-%X", b[0:4], b[4:6], b[6:8], b[8:10], b[10:])
			d.SetId(uuid)
			return nil
		} else {
			return err
		}
	} else {
		return nil
	}
}

func resourceCAMCRead(d *schema.ResourceData, m interface{}) error {
	//get read url
	url := d.Get("read_url").(string)
	if url != "" {
		_, err := makeRequest(d, m, url)
		if err == nil {
			return nil
		} else {
			return err
		}
	} else {
		return nil
	}
}

func resourceCAMCUpdate(d *schema.ResourceData, m interface{}) error {
	//get update url
	url := d.Get("update_url").(string)
	if url != "" {
		_, err := makeRequest(d, m, url)
		if err == nil {
			return nil
		} else {
			return err
		}
	} else {
		return nil
	}
}

func resourceCAMCDelete(d *schema.ResourceData, m interface{}) error {
	//get delete url
	url := d.Get("delete_url").(string)
	if url != "" {
		_, err := makeRequest(d, m, url)
		if err == nil {
			d.SetId("")
			return nil
		} else {
			return err
		}
	} else {
		return nil
	}
}
