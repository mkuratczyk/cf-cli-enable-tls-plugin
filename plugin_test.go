package main

import (
	"encoding/json"
	"io/ioutil"
	"path/filepath"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
)

var _ = Describe("getHostnamesFromServiceKey", func() {
	DescribeTable("can parse service keys of known service types",
		func(fixtureFile string, expectedHostnames []string) {
			fixture, _ := ioutil.ReadFile(filepath.Join("fixtures", fixtureFile))
			var serviceKey map[string]interface{}
			json.Unmarshal(fixture, &serviceKey)

			plugin := TLSEnablerPlugin{}
			hostnames := plugin.getHostnamesFromServiceKey(serviceKey)
			Expect(hostnames).To(Equal(expectedHostnames))
		},
		Entry("single node MySQL", "p.mysql-single-node.json", []string{"\"10.1.2.3\""}),
	)

})
