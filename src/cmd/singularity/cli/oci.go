// Copyright (c) 2018, Sylabs Inc. All rights reserved.
// This software is licensed under a 3-clause BSD license. Please consult the
// LICENSE.md file distributed with the sources of this project regarding your
// rights to use or distribute this software.

package cli

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/opencontainers/runtime-tools/generate"
	"github.com/spf13/cobra"
	"github.com/sylabs/singularity/src/pkg/buildcfg"
	"github.com/sylabs/singularity/src/pkg/sylog"
	"github.com/sylabs/singularity/src/pkg/util/exec"
	"github.com/sylabs/singularity/src/runtime/engines/config"
	"github.com/sylabs/singularity/src/runtime/engines/oci"
)

var ociName = ""

func init() {
	SingularityCmd.AddCommand(OciCmd)

	OciCreateCmd.Flags().SetInterspersed(false)
	OciCreateCmd.Flags().StringVar(&ociName, "container", "", "specify the OCI name")
	OciCreateCmd.Flags().SetAnnotation("container", "envkey", []string{"CONTAINER"})

	OciStartCmd.Flags().SetInterspersed(false)
	OciStartCmd.Flags().StringVar(&ociName, "container", "", "specify the OCI name")
	OciStartCmd.Flags().SetAnnotation("container", "envkey", []string{"CONTAINER"})

	OciRunCmd.Flags().SetInterspersed(false)
	OciRunCmd.Flags().StringVar(&ociName, "container", "", "specify the OCI name")
	OciRunCmd.Flags().SetAnnotation("container", "envkey", []string{"CONTAINER"})

	OciCmd.AddCommand(OciStartCmd)
	OciCmd.AddCommand(OciCreateCmd)
	OciCmd.AddCommand(OciRunCmd)
}

// OciCreateCmd represents oci create command
var OciCreateCmd = &cobra.Command{
	Args:                  cobra.ExactArgs(0),
	DisableFlagsInUseLine: true,
	Run: func(cmd *cobra.Command, args []string) {
		execOciStarter()
	},
	Use:     "create",
	Short:   "oci create",
	Long:    "oci create",
	Example: "oci create",
}

// OciRunCmd allow to create/start in row
var OciRunCmd = &cobra.Command{
	Args:                  cobra.ExactArgs(0),
	DisableFlagsInUseLine: true,
	Run: func(cmd *cobra.Command, args []string) {
		execOciStarter()
	},
	Use:     "run",
	Short:   "oci run",
	Long:    "oci run",
	Example: "oci run",
}

// OciStartCmd represents oci start command
var OciStartCmd = &cobra.Command{
	Args:                  cobra.ExactArgs(1),
	DisableFlagsInUseLine: true,
	Run: func(cmd *cobra.Command, args []string) {
		execOciStarter()
	},
	Use:     "start",
	Short:   "oci start",
	Long:    "oci start",
	Example: "oci start",
}

// OciCmd singularity oci runtime
var OciCmd = &cobra.Command{
	Run:                   nil,
	DisableFlagsInUseLine: true,

	Use:     "oci",
	Short:   "oci",
	Long:    "oci",
	Example: "oci",
}

func execOciStarter() error {
	starter := buildcfg.LIBEXECDIR + "/singularity/bin/starter"

	os.Clearenv()

	engineConfig := oci.NewConfig()
	generator := generate.Generator{Config: &engineConfig.OciConfig.Spec}
	engineConfig.SetBundlePath("/tmp")

	// load config.json from bundle path
	bundlePath := engineConfig.GetBundlePath()
	configJSON := filepath.Join(bundlePath, "config.json")
	fb, err := os.Open(configJSON)
	if err != nil {
		return fmt.Errorf("failed to open %s: %s", configJSON, err)
	}

	data, err := ioutil.ReadAll(fb)
	if err != nil {
		return fmt.Errorf("failed to read %s: %s", configJSON, err)
	}

	fb.Close()

	if err := json.Unmarshal(data, generator.Config); err != nil {
		return fmt.Errorf("failed to parse %s: %s", configJSON, err)
	}

	os.Setenv("SRUNTIME", "oci")
	os.Setenv("SINGULARITY_MESSAGELEVEL", "0")

	commonConfig := &config.Common{
		ContainerID:  "test",
		EngineName:   "oci",
		EngineConfig: engineConfig,
	}

	configData, err := json.Marshal(commonConfig)
	if err != nil {
		sylog.Fatalf("%s", err)
	}

	cmd, err := exec.PipeCommand(starter, []string{"OCI"}, os.Environ(), configData)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	return cmd.Run()
}
