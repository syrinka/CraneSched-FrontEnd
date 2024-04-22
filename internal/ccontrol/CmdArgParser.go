/**
 * Copyright (c) 2023 Peking University and Peking University
 * Changsha Institute for Computing and Digital Economy
 *
 * CraneSched is licensed under Mulan PSL v2.
 * You can use this software according to the terms and conditions of
 * the Mulan PSL v2.
 * You may obtain a copy of Mulan PSL v2 at:
 *          http://license.coscl.org.cn/MulanPSL2
 * THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS,
 * WITHOUT WARRANTIES OF ANY KIND,
 * EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
 * MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
 * See the Mulan PSL v2 for more details.
 */

package ccontrol

import (
	"CraneFrontEnd/internal/util"
	"github.com/spf13/cobra"
	"os"
	"strconv"
)

var (
	FlagName          string
	FlagCpus          float64
	FlagMem           string
	FlagPartitions    []string
	FlagNodeStr       string
	FlagPriority      uint32
	FlagAllowAccounts []string
	FlagDenyAccounts  []string
	FlagJobId         uint32
	FlagQueryAll      bool
	FlagTimeLimit     string

	FlagConfigFilePath string

	rootCmd = &cobra.Command{
		Use:   "ccontrol",
		Short: "display the state of partitions and nodes",
		Long:  "",
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			config := util.ParseConfig(FlagConfigFilePath)
			stub = util.GetStubToCtldByConfig(config)
		},
	}
	showCmd = &cobra.Command{
		Use:   "show",
		Short: "display state of identified entity, default is all records",
		Long:  "",
	}
	showNodeCmd = &cobra.Command{
		Use:   "node",
		Short: "display state of the specified node, default is all records",
		Long:  "",
		Args:  cobra.MaximumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) == 0 {
				FlagName = ""
				FlagQueryAll = true
			} else {
				FlagName = args[0]
				FlagQueryAll = false
			}
			ShowNodes(FlagName, FlagQueryAll)
		},
	}
	showPartitionCmd = &cobra.Command{
		Use:   "partition",
		Short: "display state of the specified partition, default is all records",
		Long:  "",
		Args:  cobra.MaximumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) == 0 {
				FlagName = ""
				FlagQueryAll = true
			} else {
				FlagName = args[0]
				FlagQueryAll = false
			}
			ShowPartitions(FlagName, FlagQueryAll)
		},
	}
	showTaskCmd = &cobra.Command{
		Use:   "job",
		Short: "display the state of a specified job or all jobs",
		Long:  "",
		Args:  cobra.MaximumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) == 0 {
				FlagQueryAll = true
			} else {
				id, _ := strconv.Atoi(args[0])
				FlagJobId = uint32(id)
				FlagQueryAll = false
			}
			ShowTasks(FlagJobId, FlagQueryAll)
		},
	}
	addCmd = &cobra.Command{
		Use:   "add",
		Short: "Add a partition or node",
		Long:  "",
	}
	addNodeCmd = &cobra.Command{
		Use:   "node",
		Short: "Add a new node",
		Long:  "",
		Run: func(cmd *cobra.Command, args []string) {
			AddNode(FlagName, FlagCpus, FlagMem, FlagPartitions)
		},
	}
	addPartitionCmd = &cobra.Command{
		Use:   "partition",
		Short: "Add a new partition",
		Long:  "",
		Run: func(cmd *cobra.Command, args []string) {
			AddPartition(FlagName, FlagNodeStr, FlagPriority, FlagAllowAccounts, FlagDenyAccounts)
		},
	}
	deleteCmd = &cobra.Command{
		Use:     "delete",
		Aliases: []string{"remove"},
		Short:   "Delete a partition or node",
		Long:    "",
	}
	deleteNodeCmd = &cobra.Command{
		Use:   "node",
		Short: "Delete an existing node",
		Long:  "",
		Run: func(cmd *cobra.Command, args []string) {
			DeleteNode(FlagName)
		},
	}
	deletePartitionCmd = &cobra.Command{
		Use:   "partition",
		Short: "Delete an existing partition",
		Long:  "",
		Run: func(cmd *cobra.Command, args []string) {
			DeletePartition(FlagName)
		},
	}
	updateCmd = &cobra.Command{
		Use:   "update",
		Short: "Modify job, partition, and node info",
		Long:  "",
	}
	updateJobCmd = &cobra.Command{
		Use:   "job",
		Short: "Modify job information",
		Long:  "",
		Run: func(cmd *cobra.Command, args []string) {
			ChangeTaskTimeLimit(FlagJobId, FlagTimeLimit)
		},
	}
	updateNodeCmd = &cobra.Command{
		Use:   "node",
		Short: "Modify node information",
		Long:  "",
		Run: func(cmd *cobra.Command, args []string) {
			UpdateNode(FlagName, FlagCpus, FlagMem)
		},
	}
	updatePartitionCmd = &cobra.Command{
		Use:   "partition",
		Short: "Modify partition information",
		Long:  "",
		Run: func(cmd *cobra.Command, args []string) {
			UpdatePartition(FlagName, FlagNodeStr, FlagPriority, FlagAllowAccounts, FlagDenyAccounts)
		},
	}
)

// ParseCmdArgs executes the root command.
func ParseCmdArgs() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.AddCommand(showCmd)
	rootCmd.PersistentFlags().StringVarP(&FlagConfigFilePath, "config", "C", util.DefaultConfigPath,
		"Path to configuration file")
	showCmd.AddCommand(showNodeCmd)
	showCmd.AddCommand(showPartitionCmd)
	showCmd.AddCommand(showTaskCmd)

	rootCmd.AddCommand(addCmd)
	addCmd.AddCommand(addNodeCmd)
	addNodeCmd.Flags().StringVarP(&FlagName, "name", "N", "", "Host name")
	addNodeCmd.Flags().Float64Var(&FlagCpus, "cpu", 0.0, "Number of CPU cores")
	addNodeCmd.Flags().StringVarP(&FlagMem, "memory", "M", "", "Memory size, in units of G/M/K/B")
	addNodeCmd.Flags().StringSliceVarP(&FlagPartitions, "partition", "P", nil, "The partition name to which the node belongs")
	err := addNodeCmd.MarkFlagRequired("name")
	if err != nil {
		return
	}
	err = addNodeCmd.MarkFlagRequired("cpu")
	if err != nil {
		return
	}
	err = addNodeCmd.MarkFlagRequired("memory")
	if err != nil {
		return
	}

	addCmd.AddCommand(addPartitionCmd)
	addPartitionCmd.Flags().StringVarP(&FlagName, "name", "N", "", "Partition name")
	addPartitionCmd.Flags().StringVar(&FlagNodeStr, "nodes", "", "The included nodes can be written individually, abbreviated or mixed, please write them in a string")
	addPartitionCmd.Flags().Uint32Var(&FlagPriority, "priority", 0, "Partition priority")
	addPartitionCmd.Flags().StringSliceVar(&FlagAllowAccounts, "allowlist", nil, "List of accounts allowed to use this partition")
	addPartitionCmd.Flags().StringSliceVar(&FlagDenyAccounts, "denylist", nil, "Prohibit the use of the account list in this partition. The --denylist and the --allowlist parameter can only be selected as either")
	addPartitionCmd.MarkFlagsMutuallyExclusive("allowlist", "denylist")

	rootCmd.AddCommand(deleteCmd)
	deleteCmd.AddCommand(deleteNodeCmd)
	deleteNodeCmd.Flags().StringVarP(&FlagName, "name", "N", "", "Host name")

	deleteCmd.AddCommand(deletePartitionCmd)
	deletePartitionCmd.Flags().StringVarP(&FlagName, "name", "N", "", "Partition name")

	rootCmd.AddCommand(updateCmd)
	updateCmd.AddCommand(updateNodeCmd)

	updateNodeCmd.Flags().StringVarP(&FlagName, "name", "N", "", "Host name")
	updateNodeCmd.Flags().Float64Var(&FlagCpus, "cpu", 0.0, "Number of CPU cores")
	updateNodeCmd.Flags().StringVarP(&FlagMem, "memory", "M", "", "Memory size, in units of G/M/K/B")
	//updateNodeCmd.Flags().StringSliceVarP(&FlagPartitions, "partition", "P", nil, "The partition name to which the node belongs")
	err = updateNodeCmd.MarkFlagRequired("name")
	if err != nil {
		return
	}

	updateCmd.AddCommand(updatePartitionCmd)
	updatePartitionCmd.Flags().StringVarP(&FlagName, "name", "N", "", "Partition name")
	updatePartitionCmd.Flags().StringVar(&FlagNodeStr, "nodes", "", "The included nodes can be written individually, abbreviated or mixed, please write them in a string")
	updatePartitionCmd.Flags().Uint32Var(&FlagPriority, "priority", 0, "Partition priority")
	updatePartitionCmd.Flags().StringSliceVar(&FlagAllowAccounts, "allowlist", nil, "List of accounts allowed to use this partition")
	updatePartitionCmd.Flags().StringSliceVar(&FlagDenyAccounts, "denylist", nil, "Prohibit the use of the account list in this partition. The --denylist and the --allowlist parameter can only be selected as either")
	updatePartitionCmd.MarkFlagsMutuallyExclusive("allowlist", "denylist")

	updateCmd.AddCommand(updateJobCmd)
	updateJobCmd.Flags().Uint32VarP(&FlagJobId, "job", "J", 0, "Job id")
	updateJobCmd.Flags().StringVarP(&FlagTimeLimit, "time-limit", "T", "", "time limit")
	err = updateJobCmd.MarkFlagRequired("job")
	if err != nil {
		return
	}
}
