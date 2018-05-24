package main

import (
	"testing"

	"github.com/cloudfoundry-community/go-cfclient"
)

var clearDBkey = cfclient.ServiceKey{
	Credentials: map[string]interface{}{
		"hostname": "us-cdbr-iron-east-00.cleardb.net",
		"jdbcUrl":  "jdbc:mysql://us-cdbr-iron-east-00.cleardb.net/ad_4hbbj34jh34b3h4?user=34in34i34in\u0026password=foooooo",
		"name":     "ad_4hbbj34jh34b3h4",
		"password": "foooooo",
		"port":     "3306",
		"uri":      "mysql://34in34i34in:foooooo@us-cdbr-iron-east-04.cleardb.net:3306/ad_4hbbj34jh34b3h4?reconnect=true",
		"username": "34in34i34in",
	}}

func TestGetHostnamesFromServiceKey_cleardb(t *testing.T) {
	plugin := TLSEnablerPlugin{}
	expectedHostnames := []string{"us-cdbr-iron-east-00.cleardb.net"}

	hostnames := plugin.getHostnamesFromServiceKey(clearDBkey)
	if hostnames[0] != expectedHostnames[0] {
		t.Errorf("That's not what I expected!")
	}
}
