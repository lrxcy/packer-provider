package ecs

import (
	"fmt"

	"github.com/aliyun/alibaba-cloud-sdk-go/services/ecs"
	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"
	"github.com/hashicorp/packer/template/interpolate"
)

type TagData struct {
	ALICLOUDTags []ecs.AddTagsTag
}

type AliCloudTagMap map[string]string

func (t TagData) Report(ui packer.Ui) {
	for _, tag := range t.ALICLOUDTags {
		ui.Message(fmt.Sprintf("Adding tag: \"%s\": \"%s\"",
			StringValue(&tag.Key), StringValue(&tag.Value)))
	}
}

func (t AliCloudTagMap) IsSet() bool {
	return len(t) > 0
}

func (t AliCloudTagMap) ALICLOUDTags(ictx interpolate.Context, region string, state multistep.StateBag) (TagData, error) {
	var aliCloudTags []ecs.AddTagsTag
	ictx.Data = extractBuildInfo(region, state)

	for key, value := range t {
		var tag ecs.AddTagsTag
		interpolatedKey, err := interpolate.Render(key, &ictx)
		if err != nil {
			return TagData{
				ALICLOUDTags: nil,
			}, fmt.Errorf("Error processing tag: %s:%s - %s ", key, value, err)
		}
		interpolatedValue, err := interpolate.Render(value, &ictx)
		if err != nil {
			return TagData{
				ALICLOUDTags: nil,
			}, fmt.Errorf("Error processing tag: %s:%s - %s ", key, value, err)
		}
		tag.Key = interpolatedKey
		tag.Value = interpolatedValue
		aliCloudTags = append(aliCloudTags, tag)
	}
	return TagData{
		ALICLOUDTags: aliCloudTags,
	}, nil
}

// StringValue returns the value of the string pointer passed in or
// "" if the pointer is nil.

func StringValue(v *string) string {
	if v != nil {
		return *v
	}
	return ""
}
