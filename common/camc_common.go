//
// Copyright : IBM Corporation 2016, 2016
//

package common

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"time"
	"bufio"

	"github.com/hashicorp/terraform/helper/schema"
	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
)

func TraceMessage(d *schema.ResourceData, msg string) string {
	//Escape percentage from Sprintf formatting
	msg= strings.Replace(msg, "%", "%%", -1)
	var s string = fmt.Sprintf("\n**********\n%s\n**********", msg)
	if d.Get("trace").(bool) == true {
		log.SetFlags(0)
		log.Print(s)
	}
	return s
}

func ErrorMessage(d *schema.ResourceData, msg string, err_msg string) string {
	//Escape percentage from Sprintf formatting
	msg= strings.Replace(msg, "%", "%%", -1)
	err_msg= strings.Replace(err_msg, "%", "%%", -1)
	var s string = fmt.Sprintf("\n**********\n%s\n%s\n**********", msg, err_msg)
	if d.Get("trace").(bool) == true {
		log.SetFlags(0)
		log.Print(s)
	}
	return s
}

// Like ErrorMessage but without the stars. Use when a parent method is going to use ErrorMessage to wrap your error
func SubErrorMessage(d *schema.ResourceData, msg string, err_msg string) string {
	//Escape percentage from Sprintf formatting
	msg = strings.Replace(msg, "%", "%%", -1)
	err_msg = strings.Replace(err_msg, "%", "%%", -1)
	var s string = fmt.Sprintf("\n%s\n%s\n", msg, err_msg)
	if d.Get("trace").(bool) == true {
		log.SetFlags(0)
		log.Print(s)
	}
	return s
}

func TraceMessagef(d *schema.ResourceData, fmt string, msg string) {
	if d.Get("trace").(bool) == true {
		log.SetFlags(0)
		log.Printf(fmt, msg)
	}
}

func GenUUID() string {
	b := make([]byte, 16)
	_, err := rand.Read(b)
	if err != nil {
		fmt.Println("Error:", err)
		return ""
	}
	uuid := fmt.Sprintf("%X-%X-%X-%X-%X", b[0:4], b[4:6], b[6:8], b[8:10], b[10:])
	return uuid
}

func MakeRequest(d *schema.ResourceData, m interface{}, method string) (string, error) {
	//get all possible inputs
	camc_endpoint := d.Get("camc_endpoint").(string)
	data := d.Get("data").(string)
	username := d.Get("username").(string)
	password := d.Get("password").(string)
	certFile := d.Get("cert_file").(string)
	keyFile := d.Get("key_file").(string)
	caFile := d.Get("ca_file").(string)
	skip_ssl_verify := d.Get("skip_ssl_verify").(bool)
	access_token := d.Get("access_token").(string)

	
		
	//initialize a tlsConfig structure
	tlsConfig := &tls.Config{
		InsecureSkipVerify: skip_ssl_verify,
	}

	//if client cert information provided use it to setup the tlsConfig structure
	if certFile != "" && keyFile != "" && caFile != "" {
		TraceMessage(d, "start using client cert connectivity")
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
	}
	
	//initialize the http transport
	//GO 1.9.x support only http. GO 1.10 supports https
	//We may need to move to GO 1.1.0
	tr := &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		TLSClientConfig: tlsConfig,
		Dial: (&net.Dialer{
			Timeout:   15 * time.Second,
			KeepAlive: 30 * time.Second,
		}).Dial,
	}
	
	//initialize the http client with the defined transport
	client := &http.Client{
		Transport: tr,
	}

	//process the input data
	var json_string map[string]interface{}

	if data != "null" {
		if nil != json.Unmarshal([]byte(data), &json_string) {
			err := fmt.Errorf(ErrorMessage(d, "data is not valid json", data))
			return "", err
		}

	}

	b := new(bytes.Buffer)
	if len(json_string) != 0 {
		json.NewEncoder(b).Encode(json_string)
	}

	//setup the http request
	req, _ := http.NewRequest(method, camc_endpoint, b)
	req.Close = true
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Accept", "application/json")
	if access_token != "" {
		req.Header.Add("Authorization", "Bearer "+access_token)
	} else {
		err := fmt.Errorf(ErrorMessage(d, "No access_token supplied, cannot connect to pattern manager", ""))
		return "", err
	}

	if username != "" && password != "" {
		req.SetBasicAuth(username, password)
	}

	//make the call
	resp, err := client.Do(req)
	if err != nil {
		//return all errors
		msgStr := []string{"Unable to connect to endpoint ",camc_endpoint}
		err := fmt.Errorf(ErrorMessage(d, strings.Join(msgStr, "") , err.Error()))
		return "", err
	}

	//access the response body
	defer resp.Body.Close()
	respBody, _ := ioutil.ReadAll(resp.Body)

	rb := string(respBody)

	// don't save the data, may have secrets
	//TraceMessage(d, "[Trace] setting data to 'null' so not stored in state file")
	//d.Set("data", "null")

	if resp.StatusCode == 200 || resp.StatusCode == 201 {
		TraceMessage(d, fmt.Sprintf("Good response:\nStatusCode:%v\nMessage:\n%s", resp.StatusCode, rb))
		return rb, nil
	} else {
		//return all errors
		err := fmt.Errorf(ErrorMessage(d, "Error: Response from pattern manager:", fmt.Sprintf("StatusCode:%v\nMessage:\n%s", resp.StatusCode, rb)))
		return "", err
	}
}

func CreateSSHConfig(d *schema.ResourceData, m interface{}) (*ssh.ClientConfig, error) {
	remoteUser := d.Get("remote_user").(string)
	remotePassword := d.Get("remote_password").(string)
	remoteKeyEnc := d.Get("remote_key").(string)

	if remoteUser == "" {
		return nil, fmt.Errorf(ErrorMessage(d, "Error: remote_user is required when specifying remote_host", ""))
	}
	config, err := createClientConfig(d, remoteUser, remotePassword, remoteKeyEnc)
	if err != nil {
		return nil, err
	}
	return config, nil
}

func CreateBastionConfig(d *schema.ResourceData) (*ssh.ClientConfig, error) {
	bastionUser := d.Get("bastion_user").(string)
	bastionPassword := d.Get("bastion_password").(string)
	bastionKeyEnc := d.Get("bastion_private_key").(string)
	if bastionUser == "" {
		return nil, fmt.Errorf(ErrorMessage(d, "Error: bastion_user is required when specifying bastion_host", ""))
	}
	config, err := createClientConfig(d, bastionUser, bastionPassword, bastionKeyEnc)
	if err != nil {
		return nil, err
	}
	return config, nil
}

func createClientConfig(d *schema.ResourceData, remoteUser string, remotePassword string, remoteKeyEnc string) (*ssh.ClientConfig, error) {
	var config *ssh.ClientConfig
	if remoteKeyEnc != "" {
		remoteKey, err := base64.StdEncoding.DecodeString(remoteKeyEnc)
		if err != nil {
			return nil, fmt.Errorf(ErrorMessage(d, "Error: error decoding private key. The private key must be base64 encoded", ""))
		}
		key, err := ssh.ParsePrivateKey([]byte(remoteKey))
		if err != nil {
			return nil, fmt.Errorf(ErrorMessage(d, "Error: error parsing private key:", err.Error()))
		}
		// Authentication
		config = &ssh.ClientConfig{
			User: remoteUser,
			Auth: []ssh.AuthMethod{
				ssh.PublicKeys(key),
			},
			HostKeyCallback: func(hostname string, remote net.Addr, key ssh.PublicKey) error {
				return nil
			},
		}
	} else if remotePassword != "" {
		config = &ssh.ClientConfig{
			User: remoteUser,
			Auth: []ssh.AuthMethod{
				ssh.Password(remotePassword),
			},
			HostKeyCallback: func(hostname string, remote net.Addr, key ssh.PublicKey) error {
				return nil
			},
		}
	} else {
		return nil, fmt.Errorf(ErrorMessage(d, "Error: one of user password or key is required when specifying remote_host or bastion_host", ""))
	}
	return config, nil
}

func BuildWgetCmd(d *schema.ResourceData, m interface{}, destination string) []string {
	source := d.Get("source").(string)
	sourceUser := d.Get("source_user").(string)
	sourcePass := d.Get("source_password").(string)
	sourceNoCheckCert := d.Get("source_no_check_cert").(bool)

	var wgetCommand []string
	if strings.HasPrefix(source, "http://") {
		wgetCommand = []string{"wget", "-O", destination, source}
	} else {
		wgetCommand = []string{"wget", "--user", sourceUser, "--password", sourcePass, "-O", destination, source}
		if sourceNoCheckCert {
			wgetCommand = append(wgetCommand, "--no-check-certificate")
		}
	}
	return wgetCommand
}

// Helper function to transfer files from the local file system to a remote file system
func TransferLocalToRemote(d *schema.ResourceData, m interface{}, localSource string) error {
	destination := d.Get("destination").(string)
	remoteHost := d.Get("remote_host").(string)
	var client *ssh.Client
	var source string
	var clienterr error
	var sshConn ssh.Conn
	var bastTohostConn net.Conn
	var bastionclient *ssh.Client
	if localSource == "" {
		source = d.Get("source").(string)
	} else {
		source = localSource
	}
	config, err := CreateSSHConfig(d, m)
	if err != nil {
		return err
	}
	bastionHost := d.Get("bastion_host").(string)
	if bastionHost != "" {
		client, bastionclient, bastTohostConn, sshConn, clienterr = getClientUsingBastionConn(d, config)
		if clienterr != nil {
			return fmt.Errorf(SubErrorMessage(d, "Error: error connecting to bastion host", clienterr.Error()))
		}
		defer client.Close()
		defer bastionclient.Close()
		defer bastTohostConn.Close()
		defer sshConn.Close()
	} else {
		client, clienterr = ssh.Dial("tcp", remoteHost+":22", config)
		if clienterr != nil {
			return fmt.Errorf(ErrorMessage(d, "Error: error connecting to remote host", fmt.Sprintf("%s: %s", remoteHost, clienterr)))
		}
	}
	// open an SFTP session over an existing ssh connection.
	sftp, err := sftp.NewClient(client)
	if err != nil {
		return fmt.Errorf(ErrorMessage(d, "Error: error creating sftp client to host", fmt.Sprintf("%s: %s", remoteHost, err)))
	}
	defer sftp.Close()

	// Open the source file
	srcFile, err := os.Open(source)
	if err != nil {
		return fmt.Errorf(ErrorMessage(d, "Error: error opening local script", err.Error()))
	}
	defer srcFile.Close()

	// Create the destination file
	dstFile, err := sftp.Create(destination)
	if err != nil {
		return fmt.Errorf(ErrorMessage(d, "Error: error creating destination file on remote host:", err.Error()))
	}
	defer dstFile.Close()

	buf := make([]byte, 1024)
	for {
		n, _ := srcFile.Read(buf)
		if n == 0 {
			break
		}
		dstFile.Write(buf[:n])
	}
	return nil
}

// Determines what commands are possible and downloads a file to a remote system.
func DownloadRemoteFile(d *schema.ResourceData, m interface{}) error {
	source := d.Get("source").(string)
	destination := d.Get("destination").(string)
	sourceUser := d.Get("source_user").(string)
	sourcePass := d.Get("source_password").(string)
	sourceNoCheckCert := d.Get("source_no_check_cert").(bool)

	// if the remote system has wget, use wget
	whichCmdWget := []string{"which", "wget"}
	config, err := CreateSSHConfig(d, m)
	if err != nil {
		return err
	}

	_, err = RemoteExec(d, whichCmdWget, nil, config)
	if err == nil {
		wgetCommand := BuildWgetCmd(d, m, destination)
		_, err := RemoteExec(d, wgetCommand, nil, config)
		return err
	}

	// if the remote system has curl, use curl
	whichCmdCurl := []string{"which", "curl"}
	_, err = RemoteExec(d, whichCmdCurl, nil, config)
	if err == nil {
		var curlCommand []string
		if strings.HasPrefix(source, "http://") {
			curlCommand = []string{"curl", "-o", destination, source}
		} else {
			curlCommand = []string{"curl", "-u", fmt.Sprintf("%s:%s", sourceUser, sourcePass), "-o", destination, source}
			if sourceNoCheckCert {
				curlCommand = append(curlCommand, "-k")
			}
		}
		_, err := RemoteExec(d, curlCommand, nil, config)
		return err
	}

	// if the remote system doesn't have either, then download locally and transfer
	splitSource := strings.Split(source, "/")
	localSource := splitSource[len(splitSource)-1]
	localSourceDir, err := ioutil.TempDir("/tmp", "")
	localSourceFile := fmt.Sprintf("%s/%s", localSourceDir, localSource)
	if err != nil {
		return fmt.Errorf(ErrorMessage(d, fmt.Sprintf("Error: could not create temporary directory for source file"), err.Error()))
	}
	wgetCommand := BuildWgetCmd(d, m, localSourceFile)
	_, err = LocalExec(d, wgetCommand, nil)
	if err != nil {
		return err
	}
	err = TransferLocalToRemote(d, m, localSourceFile)
	os.Remove(localSourceFile)
	return err
}

// Determines if, how, and where to transfer the source script
func HandleSourceAndDest(d *schema.ResourceData, m interface{}) error {
	source := d.Get("source").(string)
	destination := d.Get("destination").(string)
	remoteHost := d.Get("remote_host").(string)
	sourceUser := d.Get("source_user").(string)
	sourcePass := d.Get("source_password").(string)
	sourceNoCheckCert := d.Get("source_no_check_cert").(bool)

	if source != "" {
		if destination == "" {
			return fmt.Errorf(ErrorMessage(d, "Error: source was specified, but not destination", ""))
		}
		// If source is http(s), download file locally or remotely
		if strings.HasPrefix(source, "http://") || strings.HasPrefix(source, "https://") {
			if strings.Contains(source, ";") {
				return fmt.Errorf(ErrorMessage(d, "Error: source contains illegal chracter: ", ";"))
			}
			if strings.HasPrefix(source, "https://") && (sourceUser == "" || sourcePass == "") {
				return fmt.Errorf(ErrorMessage(d, "Error: the source_user and source_password properties are required when source is an https URL", ""))
			}

			if remoteHost != "" {
				err := DownloadRemoteFile(d, m)
				if err != nil {
					return fmt.Errorf(ErrorMessage(d, "Error downloading source script", err.Error()))
				}
			} else {
				if strings.HasPrefix(strings.TrimSpace(destination), "/") || strings.HasPrefix(strings.TrimSpace(destination), "~") {
					return fmt.Errorf(ErrorMessage(d, "The destination parameter must be a relative path when downloading locally", ""))
				}
				if strings.Contains(destination, "..") {
					return fmt.Errorf(ErrorMessage(d, "The destination parameter cannot contain reference to a parent directory", ""))
				}
				// Download locally using wget
				wgetCommand := BuildWgetCmd(d, m, destination)
				_, err := LocalExec(d, wgetCommand, nil)
				if err != nil {
					if !sourceNoCheckCert && strings.Contains(err.Error(), "no-check-certificate") {
						return fmt.Errorf(ErrorMessage(d, fmt.Sprintf("Error: Could not verify the certificate for %s. Set the source_no_check_cert parameter to true if you want to ignore the certificate from the source URL.", source), err.Error()))
					} else {
						return fmt.Errorf(ErrorMessage(d, "Error downloading source script to local system", err.Error()))
					}
				}
			}
		} else {
			// Local file verification
			if _, err := os.Stat(source); os.IsNotExist(err) {
				return fmt.Errorf(ErrorMessage(d, fmt.Sprintf("Error: can't find source program %q", source), ""))
			}
			if remoteHost != "" {
				return TransferLocalToRemote(d, m, source)
			} else {
				return fmt.Errorf(ErrorMessage(d, "Error: copying a file from one directory to another on the provider container is not supported", ""))
			}
		}
	} else if destination != "" {
		return fmt.Errorf(ErrorMessage(d, "Error: one of remote_password or remote_key is required when specifying remote_host", ""))
	}
	return nil
}

// Runs a script and returns the output and/or an error if it fails.
// If the script returns JSON (recommended) it will be loaded into a map and returned
// If the script returns a String, the String will be returned
func RunScript(d *schema.ResourceData, m interface{}) (map[string]string, error) {
	programI := d.Get("program").([]interface{})
	programSens := d.Get("program_sensitive").([]interface{})
	query := d.Get("query").(map[string]interface{})
	querySens := d.Get("query_sensitive").(map[string]interface{})
	remote_host := d.Get("remote_host").(string)

	if len(programI) < 1 && len(programSens) < 1 {
		return nil, fmt.Errorf(ErrorMessage(d, "Error: program list must contain at least one element", ""))
	}

	for i, vI := range programI {
		if _, ok := vI.(string); !ok {
			return nil, fmt.Errorf(ErrorMessage(d, fmt.Sprintf("Error: program element %d is %T. a string is required", i, vI), ""))
		}
	}

	for i, vI := range programSens {
		if _, ok := vI.(string); !ok {
			return nil, fmt.Errorf(ErrorMessage(d, fmt.Sprintf("Error: program_sensitive element %d is %T. a string is required", i, vI), ""))
		}
	}

	if remote_host != "" {
		return RunRemoteScript(d, m)
	} else {
		err := HandleSourceAndDest(d, m)
		if err != nil {
			return nil, err
		}
	}

	// first element is assumed to be an executable command, possibly found
	// using the PATH environment variable.
	_, err := exec.LookPath(programI[0].(string))
	if err != nil {
		return nil, fmt.Errorf(ErrorMessage(d, fmt.Sprintf("Error: can't find external program %q", programI[0]), ""))
	}

	for i, v := range querySens {
		query[i] = v
	}

	// Merge the program and program_sensitive arrays
	program := make([]string, len(programI)+len(programSens))
	count := 0
	for i, vI := range programI {
		program[i] = vI.(string)
		count++
	}

	for i, vS := range programSens {
		program[i+count] = vS.(string)
	}

	cmdOutput, err := LocalExec(d, program, query)
	if err != nil {
		return nil, fmt.Errorf(ErrorMessage(d, "Error: error executing local program", err.Error()))
	}

	var result map[string]string
	err = json.Unmarshal(cmdOutput, &result)
	if err != nil {
		// Check if it's JSON, but not key/value pair. If so, we'll return an error.
		var tryTwo map[string]interface{}
		err = json.Unmarshal(cmdOutput, &tryTwo)
		if err == nil {
			return nil, fmt.Errorf(ErrorMessage(d, fmt.Sprintf("Error: command %q produced JSON that was not key value pairs of strings, which is required by Terraform.", program[0]),""))
		}
		// The command did not return JSON, but it did return successfully. Return the response as a String
		result = make(map[string]string)
		result["stdout"] = strings.TrimSpace(string(cmdOutput[:]))
	}
	return result, nil
}

// Contains the base function for executing a command locally. Helper method to RunScript
func LocalExec(d *schema.ResourceData, program []string, query map[string]interface{}) ([]byte, error) {
	cmd := exec.Command(program[0], program[1:]...)

	queryJson, err := json.Marshal(query)
	if err != nil {
		// Should never happen, since we know query will always be a map
		// from string to string, as guaranteed by d.Get and our schema.
		return nil, fmt.Errorf(SubErrorMessage(d, "Error: error converting query JSON to map", err.Error()))
	}

	cmd.Stdin = bytes.NewReader(queryJson)

	cmdOutput, err := cmd.Output()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			if exitErr.Stderr != nil && len(exitErr.Stderr) > 0 {
				return nil, fmt.Errorf(SubErrorMessage(d, fmt.Sprintf("Error: failed to execute %q", program[0]), string(exitErr.Stderr)))
			}
			return nil, fmt.Errorf(SubErrorMessage(d, fmt.Sprintf("Error: command %q failed with no error message", program[0]), ""))
		} else {
			return nil, fmt.Errorf(SubErrorMessage(d, fmt.Sprintf("Error: failed to execute %q", program[0]), err.Error()))
		}
	}
	return cmdOutput, nil
}

// Transfers (if applicable) and executes a command or script on a remote system
func RunRemoteScript(d *schema.ResourceData, m interface{}) (map[string]string, error) {
	programI := d.Get("program").([]interface{})
	programSens := d.Get("program_sensitive").([]interface{})
	query := d.Get("query").(map[string]interface{})
	querySens := d.Get("query_sensitive").(map[string]interface{})

	err := HandleSourceAndDest(d, m)
	if err != nil {
		return nil, err
	}

	config, err := CreateSSHConfig(d, m)
	if err != nil {
		return nil, err
	}

	for i, v := range querySens {
		query[i] = v
	}

	// Merge the program and program_sensitive arrays
	program := make([]string, len(programI)+len(programSens))
	count := 0
	for i, vI := range programI {
		program[i] = vI.(string)
		count++
	}

	for i, vS := range programSens {
		program[i+count] = vS.(string)
	}

	cmdOutput, err := RemoteExec(d, program, query, config)
	if err != nil {
		return nil, fmt.Errorf(ErrorMessage(d, "Error executing remote program", err.Error()))
	}
	var result map[string]string
	err = json.Unmarshal(cmdOutput, &result)
	if err != nil {
		// Check if it's JSON, but not key/value pair. If so, we'll return an error.
		var tryTwo map[string]interface{}
		err = json.Unmarshal(cmdOutput, &tryTwo)
		if err == nil {
			return nil, fmt.Errorf(ErrorMessage(d, fmt.Sprintf("Error: command %q produced JSON that was not key value pairs of strings, which is required by Terraform.", program[0]),""))
		}
		// The command did not return JSON, but it did return successfully. Return the response as a String
		result = make(map[string]string)
		result["stdout"] = strings.TrimSpace(string(cmdOutput[:]))
	}
	return result, nil
}

func getClientUsingBastionConn(d *schema.ResourceData, rmthostConfig *ssh.ClientConfig) (*ssh.Client, *ssh.Client, net.Conn, ssh.Conn, error) {
	var bastionConfig *ssh.ClientConfig
	var basterr error
	bastionHost := d.Get("bastion_host").(string)
	if bastionHost != "" {
		TraceMessage(d, fmt.Sprintf("Using bastion host %s to connect", bastionHost))
		bastionPassword := d.Get("bastion_password").(string)
		bastionPrivateKey := d.Get("bastion_private_key").(string)
		bastionPort := d.Get("bastion_port").(string)
		remoteHost := d.Get("remote_host").(string)
		if bastionPort == "" {
			bastionPort = "22"
		}
		if bastionPassword == "" && bastionPrivateKey == "" {
			return nil, nil, nil, nil, fmt.Errorf(ErrorMessage(d, "Error: Bastion host password and private key is empty", "Provide value for bastion_password or bastion_private_key"))
		}
		bastionConfig, basterr = CreateBastionConfig(d)
		if basterr != nil {
			return nil, nil, nil, nil, fmt.Errorf(SubErrorMessage(d, "Error: error creating bastion ssh config", basterr.Error()))
		}
		bastionclient, err := ssh.Dial("tcp", bastionHost+":"+bastionPort, bastionConfig)
		if err != nil {
			return nil, nil, nil, nil, fmt.Errorf(SubErrorMessage(d, "Error: error creating bastion client", err.Error()))
		}
		//defer bastionclient.Close()
		bastTohostConn, err := bastionclient.Dial("tcp", remoteHost+":22")
		if err != nil {
			return nil, nil, nil, nil, fmt.Errorf(SubErrorMessage(d, "Error: error creating connection to remote host using bastion client", err.Error()))
		}
		//defer bastTohostConn.Close()
		sshConn, sshChan, req, err := ssh.NewClientConn(bastTohostConn, remoteHost, rmthostConfig)
		if err != nil {
			return nil, nil, nil, nil, fmt.Errorf(SubErrorMessage(d, "Error: error creating connection to remote host using bastion connection to remote host", err.Error()))
		}
		//defer sshConn.Close()
		localToRemoteClient := ssh.NewClient(sshConn, sshChan, req)
		return localToRemoteClient, bastionclient, bastTohostConn, sshConn, nil
	} else {
		return nil, nil, nil, nil, nil
	}
}

// Contains the base function for executing a command remotely. Helper method to RunRemoteScript
func RemoteExec(d *schema.ResourceData, program []string, query map[string]interface{}, config *ssh.ClientConfig) ([]byte, error) {
	var session *ssh.Session
	var sessionerr error
	bastionHost := d.Get("bastion_host").(string)
	remoteHost := d.Get("remote_host").(string)
	queryJson, err := json.Marshal(query)
	if err != nil {
		// Should never happen, since we know query will always be a map
		// from string to string, as guaranteed by d.Get and our schema.
		return nil, fmt.Errorf(SubErrorMessage(d, "Error: error converting query JSON to map", err.Error()))
	}    
	var clientcp *ssh.Client
	if bastionHost != "" {
		localToRemoteClient, bastionclient, bastTohostConn, sshConn, err := getClientUsingBastionConn(d, config)
		clientcp = localToRemoteClient
		if err != nil {
			return nil, fmt.Errorf(SubErrorMessage(d, "Error: error connecting using bastion host", err.Error()))
		}
		defer bastionclient.Close()
		defer bastTohostConn.Close()
		defer sshConn.Close()
		if localToRemoteClient != nil {
			defer localToRemoteClient.Close()
			session, sessionerr = localToRemoteClient.NewSession()
			if sessionerr != nil {
				return nil, fmt.Errorf(SubErrorMessage(d, "Error: error creating new session to remote host using bastion host", sessionerr.Error()))
			}
			defer session.Close()
		} else {
			return nil, fmt.Errorf(ErrorMessage(d, "Error: error connecting using bastion host", ""))
		}
	} else {
		// Connect
		client, err := ssh.Dial("tcp", remoteHost+":22", config)
		clientcp = client
		if err != nil {
			return nil, fmt.Errorf(SubErrorMessage(d, "Error: error creating remote client", err.Error()))
		}
		session, sessionerr = client.NewSession()
		if sessionerr != nil {
			return nil, fmt.Errorf(SubErrorMessage(d, "Error: error creating remote connection", sessionerr.Error()))
		}
		defer session.Close()
	}
	//Create a context that will be used to send Done event to keepalive go routine 
	//when script execution is completed or results in error.
	ctx, cancelKeepAlive := context.WithCancel(context.TODO())
	//go routine to send async ssh request to server every 15 seconds - mimics keepalive.
	go func() {
		t := time.NewTicker(15 * time.Second)
		defer t.Stop()
		for {
			select {
			case <-t.C:
				//TraceMessage(d, fmt.Sprintf("Executing script ..."))
				_, _, err := clientcp.SendRequest("keepalive@ibm.com", true, nil)
				if err != nil {
					return
				}
			case <-ctx.Done():
				return
			}
		}
	}()
	var b bytes.Buffer
	session.Stdin = bytes.NewBufferString(string(queryJson[:]))
	session.Stdout = &b // get output
	errpipe, _ := session.StderrPipe()	// get error pipe
	err = session.Run(strings.Join(program, " "))
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			if exitErr.Stderr != nil && len(exitErr.Stderr) > 0 {
				//Terminate keepalive routine by sending context cancel message to trigger ctx.Done.
				cancelKeepAlive()
				return nil, fmt.Errorf(SubErrorMessage(d, fmt.Sprintf("Error: failed to execute %q", program[0]), string(exitErr.Stderr)))
			}
			//Terminate keepalive routine by sending context cancel message to trigger ctx.Done.
			cancelKeepAlive()
			return nil, fmt.Errorf(SubErrorMessage(d, fmt.Sprintf("Error: command %q failed with no error message", program[0]), ""))	
		} else {		
			//Remote error exit throws ssh.ExitError.
			//Get the error message from error pipe.
			errout := ""
			if errpipe != nil{
	            sc := bufio.NewScanner(errpipe)
	            var errbuffer bytes.Buffer
	            for sc.Scan() {
	            	errbuffer.WriteString(sc.Text())
	            }
	            errbuffer.WriteString("\n")
	            errbuffer.WriteString(err.Error())
	            errout = errbuffer.String()
			}
			if errout == ""{
				errout = err.Error()
			} 
			//Terminate keepalive routine by sending context cancel message to trigger ctx.Done.
			cancelKeepAlive()
			return nil, fmt.Errorf(SubErrorMessage(d, fmt.Sprintf("Error: failed to execute %q", program[0]), errout))
		}
	}
	//Terminate keepalive routine by sending context cancel message to trigger ctx.Done.
	cancelKeepAlive()
	return b.Bytes(), nil
}
