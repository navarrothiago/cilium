// Copyright 2020 Authors of Cilium
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package k8sTest

import (
	"context"
	"fmt"

	. "github.com/cilium/cilium/test/ginkgo-ext"
	"github.com/cilium/cilium/test/helpers"

	. "github.com/onsi/gomega"
)

const (
	SpireNamespace             = "spire"
	SpireServerLabel           = "app=spire-server"
	SpireAgentLabel            = "app=spire-agent"
	PoddefaultLabel            = "app=poddefault"
	PodfooLabel                = "app=podfoo"
	XwingLabel                 = "org=alliance"
	DeathStarLabel             = "org=empire"
	spiffeIdSADefaultNSDefault = "spiffe://example.org/ns/default/sa/default"
	spiffeIdSAFooNSDefault     = "spiffe://example.org/ns/default/sa/foo"
	spiffeIdXwing              = "spiffe://example.org/xwing"
	spiffeIdDeathstar          = "spiffe://example.org/deathstar"
	spiffeIdCiliumAgent        = "spiffe://example.org/ciliumagent"
	spiffeIdSpireAgent         = "spiffe://example.org/ns/spire/sa/spire-agent"
	CmdShowEntries             = "/opt/spire/bin/spire-server entry show"
)

var _ = SkipDescribeIf(func() bool {
	return helpers.RunsOnGKE() || helpers.RunsOn419Kernel() || helpers.RunsOn54Kernel()
}, "K8sSpiffe", func() {
	var (
		kubectl        *helpers.Kubectl
		ciliumFilename string
		spire          string
		spireServerPod string
	)

	checkSpiffeIdAssignment := func(podLabel, spiffeId string) {
		ciliumPodK8s1, err := kubectl.GetCiliumPodOnNode(helpers.K8s1)
		ExpectWithOffset(1, err).ShouldNot(HaveOccurred(), "Cannot determine cilium pod name")

		pods, err := kubectl.GetPodNames(helpers.DefaultNamespace, podLabel)
		Expect(err).To(BeNil(), "Cannot get pods names")
		Expect(len(pods)).To(BeNumerically(">", 0), "No pods available to spire scenario 01")

		cmd := fmt.Sprintf("cilium endpoint get -l '%s'", spiffeId)
		kubectl.CiliumExecUntilMatch(ciliumPodK8s1, cmd, spiffeId)

		kubectl.CiliumExecContext(
			context.TODO(),
			ciliumPodK8s1,
			cmd,
		).ExpectSuccess()
	}

	createEntryOnSpireServer := func(spiffeId, parentId string, selectors []string, ttl string, node bool) {

		By(fmt.Sprintf("creating restriation entries for %s", SpireServerLabel))

		spireServerPod, _ = fetchPodsWithOffset(kubectl, SpireNamespace, "spire-server", SpireServerLabel, "", false, 0)
		cmd := fmt.Sprintf("/opt/spire/bin/spire-server entry create -spiffeID %s", spiffeId)

		if node {
			cmd = fmt.Sprintf("%s -node", cmd)
		}

		if ttl != "" {
			cmd = fmt.Sprintf("%s -ttl %s", cmd, ttl)
		}

		if parentId != "" {
			cmd = fmt.Sprintf("%s -parentID %s", cmd, parentId)
		}

		for _, selector := range selectors {
			cmd = fmt.Sprintf("%s -selector %s", cmd, selector)
		}

		kubectl.ExecPodCmd(SpireNamespace, spireServerPod, cmd)
	}

	BeforeAll(func() {
		kubectl = helpers.CreateKubectl(helpers.K8s1VMName(), logger)

		ciliumFilename = helpers.TimestampFilename("cilium.yaml")
		DeployCiliumOptionsAndDNS(kubectl, ciliumFilename, map[string]string{
			"spiffe.enabled": "true",
		})

		_, err := kubectl.CiliumNodesWait()
		ExpectWithOffset(1, err).Should(BeNil(), "Failure while waiting for k8s nodes to be annotated by Cilium")

		By("making sure all endpoints are in ready state")
		err = kubectl.CiliumEndpointWaitReady()
		ExpectWithOffset(1, err).To(BeNil(), "Failure while waiting for all cilium endpoints to reach ready state")

		By("deploying spire components")
		spire = helpers.ManifestGet(kubectl.BasePath(), "spire.yaml")
		kubectl.ApplyDefault(spire).ExpectSuccess("Cannot import spire components")
		testNamespace := SpireNamespace

		By("making sure all spire components are in ready state")
		err = kubectl.WaitforPods(testNamespace, fmt.Sprintf("-l %s", SpireAgentLabel), helpers.HelperTimeout)
		Expect(err).Should(BeNil())
		err = kubectl.WaitforPods(testNamespace, fmt.Sprintf("-l %s", SpireServerLabel), helpers.HelperTimeout)
		Expect(err).Should(BeNil())

		createEntryOnSpireServer(spiffeIdSpireAgent, "", []string{"k8s_sat:cluster:demo-cluster", "k8s_sat:agent_ns:spire", "k8s_sat:agent_sa:spire-agent"}, "", true)
		createEntryOnSpireServer(spiffeIdCiliumAgent, "", []string{"unix:uid:0"}, "", true)

	})

	AfterAll(func() {
		// TODO navarrrothiago - delete registrations entries
		kubectl.Delete(spire)
		UninstallCiliumFromManifest(kubectl, ciliumFilename)
		kubectl.CloseSSHClient()
	})

	AfterFailed(func() {
		kubectl.CiliumReport("cilium endpoint list")
	})

	JustAfterEach(func() {
		kubectl.ValidateNoErrorsInLogs(CurrentGinkgoTestDescription().Duration)
	})

	Context("when poddefault and podfoo are deployed", func() {
		var (
			spireScenario01 string
		)

		BeforeAll(func() {
			spireScenario01 = helpers.ManifestGet(kubectl.BasePath(), "spire-scenario01.yaml")
			kubectl.ApplyDefault(spireScenario01).ExpectSuccess("Cannot import spire scenario 01 components")

			By("making sure all scenario01 are in ready state")
			err := kubectl.WaitforPods(helpers.DefaultNamespace, fmt.Sprintf("-l %s", PoddefaultLabel), helpers.HelperTimeout)
			Expect(err).Should(BeNil())

			createEntryOnSpireServer(spiffeIdSADefaultNSDefault, spiffeIdSpireAgent, []string{"k8s:ns:default", "k8s:sa:default"}, "60", false)
			createEntryOnSpireServer(spiffeIdSAFooNSDefault, spiffeIdSpireAgent, []string{"k8s:ns:default", "k8s:sa:foo"}, "60", false)

		})

		AfterAll(func() {
			kubectl.Delete(spireScenario01)
			// TODO navarrrothiago - delete registrations entries
		})

		It("should assign a spiffe id label to them", func() {
			checkSpiffeIdAssignment(PoddefaultLabel, spiffeIdSADefaultNSDefault)
			checkSpiffeIdAssignment(PodfooLabel, spiffeIdSAFooNSDefault)
		})

		Context("and when the workload poddefault is allowed to communicate only with workload podfoo", func() {
			var (
				cnpSpiffeAllowDefault string
				cnpDefaultDenyEgress  string
			)

			BeforeAll(func() {
				cnpSpiffeAllowDefault = helpers.ManifestGet(kubectl.BasePath(), "cnp-spiffe-allow-sa-default-ns-default.yaml")
				cnpDefaultDenyEgress = helpers.ManifestGet(kubectl.BasePath(), "cnp-default-deny-egress.yaml")

				By("Applying deny all egress policy for poddefault")
				_, err := kubectl.CiliumPolicyAction(helpers.DefaultNamespace, cnpDefaultDenyEgress, helpers.KubectlApply, helpers.HelperTimeout)
				Expect(err).Should(BeNil(), "L3 deny all Policy cannot be applied in %q namespace", helpers.DefaultNamespace)

				By("Applying allow policy for poddefault to podfoo")
				_, err = kubectl.CiliumPolicyAction(helpers.DefaultNamespace, cnpSpiffeAllowDefault, helpers.KubectlApply, helpers.HelperTimeout)
				Expect(err).Should(BeNil(), "L3 spiffeAllowDefault Policy cannot be applied in %q namespace", helpers.DefaultNamespace)
			})

			AfterAll(func() {
				kubectl.Delete(cnpSpiffeAllowDefault)
				kubectl.Delete(cnpDefaultDenyEgress)
			})

			It("should allow to ping from poddefault to podfoo", func() {
				By("Retrieving the poddefault IP address")
				podfoo, podfooJson := fetchPodsWithOffset(kubectl, helpers.DefaultNamespace, "poddefault", PoddefaultLabel, "", false, 0)
				podfooIpAddress, err := podfooJson.Filter("{.status.podIP}")
				Expect(err).Should(BeNil(), "Failure to retrieve IP of pod %s", podfoo)

				pods, err := kubectl.GetPodNames(helpers.DefaultNamespace, PoddefaultLabel)
				Expect(err).To(BeNil(), "Cannot get pods names")
				Expect(len(pods)).To(BeNumerically(">", 0), "No pods available to spire scenario 01")

				By("Pinging from poddefault to podfoo")
				res := kubectl.ExecPodCmd(helpers.DefaultNamespace, pods[0], helpers.Ping(podfooIpAddress.String()))
				res.ExpectSuccess(fmt.Sprintf("Failed to ping from %s to %s", pods[0], podfooIpAddress.String()))
			})

			It("should not allow to make a HTTPs request from poddefault to cilium.io", func() {
				pods, err := kubectl.GetPodNames(helpers.DefaultNamespace, PoddefaultLabel)
				Expect(err).To(BeNil(), "Cannot get pods names")
				Expect(len(pods)).To(BeNumerically(">", 0), "No pods available to spire scenario 01")

				By("HTTPs request from poddefault to cilium.io")
				res := kubectl.ExecPodCmd(helpers.DefaultNamespace, pods[0], helpers.CurlFail("https://cilium.io"))
				res.ExpectFail("Unexpected connection from %q to 'https://cilium.io'", pods[0])
			})
		})
	})

	Context("when upgrade the connection to mTLS between xwing and deathstar", func() {
		var (
			spireScenario03  string
			cnpXwingMTLS     string
			cnpDeathstarMTLS string
		)

		BeforeAll(func() {
			spireScenario03 = helpers.ManifestGet(kubectl.BasePath(), "spire-scenario03.yaml")
			kubectl.ApplyDefault(spireScenario03).ExpectSuccess("Cannot import spire scenario 03 components")

			By("making sure all scenario03 are in ready state")
			err := kubectl.WaitforPods(helpers.DefaultNamespace, fmt.Sprintf("-l %s", XwingLabel), helpers.HelperTimeout)
			Expect(err).Should(BeNil())
			err = kubectl.WaitforPods(helpers.DefaultNamespace, fmt.Sprintf("-l %s", DeathStarLabel), helpers.HelperTimeout)
			Expect(err).Should(BeNil())

			createEntryOnSpireServer(spiffeIdXwing, spiffeIdSpireAgent, []string{"k8s:pod-label:class:xwing"}, "60", false)
			createEntryOnSpireServer(spiffeIdDeathstar, spiffeIdSpireAgent, []string{"k8s:pod-label:class:deathstar"}, "60", false)

			checkSpiffeIdAssignment(XwingLabel, spiffeIdXwing)
			checkSpiffeIdAssignment(DeathStarLabel, spiffeIdDeathstar)

			By("Applying xwing mTLS upgrade connection policy")
			cnpXwingMTLS = helpers.ManifestGet(kubectl.BasePath(), "cnp-mtls-xwing-upgrade.yaml")
			_, err = kubectl.CiliumPolicyAction(helpers.DefaultNamespace, cnpXwingMTLS, helpers.KubectlApply, helpers.HelperTimeout)
			Expect(err).Should(BeNil(), "xwing mTLS upgrade policy cannot be applied in %q namespace", helpers.DefaultNamespace)

			By("Applying deathstar mTLS downgrade connection policy")
			cnpDeathstarMTLS = helpers.ManifestGet(kubectl.BasePath(), "cnp-mtls-deathstar-downgrade.yaml")
			_, err = kubectl.CiliumPolicyAction(helpers.DefaultNamespace, cnpDeathstarMTLS, helpers.KubectlApply, helpers.HelperTimeout)
			Expect(err).Should(BeNil(), "deathstar mTLS downgrade policy cannot be applied in %q namespace", helpers.DefaultNamespace)
		})

		AfterAll(func() {
			// TODO navarrrothiago - delete registrations entries
			kubectl.Delete(cnpXwingMTLS)
			kubectl.Delete(cnpDeathstarMTLS)
			kubectl.Delete(spireScenario03)
		})

		It("should allow the communicatation from swing to deathstar", func() {
			pods, err := kubectl.GetPodNames(helpers.DefaultNamespace, XwingLabel)
			Expect(err).To(BeNil(), "Cannot get pods names")
			Expect(len(pods)).To(BeNumerically(">", 0), "No pods available to spire scenario 03")

			By("Curl from swing to deathstar")
			url := "http://deathstar.default.svc.cluster.local/v1/request-landing"
			res := kubectl.ExecPodCmd(helpers.DefaultNamespace, pods[0], helpers.CurlFail(fmt.Sprintf("-XPOST %s", url)))
			res.ExpectSuccess("Unexpected connection from %q to '%s'", pods[0], url)
		})

		Context("and when change the peerId to an unregister entry in deathstar mTLS policy", func() {
			BeforeAll(func() {
				By("Updating deathstar mTLS downgrade connection policy")

				res := kubectl.DeleteAndWait(cnpDeathstarMTLS, false)
				res.ExpectSuccess(fmt.Sprintf("Failed do delete cnp %s", cnpDeathstarMTLS))

				cnpDeathstarMTLS = helpers.ManifestGet(kubectl.BasePath(), "cnp-mtls-deathstar-downgrade-unregister.yaml")
				_, err := kubectl.CiliumPolicyAction(helpers.DefaultNamespace, cnpDeathstarMTLS, helpers.KubectlApply, helpers.HelperTimeout)
				Expect(err).Should(BeNil(), "deathstar mTLS downgrade policy cannot be applied in %q namespace", helpers.DefaultNamespace)
			})

			It("should deny the communication from swing to deathstar", func() {
				pods, err := kubectl.GetPodNames(helpers.DefaultNamespace, XwingLabel)
				Expect(err).To(BeNil(), "Cannot get pods names")
				Expect(len(pods)).To(BeNumerically(">", 0), "No pods available to spire scenario 03")

				By("Curl from swing to deathstar")
				url := "http://deathstar.default.svc.cluster.local/v1/request-landing"
				res := kubectl.ExecPodCmd(helpers.DefaultNamespace, pods[0], helpers.CurlFail(fmt.Sprintf("-XPOST %s", url)))
				res.ExpectFail("Unexpected connection from %q to '%s'", pods[0], url)
			})
		})
	})
})
