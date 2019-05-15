package ecs

import (
	"github.com/aliyun/alibaba-cloud-sdk-go/services/ecs"
	"github.com/hashicorp/packer/helper/multistep"
)

type BuildInfoTemplate struct {
	BuildRegion     string
	SourceIMAGE     string
	SourceIMAGEName string
	SourceIMAGETags map[string]string
}

func extractBuildInfo(region string, state multistep.StateBag) *BuildInfoTemplate {
	rawSourceIMAGE, hasSourceIMAGE := state.GetOk("source_image")
	if !hasSourceIMAGE {
		return &BuildInfoTemplate{
			BuildRegion: region,
		}
	}

	sourceIMAGE := rawSourceIMAGE.(*ecs.Image)
	sourceIMAGETags := make(map[string]string, len(sourceIMAGE.Tags.Tag))
	for _, tag := range sourceIMAGE.Tags.Tag {
		sourceIMAGETags[StringValue(&tag.TagKey)] = StringValue(&tag.TagValue)
	}

	return &BuildInfoTemplate{
		BuildRegion:     region,
		SourceIMAGE:     StringValue(&sourceIMAGE.ImageId),
		SourceIMAGEName: StringValue(&sourceIMAGE.ImageName),
		SourceIMAGETags: sourceIMAGETags,
	}
}
