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

import (
	"fmt"

	"github.com/spf13/pflag"
)

var (
	specComposeID       string
	specEnv             string
	specRespin          string
	specImageType       string
	specTimestamp       string
	awsFedoraPartitions = []awsPartitionSpec{
		awsPartitionSpec{
			Name:         "AWS",
			Profile:      "default",
			Bucket:       "fedora-s3-bucket-fedimg",
			BucketRegion: "us-east-1",
			LaunchPermissions: []string{
				"125523088429", // fedora production account
			},
			Regions: []string{
				"ap-northeast-2",
				"us-east-2",
				"ap-southeast-1",
				"ap-southeast-2",
				"ap-south-1",
				"eu-west-1",
				"sa-east-1",
				"us-east-1",
				"us-west-2",
				"us-west-1",
				"eu-central-1",
				"ap-northeast-1",
				"ca-central-1",
				"eu-west-2",
				"eu-west-3",
			},
		},
	}
	awsFedoraUserPartitions = []awsPartitionSpec{
		awsPartitionSpec{
			Name:         "AWS",
			Profile:      "default",
			Bucket:       "fedora-s3-bucket-fedimg-test",
			BucketRegion: "us-east-1",
			LaunchPermissions: []string{
				"013116697141", // fedora community dev test account
			},
			Regions: []string{
				"us-east-2",
				"us-east-1",
			},
		},
	}

	fedoraSpecs = map[string]channelSpec{
		"rawhide": channelSpec{
			BaseURL: "https://koji.fedoraproject.org/compose/rawhide",
			AWS: awsSpec{
				BaseName:        "Fedora",
				BaseDescription: "Fedora AMI",
				Prefix:          "fedora_{{.Env}}_ami_",
				Image:           "Fedora-AtomicHost-{{.Version}}-{{.Timestamp}}.{{.Respin}}.{{.Arch}}.raw.xz",
				Partitions:      awsFedoraPartitions,
			},
		},
		"updates": channelSpec{
			BaseURL: "https://koji.fedoraproject.org/compose/updates",
			AWS: awsSpec{
				BaseName:        "Fedora",
				BaseDescription: "Fedora AMI",
				Prefix:          "fedora_{{.Env}}_ami_",
				Image:           "Fedora-AtomicHost-{{.Version}}-{{.Timestamp}}.{{.Respin}}.{{.Arch}}.raw.xz",
				Partitions:      awsFedoraPartitions,
			},
		},
		"branched": channelSpec{
			BaseURL: "https://koji.fedoraproject.org/compose/branched",
			AWS: awsSpec{
				BaseName:        "Fedora",
				BaseDescription: "Fedora AMI",
				Prefix:          "fedora_{{.Env}}_ami_",
				Image:           "Fedora-{{.ImageType}}-{{.Version}}-{{.Timestamp}}.n.{{.Respin}}.{{.Arch}}.raw.xz",
				Partitions:      awsFedoraPartitions,
			},
		},
		"cloud": channelSpec{
			BaseURL: "https://koji.fedoraproject.org/compose/cloud",
			AWS: awsSpec{
				BaseName:        "Fedora",
				BaseDescription: "Fedora AMI",
				Prefix:          "fedora_{{.Env}}_ami_",
				Image:           "Fedora-{{.ImageType}}-{{.Version}}-{{.Timestamp}}.{{.Respin}}.{{.Arch}}.raw.xz",
				Partitions:      awsFedoraPartitions,
			},
		},
		"user": channelSpec{
			BaseURL: "https://koji.fedoraproject.org/compose/cloud",
			AWS: awsSpec{
				BaseName:        "Fedora",
				BaseDescription: "Fedora AMI",
				Prefix:          "fedora_{{.Env}}_ami_",
				Image:           "Fedora-{{.ImageType}}-{{.Version}}-{{.Timestamp}}.{{.Respin}}.{{.Arch}}.raw.xz",
				Partitions:      awsFedoraUserPartitions,
			},
		},
	}
)

func AddFedoraSpecFlags(flags *pflag.FlagSet) {
	flags.StringVarP(&specEnv, "environment", "E", "prod", "instance environment")
	flags.StringVarP(&specImageType, "imagetype", "I", "Cloud-Base", "type of image")
	flags.StringVarP(&specTimestamp, "timestamp", "T", "", "compose timestamp")
	flags.StringVarP(&specRespin, "respin", "R", "0", "compose respin")
	flags.StringVarP(&specComposeID, "composeid", "O", "", "compose id")
}

func ChannelFedoraSpec() (channelSpec, error) {
	if specComposeID == "" {
		plog.Fatal("--composeid is required")
	}
	if specTimestamp == "" {
		plog.Fatal("--timestamp is required")
	}
	if specVersion == "" {
		plog.Fatal("--version is required")
	}
	if specBoard == "" {
		specBoard = "x86_64"
	}
	spec, ok := fedoraSpecs[specChannel]
	if !ok {
		return channelSpec{}, fmt.Errorf("Unknown channel: %q", specChannel)
	}

	return spec, nil
}
