package ecs

import (
	"context"
	"fmt"
	"time"

	"github.com/aliyun/alibaba-cloud-sdk-go/sdk/errors"
	"github.com/aliyun/alibaba-cloud-sdk-go/services/ecs"
	"github.com/hashicorp/packer/common/retry"
	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"
	"github.com/hashicorp/packer/template/interpolate"
)

type stepCreateTags struct {
	Tags AliCloudTagMap
	Ctx  interpolate.Context
}

func (s *stepCreateTags) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	config := state.Get("config").(*Config)
	client := state.Get("client").(*ClientWrapper)
	ui := state.Get("ui").(packer.Ui)
	imageId := state.Get("alicloudimage").(string)
	snapshotIds := state.Get("alicloudsnapshots").([]string)

	if !s.Tags.IsSet() {
		return multistep.ActionContinue
	}

	// Convert tags to ecs.Tag format
	ui.Say("Creating IMAGE tags")
	tags, err := s.Tags.ALICLOUDTags(s.Ctx, config.AlicloudRegion, state)
	if err != nil {
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}
	tags.Report(ui)

	// Retry creating tags for about 2.5 minutes
	err = retry.Config{
		Tries: 11,
		ShouldRetry: func(error) bool {
			if ecsErr, ok := err.(errors.Error); ok {
				switch ecsErr.ErrorCode() {
				case "InvalidResourceId.NotFound":
					return true
				}
			}
			return false
		},
		RetryDelay: (&retry.Backoff{InitialBackoff: 200 * time.Millisecond, MaxBackoff: 30, Multiplier: 2}).Linear,
	}.Run(ctx, func(ctx context.Context) error {
		// Tag images and snapshots
		ui.Say(fmt.Sprintf("Adding tags(%s) to image: %s", s.Tags, imageId))

		addTagsRequest := ecs.CreateAddTagsRequest()
		addTagsRequest.RegionId = config.AlicloudRegion
		addTagsRequest.ResourceId = imageId
		addTagsRequest.ResourceType = TagResourceImage
		addTagsRequest.Tag = &tags.ALICLOUDTags
		if _, err := client.AddTags(addTagsRequest); err != nil {
			return err
		}

		// Override tags on snapshots
		for _, snapshotId := range snapshotIds {
			ui.Say(fmt.Sprintf("Adding tags(%s) to snapshot: %s", s.Tags, snapshotId))

			addTagsRequest := ecs.CreateAddTagsRequest()
			addTagsRequest.RegionId = config.AlicloudRegion
			addTagsRequest.ResourceId = snapshotId
			addTagsRequest.ResourceType = TagResourceSnapshot
			addTagsRequest.Tag = &tags.ALICLOUDTags

			if _, err := client.AddTags(addTagsRequest); err != nil {
				return err
			}
		}
		return nil
	})

	if err != nil {
		err := fmt.Errorf("Error adding tags to Resources (%#v): %s ", imageId, err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	return multistep.ActionContinue
}
func (s *stepCreateTags) Cleanup(state multistep.StateBag) {
	// Nothing need to do, tags will be cleaned when the resource is cleaned
}
