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

var supportedServices = map[string]bool{
	"p.rabbitmq":             true,
	"p.mysql":                true,
	"rabbitmq-odb-bosh-lite": true,
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

	randomKeyName := "temporary-key-to-enable-tls"

	t.cliConnection.CliCommand("create-service-key", serviceName, randomKeyName)

	sans := t.getHostnames(serviceInfo.Guid, randomKeyName)

	if _, ok := supportedServices[serviceInfo.ServiceOffering.Name]; !ok {
		log.Fatalf("Sorry, I don't know how to enable TLS on an instance of %v service\n", serviceInfo.ServiceOffering.Name)
	}

	arbitraryParameters := fmt.Sprintf("{\"tls\": \"%v\"}", sans)
	_, err = t.cliConnection.CliCommand("update-service", serviceName, "-c", arbitraryParameters)
	if err != nil {
		return err
	}

	t.cliConnection.CliCommand("delete-service-key", "-f", serviceName, randomKeyName)

	return nil
}

func (t *TLSEnablerPlugin) getHostnames(serviceGUID string, randomKeyName string) []string {
	apiEndpoint, err := t.cliConnection.ApiEndpoint()
	if err != nil {
		return nil
	}
	apiToken, err := t.cliConnection.AccessToken()
	if err != nil {
		return nil
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

	return []string{serviceKey[0].Credentials.(map[string]interface{})["hostname"].(string)}
}
