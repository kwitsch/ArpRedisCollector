package config

import (
	"os"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Config", func() {
	When("Get", func() {
		It("Subnet", func() {
			const adr string = "192.168.100.100:6379"

			os.Setenv("ARC_REDIS_ADDRESS", adr)
			os.Setenv("ARC_ARP_SUBNET_1", "192.168.100.0/24")

			c, err := Get()

			Expect(err).To(Succeed())

			Expect(c.Redis.Address).To(Equal(adr))
			Expect(len(c.Arp.Subnets)).To(Equal(1))
			Expect(c.Arp.Subnets[0].IP.String()).To(Equal("192.168.100.0"))
		})
	})
})
