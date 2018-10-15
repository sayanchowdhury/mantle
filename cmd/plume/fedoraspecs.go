// Copyright 2018 Red Hat Inc.
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

package main

type awsPartitionFedoraSpec struct {
	Name              string
	Profile           string
	Bucket            string
	BucketRegion      string
	LaunchPermissions []string
	Regions           []string
}

type awsFedoraSpec struct {
	BaseName        string
	BaseDescription string
	Prefix          string
	Image           string
	Partitions      []awsPartitionFedoraSpec
}

type channelFedoraSpec struct {
	BaseURL string
	AWS     awsFedoraSpec
}

var (
	specFedoraBoard     string
	awsFedoraBoards     = []string{"amd64-usr"}
	awsFedoraPartitions = []awsPartitionFedoraSpec{
		awsPartitionFedoraSpec{
			Name:         "AWS",
			Profile:      "default",
			Bucket:       "fedora-s3-prod-bucket-us-east-1",
			BucketRegion: "us-west-2",
			LaunchPermissions: []string{
				"0123456789",
			},
			Regions: []string{
				"us-east-1",
				"us-east-2",
				"us-west-1",
				"us-west-2",
			},
		},
		awsPartitionFedoraSpec{
			Name:         "AWS GovCloud",
			Profile:      "govcloud",
			Bucket:       "fedora-s3-prod-bucket-us-gov-west-1",
			BucketRegion: "us-gov-west-1",
			Regions: []string{
				"us-gov-west-1",
			},
		},
	}

	fedoraSpecs = map[string]channelFedoraSpec{
		"prod": channelFedoraSpec{
			BaseURL: "https://koji.fedoraproject.org/",
			AWS: awsFedoraSpec{
				BaseName:        "Fedora",
				BaseDescription: "Fedora AMI",
				Prefix:          "fedora_production_ami_",
				Image:           "fedora_production_ami_raw_image.raw.xz",
				Partitions:      awsFedoraPartitions,
			},
		},
	}
)

func ChannelFedoraSpec() channelFedoraSpec {
	if specChannel == "" {
		plog.Fatal("--channel is required")
	}

	spec, ok := fedoraSpecs[specChannel]
	if !ok {
		plog.Fatalf("Unknown channel: %s", specChannel)
	}

	return spec
}
