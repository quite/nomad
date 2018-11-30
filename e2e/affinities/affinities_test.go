package e2e

import (
	"testing"

	"github.com/hashicorp/nomad/api"
	"github.com/hashicorp/nomad/e2e/framework"
	"github.com/hashicorp/nomad/helper"
	"github.com/hashicorp/nomad/helper/uuid"
	"github.com/hashicorp/nomad/jobspec"
	"github.com/stretchr/testify/require"

	"time"

	. "github.com/onsi/gomega"
)

type AffinitiesTest struct {
	framework.TC
	jobIds []string
}

func (tc *AffinitiesTest) TestSingleAffinities(f *framework.F) {
	nomadClient := tc.Nomad()

	// Parse job
	job, err := jobspec.ParseFile("input/aff1.nomad")
	require := require.New(f.T())
	require.Nil(err)
	jobId := uuid.Generate()
	job.ID = helper.StringToPtr(jobId)

	//TODO(preetha) fix me - this means tests can't run in parallel
	tc.jobIds = append(tc.jobIds, jobId)

	// Register job
	jobs := nomadClient.Jobs()
	resp, _, err := jobs.Register(job, nil)
	require.Nil(err)
	require.NotEmpty(resp.EvalID)

	g := NewGomegaWithT(f.T())

	// Wrap in retry to wait until placement
	g.Eventually(func() []*api.AllocationListStub {
		// Look for allocations
		allocs, _, _ := jobs.Allocations(*job.ID, false, nil)
		return allocs
	}, 2*time.Second, time.Second).ShouldNot(BeEmpty())

	jobAllocs := nomadClient.Allocations()

	allocs, _, _ := jobs.Allocations(*job.ID, false, nil)

	// Verify affinity score metadata
	for _, allocStub := range allocs {
		alloc, _, err := jobAllocs.Info(allocStub.ID, nil)
		require.Nil(err)
		require.NotEmpty(alloc.Metrics.ScoreMetaData)
		// Expect node affinity score to be 1.0 for all nodes
		// This only passes in a cluster where all nodes are in dc1
		// TODO(preetha) make this more realistic
		for _, sm := range alloc.Metrics.ScoreMetaData {
			require.Equal(1.0, sm.Scores["node-affinity"])
		}
	}

}

func (tc *AffinitiesTest) AfterEach(f *framework.F) {
	nomadClient := tc.Nomad()
	jobs := nomadClient.Jobs()
	// Stop all jobs in test
	for _, id := range tc.jobIds {
		jobs.Deregister(id, true, nil)
	}
	// Garbage collect
	nomadClient.System().GarbageCollect()
}

func TestCalledFromGoTest(t *testing.T) {
	framework.New().AddSuites(&framework.TestSuite{
		Component: "foo",
		Cases: []framework.TestCase{
			new(AffinitiesTest),
		},
	}).Run(t)
}
