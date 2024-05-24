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
	"CraneFrontEnd/generated/protos"
	"CraneFrontEnd/internal/util"
	"bufio"
	"context"
	"fmt"
	log "github.com/sirupsen/logrus"
	"math"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"
)

var (
	stub protos.CraneCtldClient
)

func ShowNodes(nodeName string, queryAll bool) {
	var req *protos.QueryCranedInfoRequest

	req = &protos.QueryCranedInfoRequest{CranedName: nodeName}
	reply, err := stub.QueryCranedInfo(context.Background(), req)
	if err != nil {
		util.GrpcErrorPrintf(err, "Failed to show nodes")
	}

	var B2MBRatio uint64 = 1024 * 1024

	if queryAll {
		if len(reply.CranedInfoList) == 0 {
			fmt.Printf("No node is avalable.\n")
		} else {
			for _, nodeInfo := range reply.CranedInfoList {
				fmt.Printf("NodeName=%v State=%v CPU=%.2f AllocCPU=%.2f FreeCPU=%.2f\n"+
					"\tRealMemory=%dM AllocMem=%dM FreeMem=%dM\n"+
					"\tPatition=%s RunningJob=%d\n\n",
					nodeInfo.Hostname, nodeInfo.State.String()[6:], nodeInfo.Cpu,
					math.Abs(nodeInfo.AllocCpu),
					math.Abs(nodeInfo.FreeCpu),
					nodeInfo.RealMem/B2MBRatio, nodeInfo.AllocMem/B2MBRatio, nodeInfo.FreeMem/B2MBRatio,
					strings.Join(nodeInfo.PartitionNames, ","), nodeInfo.RunningTaskNum)
			}
		}
	} else {
		if len(reply.CranedInfoList) == 0 {
			fmt.Printf("Node %s not found.\n", nodeName)
		} else {
			for _, nodeInfo := range reply.CranedInfoList {
				fmt.Printf("NodeName=%v State=%v CPU=%.2f AllocCPU=%.2f FreeCPU=%.2f\n"+
					"\tRealMemory=%dM AllocMem=%dM FreeMem=%dM\n"+
					"\tPatition=%s RunningJob=%d\n\n",
					nodeInfo.Hostname, nodeInfo.State.String()[6:], nodeInfo.Cpu, nodeInfo.AllocCpu, nodeInfo.FreeCpu,
					nodeInfo.RealMem/B2MBRatio, nodeInfo.AllocMem/B2MBRatio, nodeInfo.FreeMem/B2MBRatio,
					strings.Join(nodeInfo.PartitionNames, ","), nodeInfo.RunningTaskNum)
			}
		}
	}
}

func ShowPartitions(partitionName string, queryAll bool) {
	var req *protos.QueryPartitionInfoRequest

	req = &protos.QueryPartitionInfoRequest{PartitionName: partitionName}
	reply, err := stub.QueryPartitionInfo(context.Background(), req)
	if err != nil {
		util.GrpcErrorPrintf(err, "Failed to show partition")
	}

	var B2MBRatio uint64 = 1024 * 1024

	if queryAll {
		if len(reply.PartitionInfo) == 0 {
			fmt.Printf("No node is avalable.\n")
		} else {
			for _, partitionInfo := range reply.PartitionInfo {
				fmt.Printf("PartitionName=%v State=%v\n"+
					"\tTotalNodes=%d AliveNodes=%d\n"+
					"\tTotalCPU=%.2f AvailCPU=%.2f AllocCPU=%.2f\n"+
					"\tTotalMem=%dM AvailMem=%dM AllocMem=%dM\n\tHostList=%v\n\n",
					partitionInfo.Name, partitionInfo.State.String()[10:],
					partitionInfo.TotalNodes, partitionInfo.AliveNodes,
					partitionInfo.TotalCpu, partitionInfo.AvailCpu, partitionInfo.AllocCpu,
					partitionInfo.TotalMem/B2MBRatio, partitionInfo.AvailMem/B2MBRatio, partitionInfo.AllocMem/B2MBRatio, partitionInfo.Hostlist)
			}
		}
	} else {
		if len(reply.PartitionInfo) == 0 {
			fmt.Printf("Partition %s not found.\n", partitionName)
		} else {
			for _, partitionInfo := range reply.PartitionInfo {
				fmt.Printf("PartitionName=%v State=%v\n"+
					"\tTotalNodes=%d AliveNodes=%d\n"+
					"\tTotalCPU=%.2f AvailCPU=%.2f AllocCPU=%.2f\n"+
					"\tTotalMem=%dM AvailMem=%dM AllocMem=%dM\n\tHostList=%v\n\n",
					partitionInfo.Name, partitionInfo.State.String()[10:],
					partitionInfo.TotalNodes, partitionInfo.AliveNodes,
					partitionInfo.TotalCpu, partitionInfo.AvailCpu, partitionInfo.AllocCpu,
					partitionInfo.TotalMem/B2MBRatio, partitionInfo.AvailMem/B2MBRatio, partitionInfo.AllocMem/B2MBRatio, partitionInfo.Hostlist)
			}
		}
	}
}

func ShowTasks(taskId uint32, queryAll bool) {
	var req *protos.QueryTasksInfoRequest
	var taskIdList []uint32
	taskIdList = append(taskIdList, taskId)
	if queryAll {
		req = &protos.QueryTasksInfoRequest{}
	} else {
		req = &protos.QueryTasksInfoRequest{FilterTaskIds: taskIdList}
	}

	reply, err := stub.QueryTasksInfo(context.Background(), req)
	if err != nil {
		util.GrpcErrorPrintf(err, "Failed to show the task")
	}

	if !reply.GetOk() {
		log.Errorf("Failed to retrive the information of job %d", taskId)
		os.Exit(1)
	}

	if len(reply.TaskInfoList) == 0 {
		if queryAll {
			fmt.Printf("No job is running.\n")
		} else {
			fmt.Printf("Job %d is not running.\n", taskId)
		}

	} else {
		for _, taskInfo := range reply.TaskInfoList {
			timeSubmit := taskInfo.SubmitTime.AsTime()
			timeSubmitStr := "unknown"
			if !timeSubmit.Before(time.Date(1980, 1, 1, 0, 0, 0, 0, time.UTC)) {
				timeSubmitStr = timeSubmit.Local().String()
			}
			timeStart := taskInfo.StartTime.AsTime()
			timeStartStr := "unknown"
			runTime := "unknown"
			if !timeStart.Before(time.Date(1980, 1, 1, 0, 0, 0, 0, time.UTC)) {
				timeStartStr = timeStart.Local().String()
				runTime = time.Now().Sub(timeStart).String()
			}
			timeEnd := taskInfo.EndTime.AsTime()
			timeEndStr := "unknown"
			if timeEnd.After(timeStart) {
				timeEndStr = timeEnd.Local().String()
				runTime = timeEnd.Sub(timeStart).String()
			}
			if taskInfo.Status.String() == "Running" {
				timeEndStr = timeStart.Add(taskInfo.TimeLimit.AsDuration()).String()
			}
			var timeLimitStr string
			if taskInfo.TimeLimit.Seconds >= util.InvalidDuration().Seconds {
				timeLimitStr = "unlimited"
				timeEndStr = "unknown"
			} else {
				timeLimitStr = util.SecondTimeFormat(taskInfo.TimeLimit.Seconds)
			}

			fmt.Printf("JobId=%v JobName=%v\n\tUserId=%d GroupId=%d Account=%v\n\tJobState=%v RunTime=%v "+
				"TimeLimit=%s SubmitTime=%v\n\tStartTime=%v EndTime=%v Partition=%v NodeList=%v "+
				"NumNodes=%d\n\tCmdLine=%v Workdir=%v\n",
				taskInfo.TaskId, taskInfo.Name, taskInfo.Uid, taskInfo.Gid,
				taskInfo.Account, taskInfo.Status.String(), runTime, timeLimitStr,
				timeSubmitStr, timeStartStr, timeEndStr, taskInfo.Partition,
				taskInfo.CranedList, taskInfo.NodeNum, taskInfo.CmdLine, taskInfo.Cwd)
		}
	}
}

func ChangeTaskTimeLimit(taskId uint32, timeLimit string) {
	re := regexp.MustCompile(`((.*)-)?(.*):(.*):(.*)`)
	result := re.FindAllStringSubmatch(timeLimit, -1)

	if result == nil || len(result) != 1 {
		log.Fatalf("Time format error")
	}
	var dd uint64
	if result[0][2] != "" {
		dd, _ = strconv.ParseUint(result[0][2], 10, 32)
	}
	hh, err := strconv.ParseUint(result[0][3], 10, 32)
	if err != nil {
		log.Fatalf("The hour time format error")
	}
	mm, err := strconv.ParseUint(result[0][4], 10, 32)
	if err != nil {
		log.Fatalf("The minute time format error")
	}
	ss, err := strconv.ParseUint(result[0][5], 10, 32)
	if err != nil {
		log.Fatalf("The second time format error")
	}

	seconds := int64(60*60*24*dd + 60*60*hh + 60*mm + ss)

	var req *protos.ModifyTaskRequest

	req = &protos.ModifyTaskRequest{
		Uid:       uint32(os.Getuid()),
		TaskId:    taskId,
		Attribute: protos.ModifyTaskRequest_TimeLimit,
		Value: &protos.ModifyTaskRequest_TimeLimitSeconds{
			TimeLimitSeconds: seconds,
		},
	}
	reply, err := stub.ModifyTask(context.Background(), req)
	if err != nil {
		util.GrpcErrorPrintf(err, "Failed to change task time limit")
	}

	if reply.Ok {
		fmt.Println("Change time limit success!")
	} else {
		fmt.Printf("Chang time limit failed: %s\n", reply.GetReason())
	}
}

func AddNode(name string, cpu float64, mem string, partition []string) {
	memInByte, err := util.ParseMemStringAsByte(mem)
	if err != nil {
		log.Fatal(err)
	}

	var req *protos.AddNodeRequest
	req = &protos.AddNodeRequest{
		Uid: uint32(os.Getuid()),
		Node: &protos.CranedInfo{
			Hostname:       name,
			Cpu:            cpu,
			RealMem:        memInByte,
			PartitionNames: partition,
		},
	}
	reply, err := stub.AddNode(context.Background(), req)
	if err != nil {
		util.GrpcErrorPrintf(err, "Failed to add a new node")
	}

	if reply.Ok {
		fmt.Println("Add node success!")
	} else {
		fmt.Printf("Add node failed: %s\n", reply.GetReason())
	}
}

func AddPartition(name string, nodes string, priority int64, allowlist []string, denylist []string) {
	var req *protos.AddPartitionRequest
	req = &protos.AddPartitionRequest{
		Uid: uint32(os.Getuid()),
		Partition: &protos.PartitionInfo{
			Name:      name,
			Hostlist:  nodes,
			Priority:  priority,
			AllowList: allowlist,
			DenyList:  denylist,
		},
	}
	reply, err := stub.AddPartition(context.Background(), req)
	if err != nil {
		util.GrpcErrorPrintf(err, "Failed to add a new partition")
	}

	if reply.Ok {
		fmt.Println("Add partition success!")
	} else {
		fmt.Printf("Add partition failed: %s\n", reply.GetReason())
	}
}

func DeleteNode(name string) {
	var queryReq *protos.QueryCranedInfoRequest
	queryReq = &protos.QueryCranedInfoRequest{
		CranedName: name,
	}
	queryReply, err := stub.QueryCranedInfo(context.Background(), queryReq)
	if err != nil {
		util.GrpcErrorPrintf(err, "Failed to query craned info")
	}
	if len(queryReply.CranedInfoList) == 0 {
		fmt.Printf("Node %s not found.\n", name)
	} else {
		if queryReply.CranedInfoList[0].RunningTaskNum > 0 {
			for {
				fmt.Printf("There are still %d jobs running on this node. Are you sure you want to delete this node? (yes/no) ", queryReply.CranedInfoList[0].RunningTaskNum)
				reader := bufio.NewReader(os.Stdin)
				response, err := reader.ReadString('\n')
				if err != nil {
					util.GrpcErrorPrintf(err, "Failed to read from stdin")
				}

				response = strings.TrimSpace(response)
				if strings.ToLower(response) == "no" || strings.ToLower(response) == "n" {
					fmt.Println("Operation canceled.")
					os.Exit(0)
				} else if strings.ToLower(response) == "yes" || strings.ToLower(response) == "y" {
					break
				} else {
					fmt.Println("Unknown reply, please re-enter!")
				}
			}
		}
	}

	var req *protos.DeleteNodeRequest
	req = &protos.DeleteNodeRequest{
		Uid:  uint32(os.Getuid()),
		Name: name,
	}
	reply, err := stub.DeleteNode(context.Background(), req)
	if err != nil {
		util.GrpcErrorPrintf(err, "Failed to delete node")
	}

	if reply.Ok {
		fmt.Println("Delete node success!")
	} else {
		fmt.Printf("Delete node %s failed: %s\n", name, reply.GetReason())
	}
}

func DeletePartition(name string) {
	var req *protos.DeletePartitionRequest
	req = &protos.DeletePartitionRequest{
		Uid:  uint32(os.Getuid()),
		Name: name,
	}
	reply, err := stub.DeletePartition(context.Background(), req)
	if err != nil {
		util.GrpcErrorPrintf(err, "Failed to delete partition!")
	}

	if reply.Ok {
		fmt.Println("Delete partition success!")
	} else {
		fmt.Printf("Delete partition %s failed: %s\n", name, reply.GetReason())
	}
}

func UpdateNode(name string, cpu float64, mem string) {
	memInByte := uint64(0)
	if mem != "" {
		memParsed, err := util.ParseMemStringAsByte(mem)
		if err != nil {
			log.Fatal(err)
		}
		memInByte = memParsed
	}

	var req *protos.UpdateNodeRequest
	req = &protos.UpdateNodeRequest{
		Uid: uint32(os.Getuid()),
		Node: &protos.CranedInfo{
			Hostname: name,
			Cpu:      cpu,
			RealMem:  memInByte,
		},
	}
	reply, err := stub.UpdateNode(context.Background(), req)
	if err != nil {
		util.GrpcErrorPrintf(err, "Failed to update node!")
	}

	if reply.Ok {
		fmt.Println("Update node success!")
	} else {
		fmt.Printf("Update node failed: %s\n", reply.GetReason())
	}
}

func UpdatePartition(name string, nodes string, priority int64, allowlist []string, denylist []string) {
	var req *protos.UpdatePartitionRequest
	req = &protos.UpdatePartitionRequest{
		Uid: uint32(os.Getuid()),
		Partition: &protos.PartitionInfo{
			Name:      name,
			Hostlist:  nodes,
			Priority:  priority,
			AllowList: allowlist,
			DenyList:  denylist,
		},
	}
	reply, err := stub.UpdatePartition(context.Background(), req)
	if err != nil {
		util.GrpcErrorPrintf(err, "Failed to update partition")
	}

	if reply.Ok {
		fmt.Println("Update partition success!")
	} else {
		fmt.Printf("Update partition failed: %s\n", reply.GetReason())
	}
}
