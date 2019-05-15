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
	ImageTags    AliCloudTagMap
	SnapshotTags AliCloudTagMap
	Ctx          interpolate.Context
}

func (s *stepCreateTags) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	config := state.Get("config").(*Config)
	client := state.Get("client").(*ClientWrapper)
	ui := state.Get("ui").(packer.Ui)
	imageId := state.Get("alicloudimage").(string)
	snapshotIds := state.Get("alicloudsnapshots").([]string)

	if !s.ImageTags.IsSet() && !s.SnapshotTags.IsSet() {
		return multistep.ActionContinue
	} else {
		if s.ImageTags.IsSet() && !s.SnapshotTags.IsSet() {
			s.SnapshotTags = s.ImageTags
		}
		if !s.ImageTags.IsSet() && s.SnapshotTags.IsSet() {
			s.ImageTags = s.SnapshotTags
		}
	}

	// Convert tags to ecs.Tag format
	ui.Say("Creating IMAGE tags")
	imageTags, err := s.ImageTags.ALICLOUDTags(s.Ctx, config.AlicloudRegion, state)
	if err != nil {
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}
	imageTags.Report(ui)

	ui.Say("Creating snapshot tags")
	snapshotTags, err := s.SnapshotTags.ALICLOUDTags(s.Ctx, config.AlicloudRegion, state)
	if err != nil {
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}
	snapshotTags.Report(ui)

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
		ui.Say(fmt.Sprintf("Adding tags(%s) to image: %s", s.ImageTags, imageId))

		addTagsRequest := ecs.CreateAddTagsRequest()
		addTagsRequest.RegionId = config.AlicloudRegion
		addTagsRequest.ResourceId = imageId
		addTagsRequest.ResourceType = TagResourceImage
		addTagsRequest.Tag = &imageTags.ALICLOUDTags
		if _, err := client.AddTags(addTagsRequest); err != nil {
			return err
		}

		// Override tags on snapshots
		for _, snapshotId := range snapshotIds {
			ui.Say(fmt.Sprintf("Adding tags(%s) to snapshot: %s", s.SnapshotTags, snapshotId))

			addTagsRequest := ecs.CreateAddTagsRequest()
			addTagsRequest.RegionId = config.AlicloudRegion
			addTagsRequest.ResourceId = snapshotId
			addTagsRequest.ResourceType = TagResourceSnapshot
			addTagsRequest.Tag = &snapshotTags.ALICLOUDTags

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
