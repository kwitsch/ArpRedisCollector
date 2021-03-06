package net

import (
	"fmt"
	"net"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Net", func() {
	When("GetAllLocalNets", func() {
		nets, err := GetAllLocalNets()

		Expect(err).To(Succeed())
		Expect(len(nets)).To(BeNumerically(">", 1))

		fmt.Println("GetAllLocalNets:")
		for _, n := range nets {
			fmt.Println(n.String())
		}
		fmt.Println("-------------------------------")
	})
	When("GetFilteredLocalNets", func() {
		_, n, err1 := net.ParseCIDR("192.168.100.0/24")

		Expect(err1).To(Succeed())

		filter := []*net.IPNet{n}

		res, err2 := GetFilteredLocalNets(filter)

		Expect(err2).To(Succeed())
		Expect(len(res)).To(BeNumerically("==", 1))

		fmt.Println("GetFilteredLocalNets:")
		for _, r := range res {
			fmt.Println(r.String())
		}
		fmt.Println("-------------------------------")
	})
})
