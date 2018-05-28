package main_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestCfCliEnableTlsPlugin(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "CfCliEnableTlsPlugin Suite")
}
