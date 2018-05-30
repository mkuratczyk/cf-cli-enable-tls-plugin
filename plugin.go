package main

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"code.cloudfoundry.org/cli/plugin"
	"code.cloudfoundry.org/cli/plugin/models"
)

// TLSEnablerPlugin allows you to quickly enable TLS for selected service instances
type TLSEnablerPlugin struct {
	cliConnection plugin.CliConnection
	serviceName   string
}

// maps supported service type to the arbitrary parameter name
var supportedServices = map[string]string{
	"p.rabbitmq":             "tls",
	"p.mysql":                "enable_tls",
	"rabbitmq-odb-bosh-lite": "tls",
}

// Run is the main entry point for CF CLI plugins
func (t *TLSEnablerPlugin) Run(cliConnection plugin.CliConnection, args []string) {
	t.cliConnection = cliConnection

	switch args[0] {
	case "CLI-MESSAGE-UNINSTALL":
		return
	case "enable-tls":
		if len(args) != 2 {
			log.Fatalln("USAGE: cf enable-tls SERVICE_NAME")
		}
		t.serviceName = args[1]
		err := t.enableTLS()
		if err != nil {
			log.Fatalf("Failed to enable TLS: %v", err)
		}
	case "create-service-with-tls":
		if len(args) < 4 {
			log.Fatalln("USAGE: create-service-with-tls SERVICE PLAN SERVICE_INSTANCE [-c PARAMETERS_AS_JSON] [-t TAGS]")
		}
		args[0] = "create-service"
		t.serviceName = args[3]
		err := t.createServiceWithTLS(args)
		if err != nil {
			log.Fatalf("Failed to create service with TLS: %v", err)
		}
	default:
		log.Fatalf("Received unexpected command %v\n", args[0])
	}

}

// GetMetadata returns plugin information
func (t *TLSEnablerPlugin) GetMetadata() plugin.PluginMetadata {
	return plugin.PluginMetadata{
		Name: "TLSEnabler",
		Version: plugin.VersionType{
			Major: 0,
			Minor: 2,
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
			{
				Name:     "create-service-with-tls",
				Alias:    "",
				HelpText: "executes create-service and then immediately enable-tls",
				UsageDetails: plugin.Usage{
					Usage: "create-service-with-tls SERVICE PLAN SERVICE_INSTANCE [-c PARAMETERS_AS_JSON] [-t TAGS]",
				},
			},
		},
	}
}

func (t *TLSEnablerPlugin) enableTLS() error {
	log.Printf("Enabling TLS on service %v\n", t.serviceName)

	serviceInfo, err := t.cliConnection.GetService(t.serviceName)
	if err != nil {
		return err
	}

	if _, ok := supportedServices[serviceInfo.ServiceOffering.Name]; !ok {
		log.Fatalf("Sorry, I don't know how to enable TLS on an instance of %v service\n", serviceInfo.ServiceOffering.Name)
	}

	arbitraryParameters, err := t.buildArbitraryParameters(serviceInfo)
	if err != nil {
		return err
	}

	_, err = t.cliConnection.CliCommand("update-service", t.serviceName, "-c", arbitraryParameters)
	if err != nil {
		return err
	}

	return nil
}

func (t *TLSEnablerPlugin) createServiceWithTLS(args []string) error {
	_, err := t.cliConnection.CliCommand(args...)
	if err != nil {
		return err
	}
	// wait for `create` to complete
	t.waitUntilServiceCreated()
	return t.enableTLS()
}

func (t *TLSEnablerPlugin) waitUntilServiceCreated() error {
	for {
		service, err := t.cliConnection.GetService(t.serviceName)
		if err != nil {
			return err
		}

		if service.LastOperation.State == "succeeded" {
			return nil
		} else if service.LastOperation.State == "failed" {
			return fmt.Errorf(
				"error %s [status: %s]",
				service.LastOperation.Description,
				service.LastOperation.State,
			)
		}
		time.Sleep(500 * time.Millisecond)
	}

}

func (t *TLSEnablerPlugin) buildArbitraryParameters(serviceInfo plugin_models.GetService_Model) (string, error) {
	serviceKeyName := "temporary-key-to-enable-tls"
	_, err := t.cliConnection.CliCommand("create-service-key", serviceInfo.Name, serviceKeyName)
	if err != nil {
		return "", err
	}

	serviceKey, err := t.getServiceKey(serviceKeyName)
	if err != nil {
		return "", err
	}
	// ideally it should be used with defer() but it doesn't work (gets triggered but the key doesn't get deleted)
	t.cliConnection.CliCommand("delete-service-key", "-f", serviceInfo.Name, serviceKeyName)

	hostnames := t.getHostnamesFromServiceKey(serviceKey)
	return fmt.Sprintf("{\"%v\": [%v]}", supportedServices[serviceInfo.ServiceOffering.Name], strings.Join(hostnames, ",")), nil
}

func (t *TLSEnablerPlugin) getServiceKey(serviceKeyName string) (map[string]interface{}, error) {
	output, err := t.cliConnection.CliCommand("service-key", t.serviceName, serviceKeyName)
	if err != nil {
		log.Fatal(err)
	}
	var serviceKey map[string]interface{}
	json.Unmarshal([]byte(strings.Join(output[2:], "")), &serviceKey)
	return serviceKey, nil
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
