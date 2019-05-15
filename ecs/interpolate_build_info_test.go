package ecs

import (
	"reflect"
	"testing"

	"github.com/aliyun/alibaba-cloud-sdk-go/services/ecs"
	"github.com/hashicorp/packer/helper/multistep"
)

func testImage() *ecs.Image {
	return &ecs.Image{
		ImageId:   "packer-abcd1234",
		ImageName: "packer_test_name",
		Tags: ecs.TagsInDescribeImages{
			Tag: []ecs.Tag{
				{
					TagKey:   "key-1",
					TagValue: "value-1",
				},
				{
					TagKey:   "key-2",
					TagValue: "value-2",
				},
			},
		},
	}
}

func testState() multistep.StateBag {
	state := new(multistep.BasicStateBag)
	return state
}

func TestInterpolateBuildInfo_extractBuildInfo_noSourceImage(t *testing.T) {
	state := testState()
	buildInfo := extractBuildInfo("foo", state)

	expected := BuildInfoTemplate{
		BuildRegion: "foo",
	}
	if !reflect.DeepEqual(*buildInfo, expected) {
		t.Fatalf("Unexpected BuildInfoTemplate: expected %#v got %#v\n", expected, *buildInfo)
	}
}

func TestInterpolateBuildInfo_extractBuildInfo_withSourceImage(t *testing.T) {
	state := testState()
	state.Put("source_image", testImage())
	buildInfo := extractBuildInfo("foo", state)

	expected := BuildInfoTemplate{
		BuildRegion:     "foo",
		SourceIMAGE:     "packer-abcd1234",
		SourceIMAGEName: "packer_test_name",
		SourceIMAGETags: map[string]string{
			"key-1": "value-1",
			"key-2": "value-2",
		},
	}
	if !reflect.DeepEqual(*buildInfo, expected) {
		t.Fatalf("Unexpected BuildInfoTemplate: expected %#v got %#v\n", expected, *buildInfo)
	}
}
