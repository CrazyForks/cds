package openstack

import (
	"context"
	"testing"

	"github.com/gophercloud/gophercloud/openstack/compute/v2/flavors"
	"github.com/ovh/cds/sdk"
	"github.com/ovh/cds/sdk/cdsclient/mock_cdsclient"
	"github.com/rockbears/log"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestHatcheryOpenstack_CanSpawn(t *testing.T) {
	h := &HatcheryOpenstack{}
	h.cache = NewCache(1, 1)

	// no model, no requirement, canSpawn must be true
	canSpawn, err := h.CanSpawn(context.TODO(), sdk.WorkerStarterWorkerModel{}, "1", nil)
	require.True(t, canSpawn)
	require.Nil(t, err)

	// no model, service requirement, canSpawn must be false: service can't be managed by openstack hatchery
	canSpawn, err = h.CanSpawn(context.TODO(), sdk.WorkerStarterWorkerModel{}, "1", []sdk.Requirement{{Name: "pg", Type: sdk.ServiceRequirement, Value: "postgres:9.5.4"}})
	require.False(t, canSpawn)
	require.Nil(t, err)

	// no model, memory prerequisite, canSpawn must be false: memory prerequisite can't be managed by openstack hatchery
	canSpawn, err = h.CanSpawn(context.TODO(), sdk.WorkerStarterWorkerModel{}, "1", []sdk.Requirement{{Name: "mem", Type: sdk.MemoryRequirement, Value: "4096"}})
	require.False(t, canSpawn)
	require.Nil(t, err)

	// no model, hostname prerequisite, canSpawn must be false: hostname can't be managed by openstack hatchery
	canSpawn, err = h.CanSpawn(context.TODO(), sdk.WorkerStarterWorkerModel{}, "1", []sdk.Requirement{{Type: sdk.HostnameRequirement, Value: "localhost"}})
	require.False(t, canSpawn)
	require.Nil(t, err)
}

func TestHatcheryOpenstack_WorkerModelsEnabled(t *testing.T) {
	log.Factory = log.NewTestingWrapper(t)

	h := &HatcheryOpenstack{
		Config: HatcheryConfiguration{
			DefaultFlavor: "b2-7",
		},
	}
	h.cache = NewCache(1, 1)

	ctrl := gomock.NewController(t)
	mockClient := mock_cdsclient.NewMockInterface(ctrl)
	h.Client = mockClient
	t.Cleanup(func() { ctrl.Finish() })

	mockClient.EXPECT().WorkerModelEnabledList().DoAndReturn(func() ([]sdk.Model, error) {
		return []sdk.Model{
			{
				ID:    1,
				Type:  sdk.Docker,
				Name:  "my-model-1",
				Group: &sdk.Group{ID: 1, Name: "mygroup"},
			},
			{
				ID:                  2,
				Type:                sdk.Openstack,
				Name:                "my-model-2",
				Group:               &sdk.Group{ID: 1, Name: "mygroup"},
				ModelVirtualMachine: sdk.ModelVirtualMachine{Flavor: "b2-120"},
			},
			{
				ID:                  3,
				Type:                sdk.Openstack,
				Name:                "my-model-3",
				Group:               &sdk.Group{ID: 1, Name: "mygroup"},
				ModelVirtualMachine: sdk.ModelVirtualMachine{Flavor: "b2-7"},
			},
			{
				ID:                  4,
				Type:                sdk.Openstack,
				Name:                "my-model-4",
				Group:               &sdk.Group{ID: 1, Name: "mygroup"},
				ModelVirtualMachine: sdk.ModelVirtualMachine{Flavor: "unknown"},
			},
			{
				ID:                  5,
				Type:                sdk.Openstack,
				Name:                "my-model-5",
				Group:               &sdk.Group{ID: 1, Name: "mygroup"},
				ModelVirtualMachine: sdk.ModelVirtualMachine{Flavor: "d2-2"},
			},
		}, nil
	})

	h.flavors = []flavors.Flavor{
		{Name: "b2-7", VCPUs: 2},
		{Name: "b2-30", VCPUs: 16},
		{Name: "b2-120", VCPUs: 32},
		{Name: "d2-2", VCPUs: 1},
	}

	// Only model that match a known flavor should be returned and sorted by CPUs asc
	ms, err := h.WorkerModelsEnabled()
	require.NoError(t, err)
	require.Len(t, ms, 3)
	assert.Equal(t, "my-model-3", ms[0].Name)
	assert.Equal(t, "my-model-5", ms[1].Name)
	assert.Equal(t, "my-model-2", ms[2].Name)
}

func TestHatcheryOpenstack_checkOverrideImagesUsername(t *testing.T) {
	tests := []struct {
		name      string
		overrides []ImageUsernameOverride
		wantErr   bool
	}{
		{
			name:      "empty",
			overrides: []ImageUsernameOverride{},
		},
		{
			name:      "nil",
			overrides: nil,
		},
		{
			name: "valid-values",
			overrides: []ImageUsernameOverride{
				{
					Image:    "foo",
					Username: "bar",
				},
				{
					Image:    "^foo-[a-z]+",
					Username: "baz123",
				},
				{
					Image:    "^baz$",
					Username: "_foobar",
				},
			},
		},
		{
			name: "invalid-image-regexp",
			overrides: []ImageUsernameOverride{
				{
					Image:    "foo[",
					Username: "bar",
				},
			},
			wantErr: true,
		},
		{
			name: "username-starting-with-dash",
			overrides: []ImageUsernameOverride{
				{
					Image:    "^foo$",
					Username: "-baz",
				},
			},
			wantErr: true,
		},
		{
			name: "username-starting-with-number",
			overrides: []ImageUsernameOverride{
				{
					Image:    "^foo$",
					Username: "1baz",
				},
			},
			wantErr: true,
		},
		{
			name: "username-too-long",
			overrides: []ImageUsernameOverride{
				{
					Image:    "^foo$",
					Username: "abcdefghijklmnopqrstuvwxyz0123456",
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := &HatcheryOpenstack{}
			if err := h.checkOverrideImagesUsername(tt.overrides); (err != nil) != tt.wantErr {
				t.Errorf("HatcheryOpenstack.checkOverrideImagesUsername() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
