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
	specEnv             string
	specRespin          string
	specImageType       string
	specTimestamp       string
	specFedoraVersion   string
	specFedoraBoard     string
	awsFedoraBoards     = []string{"amd64-usr"}
	awsFedoraPartitions = []awsPartitionSpec{
		awsPartitionSpec{
			Name:         "AWS",
			Profile:      "default",
			Bucket:       "fedora-s3-bucket-fedimg-test",
			BucketRegion: "us-east-1",
			LaunchPermissions: []string{
				"013116697141",
			},
			Regions: []string{
				"us-east-1",
				"us-east-2",
				"us-west-1",
				"us-west-2",
			},
		},
	}

	fedoraSpecs = map[string]channelSpec{
		"updates": channelSpec{
			BaseURL: "https://koji.fedoraproject.org/",
			System:  "Fedora",
			AWS: awsSpec{
				BaseName:        "Fedora",
				BaseDescription: "Fedora AMI",
				Prefix:          "fedora_{{.Env}}_ami_",
				Image:           "Fedora-AtomicHost-{{.Version}}-{{.Timestamp}}.{{.Respin}}.{{.Arch}}.raw.xz",
				Partitions:      awsFedoraPartitions,
			},
		},
		"twoweek": channelSpec{
			BaseURL: "https://koji.fedoraproject.org/",
			System:  "Fedora",
			AWS: awsSpec{
				BaseName:        "Fedora",
				BaseDescription: "Fedora AMI",
				Prefix:          "fedora_{{.Env}}_ami_",
				Image:           "Fedora-AtomicHost-{{.Version}}-{{.Timestamp}}.{{.Respin}}.{{.Arch}.raw.xz",
				Partitions:      awsFedoraPartitions,
			},
		},
		"version": channelSpec{
			BaseURL: "https://koji.fedoraproject.org/",
			System:  "Fedora",
			AWS: awsSpec{
				BaseName:        "Fedora",
				BaseDescription: "Fedora AMI",
				Prefix:          "fedora_{{.Env}}_ami_",
				Image:           "Fedora-{{.Version}}-{{.Timestamp}}.{{.Arch}}.raw.xz",
				Partitions:      awsFedoraPartitions,
			},
		},
		"branched": channelSpec{
			BaseURL: "https://koji.fedoraproject.org",
			System:  "Fedora",
			AWS: awsSpec{
				BaseName:        "Fedora",
				BaseDescription: "Fedora AMI",
				Prefix:          "fedora_{{.Env}}_ami_",
				Image:           "Fedora-{{.ImageType}}-{{.Version}}-{{.Timestamp}}.n.{{.Respin}}.{{.Arch}}.raw.xz",
				Partitions:      awsFedoraPartitions,
			},
		},
		"cloud": channelSpec{
			BaseURL: "https://koji.fedoraproject.org",
			System:  "Fedora",
			AWS: awsSpec{
				BaseName:        "Fedora",
				BaseDescription: "Fedora AMI",
				Prefix:          "fedora_{{.Env}}_ami_",
				Image:           "Fedora-{{.ImageType}}-{{.Version}}-{{.Timestamp}}.{{.Arch}}.raw.xz",
				Partitions:      awsFedoraPartitions,
			},
		},
	}
)

func AddFedoraSpecFlags(flags *pflag.FlagSet) {
	flags.StringVarP(&specEnv, "environment", "E", "prod", "instance environment")
	flags.StringVarP(&specImageType, "imagetype", "I", "Cloud-Base", "type of image")
	flags.StringVarP(&specFedoraVersion, "fedoraversion", "F", "29", "fedora release version")
	flags.StringVarP(&specTimestamp, "timestamp", "T", "20181101", "compose timestamp")
	flags.StringVarP(&specRespin, "respin", "R", "0", "compose respin")
	flags.StringVarP(&specArch, "arch", "A", "x86_64", "compose arch")
}

func ChannelFedoraSpec() (channelSpec, error) {
	if specChannel == "" {
		return channelSpec{}, fmt.Errorf("Unknown Channel %q", specChannel)
	}

	spec, ok := fedoraSpecs[specChannel]
	if !ok {
		return channelSpec{}, fmt.Errorf("Unknown channel: %q", specChannel)
	}

	return spec, nil
}
