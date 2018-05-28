package main

import (
	"fmt"
	"log"
	"strings"

	"code.cloudfoundry.org/cli/plugin"
	cfclient "github.com/cloudfoundry-community/go-cfclient"
)

// TLSEnablerPlugin allows you to quickly enabled TLS on a service instance of MySQL for PCF v2.3
type TLSEnablerPlugin struct {
	cliConnection plugin.CliConnection
}

// maps supported service type to arbitrary parameter name
var supportedServices = map[string]string{
	"p.rabbitmq":             "tls",
	"p.mysql":                "enable_tls",
	"rabbitmq-odb-bosh-lite": "tls",
}

// Run is the main entry point for CF CLI plugins
func (t *TLSEnablerPlugin) Run(cliConnection plugin.CliConnection, args []string) {
	switch args[0] {
	case "CLI-MESSAGE-UNINSTALL":
		return
	case "enable-tls":
		if len(args) != 2 {
			log.Fatalln("USAGE: cf enable-tls SERVICE_NAME")
		}
	default:
		log.Fatalf("Received unexpected command %v\n", args[0])
	}

	t.cliConnection = cliConnection

	err := t.enableTLS(args[1])
	if err != nil {
		log.Fatalf("Failed to enable TLS: %v", err)
	}
}

// GetMetadata return plugin information
func (t *TLSEnablerPlugin) GetMetadata() plugin.PluginMetadata {
	return plugin.PluginMetadata{
		Name: "TLSEnabler",
		Version: plugin.VersionType{
			Major: 0,
			Minor: 1,
			Build: 0,
		},
		Commands: []plugin.Command{
			{
				Name:     "enable-tls",
				Alias:    "",
				HelpText: "Enables TLS on the specified service instance",
				UsageDetails: plugin.Usage{
					Usage: "enable-tls NAME - enable TLS on service instance NAME",
				},
			},
		},
	}
}

func (t *TLSEnablerPlugin) enableTLS(serviceName string) error {
	log.Printf("Enabling TLS on service %v\n", serviceName)

	serviceInfo, err := t.cliConnection.GetService(serviceName)
	if err != nil {
		return err
	}

	if _, ok := supportedServices[serviceInfo.ServiceOffering.Name]; !ok {
		log.Fatalf("Sorry, I don't know how to enable TLS on an instance of %v service\n", serviceInfo.ServiceOffering.Name)
	}

	serviceKeyName := "temporary-key-to-enable-tls"
	_, err = t.cliConnection.CliCommand("create-service-key", serviceName, serviceKeyName)
	if err != nil {
		return err
	}

	serviceKey, err := t.getServiceKey(serviceInfo.Guid, serviceKeyName)
	if err != nil {
		return err
	}
	// ideally it should be used with defer() but it doesn't work (gets triggered but the key doesn't get deleted)
	t.cliConnection.CliCommand("delete-service-key", "-f", serviceName, serviceKeyName)

	hostnames := t.getHostnamesFromServiceKey(serviceKey.Credentials.(map[string]interface{}))
	arbitraryParameters := fmt.Sprintf("{\"%v\": [%v]}", supportedServices[serviceInfo.ServiceOffering.Name], strings.Join(hostnames, ","))
	_, err = t.cliConnection.CliCommand("update-service", serviceName, "-c", arbitraryParameters)
	if err != nil {
		return err
	}

	return nil
}

func (t *TLSEnablerPlugin) getServiceKey(serviceGUID string, serviceKeyName string) (cfclient.ServiceKey, error) {
	apiEndpoint, err := t.cliConnection.ApiEndpoint()
	if err != nil {
		return cfclient.ServiceKey{}, err
	}
	apiToken, err := t.cliConnection.AccessToken()
	if err != nil {
		return cfclient.ServiceKey{}, nil
	}

	c := &cfclient.Config{
		ApiAddress:        apiEndpoint,
		Token:             strings.Split(apiToken, " ")[1],
		SkipSslValidation: true,
	}

	client, err := cfclient.NewClient(c)
	if err != nil {
		log.Println(err)
	}
	serviceKey, err := client.GetServiceKeysByInstanceGuid(serviceGUID)
	if err != nil {
		log.Print(err)
	}

	return serviceKey[0], nil
}

func (t *TLSEnablerPlugin) getHostnamesFromServiceKey(serviceKey map[string]interface{}) []string {
	var hs []string
	if hostnames, ok := serviceKey["hostnames"]; ok {
		for _, h := range hostnames.([]interface{}) {
			hs = append(hs, fmt.Sprintf("\"%v\"", h.(string)))
		}
		return hs
	}

	// this is a single-node service which doesn't reutrn `hostnames`
	hs = []string{fmt.Sprintf("\"%v\"", serviceKey["hostname"].(string))}
	return hs
}
