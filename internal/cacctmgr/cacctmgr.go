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

package cacctmgr

import (
	"CraneFrontEnd/generated/protos"
	"CraneFrontEnd/internal/util"
	"context"
	"fmt"
	"math"
	"os"
	OSUser "os/user"
	"sort"
	"strconv"
	"strings"

	"github.com/olekukonko/tablewriter"
	log "github.com/sirupsen/logrus"
	"github.com/xlab/treeprint"
)

var (
	userUid uint32
	stub    protos.CraneCtldClient
)

type ServerAddr struct {
	ControlMachine      string `yaml:"ControlMachine"`
	CraneCtldListenPort string `yaml:"CraneCtldListenPort"`
}

func PrintAllUsers(userList []*protos.UserInfo) {
	if len(userList) == 0 {
		fmt.Println("There is no user in crane")
		return
	}
	sort.Slice(userList, func(i, j int) bool {
		return userList[i].Uid < userList[j].Uid
	})

	//slice to map
	userMap := make(map[string][]*protos.UserInfo)
	for _, userInfo := range userList {
		key := ""
		if userInfo.Account[len(userInfo.Account)-1] == '*' {
			key = userInfo.Account[:len(userInfo.Account)-1]
		} else {
			key = userInfo.Account
		}
		if list, ok := userMap[key]; ok {
			userMap[key] = append(list, userInfo)
		} else {
			var list = []*protos.UserInfo{userInfo}
			userMap[key] = list
		}
	}

	table := tablewriter.NewWriter(os.Stdout) //table format control
	table.SetBorders(tablewriter.Border{Left: true, Top: true, Right: true, Bottom: true})
	table.SetCenterSeparator("|")
	table.SetTablePadding("\t")
	table.SetHeader([]string{"Account", "UserName", "Uid", "AllowedPartition", "AllowedQosList", "DefaultQos", "AdminLevel", "blocked"})
	table.SetAutoFormatHeaders(false)
	table.SetAlignment(tablewriter.ALIGN_LEFT)
	tableData := make([][]string, len(userMap))

	for key, value := range userMap {
		tableData = append(tableData, []string{key})
		for _, userInfo := range value {
			if len(userInfo.AllowedPartitionQosList) == 0 {
				tableData = append(tableData, []string{
					userInfo.Account,
					userInfo.Name,
					strconv.FormatUint(uint64(userInfo.Uid), 10),
					"",
					"",
					"",
					fmt.Sprintf("%v", userInfo.AdminLevel)})
			}
			for _, allowedPartitionQos := range userInfo.AllowedPartitionQosList {
				tableData = append(tableData, []string{
					userInfo.Account,
					userInfo.Name,
					strconv.FormatUint(uint64(userInfo.Uid), 10),
					allowedPartitionQos.PartitionName,
					strings.Join(allowedPartitionQos.QosList, ", "),
					allowedPartitionQos.DefaultQos,
					fmt.Sprintf("%v", userInfo.AdminLevel),
					strconv.FormatBool(userInfo.Blocked)})
			}
		}
	}

	table.AppendBulk(tableData)
	table.Render()
}

func PrintAllQos(qosList []*protos.QosInfo) {
	if len(qosList) == 0 {
		fmt.Println("There is no qos in crane")
		return
	}
	sort.Slice(qosList, func(i, j int) bool {
		return qosList[i].Name < qosList[j].Name
	})

	table := tablewriter.NewWriter(os.Stdout) //table format control
	table.SetBorders(tablewriter.Border{Left: true, Top: true, Right: true, Bottom: true})
	table.SetCenterSeparator("|")
	table.SetTablePadding("\t")
	table.SetHeader([]string{"Name", "Description", "Priority", "MaxJobsPerUser", "MaxCpusPerUser", "MaxTimeLimitPerTask"})
	table.SetAutoFormatHeaders(false)
	table.SetAlignment(tablewriter.ALIGN_LEFT)
	tableData := make([][]string, len(qosList))
	for _, info := range qosList {
		var timeLimitStr string
		if info.MaxTimeLimitPerTask >= uint64(util.InvalidDuration().Seconds) {
			timeLimitStr = "unlimited"
		} else {
			timeLimitStr = util.SecondTimeFormat(int64(info.MaxTimeLimitPerTask))
		}
		var jobsPerUserStr string
		if info.MaxJobsPerUser == math.MaxUint32 {
			jobsPerUserStr = "unlimited"
		} else {
			jobsPerUserStr = strconv.FormatUint(uint64(info.MaxJobsPerUser), 10)
		}
		var cpusPerUserStr string
		if info.MaxCpusPerUser == math.MaxUint32 {
			cpusPerUserStr = "unlimited"
		} else {
			cpusPerUserStr = strconv.FormatUint(uint64(info.MaxCpusPerUser), 10)
		}
		tableData = append(tableData, []string{
			info.Name,
			info.Description,
			fmt.Sprint(info.Priority),
			fmt.Sprint(jobsPerUserStr),
			fmt.Sprint(cpusPerUserStr),
			fmt.Sprint(timeLimitStr)})
	}

	table.AppendBulk(tableData)
	table.Render()
}

func PrintAllAccount(accountList []*protos.AccountInfo) {
	if len(accountList) == 0 {
		fmt.Println("There is no account in crane")
		return
	}
	sort.Slice(accountList, func(i, j int) bool {
		return accountList[i].Name < accountList[j].Name
	})
	//slice to map and find the root account
	accountMap := make(map[string]*protos.AccountInfo)
	rootAccount := make([]string, 0)
	for _, accountInfo := range accountList {
		accountMap[accountInfo.Name] = accountInfo
	}
	for _, accountInfo := range accountList {
		if accountInfo.ParentAccount == "" || func() bool {
			_, ok := accountMap[accountInfo.ParentAccount]
			return !ok
		}() {
			rootAccount = append(rootAccount, accountInfo.Name)
		}
	}

	//print account tree
	tree := treeprint.NewWithRoot("AccountTree")
	for _, account := range rootAccount {
		PraseAccountTree(tree, account, accountMap)
	}

	fmt.Println(tree.String())

	//print account table
	PrintAccountTable(accountList)
}

func PrintAccountTable(accountList []*protos.AccountInfo) {
	table := tablewriter.NewWriter(os.Stdout) //table format control
	util.SetBorderTable(table)
	header := []string{"Name", "Description", "AllowedPartition", "Users", "DefaultQos", "AllowedQosList", "blocked"}
	tableData := make([][]string, len(accountList))
	for _, accountInfo := range accountList {
		tableData = append(tableData, []string{
			accountInfo.Name,
			accountInfo.Description,
			strings.Join(accountInfo.AllowedPartitions, ", "),
			strings.Join(accountInfo.Users, ", "),
			accountInfo.DefaultQos,
			strings.Join(accountInfo.AllowedQosList, ", "),
			strconv.FormatBool(accountInfo.Blocked)})
	}

	if FlagFormat != "" {
		formatTableData := make([][]string, len(accountList))
		formatReq := strings.Split(FlagFormat, " ")
		tableOutputWidth := make([]int, len(formatReq))
		tableOutputHeader := make([]string, len(formatReq))
		for i := 0; i < len(formatReq); i++ {
			if formatReq[i][0] != '%' || len(formatReq[i]) < 2 {
				fmt.Println("Invalid format.")
				os.Exit(util.ErrorInvalidTableFormat)
			}
			if formatReq[i][1] == '.' {
				if len(formatReq[i]) < 4 {
					fmt.Println("Invalid format.")
					os.Exit(util.ErrorInvalidTableFormat)
				}
				width, err := strconv.ParseUint(formatReq[i][2:len(formatReq[i])-1], 10, 32)
				if err != nil {
					fmt.Println("Invalid format.")
					os.Exit(util.ErrorInvalidTableFormat)
				}
				tableOutputWidth[i] = int(width)
			} else {
				tableOutputWidth[i] = -1
			}
			tableOutputHeader[i] = formatReq[i][len(formatReq[i])-1:]
			//"Name", "Description", "AllowedPartition", "DefaultQos", "AllowedQosList"
			switch tableOutputHeader[i] {
			case "n":
				tableOutputHeader[i] = "Name"
				for j := 0; j < len(accountList); j++ {
					formatTableData[j] = append(formatTableData[j], accountList[j].Name)
				}
			case "d":
				tableOutputHeader[i] = "Description"
				for j := 0; j < len(accountList); j++ {
					formatTableData[j] = append(formatTableData[j], accountList[j].Description)
				}
			case "P":
				tableOutputHeader[i] = "AllowedPartition"
				for j := 0; j < len(accountList); j++ {
					formatTableData[j] = append(formatTableData[j], strings.Join(accountList[j].AllowedPartitions, ", "))
				}
			case "Q":
				tableOutputHeader[i] = "DefaultQos"
				for j := 0; j < len(accountList); j++ {
					formatTableData[j] = append(formatTableData[j], accountList[j].DefaultQos)
				}
			case "q":
				tableOutputHeader[i] = "AllowedQosList"
				for j := 0; j < len(accountList); j++ {
					formatTableData[j] = append(formatTableData[j], strings.Join(accountList[j].AllowedQosList, ", "))
				}
			default:
				fmt.Println("Invalid format.")
				os.Exit(util.ErrorInvalidTableFormat)
			}
		}
		header, tableData = util.FormatTable(tableOutputWidth, tableOutputHeader, formatTableData)
	}

	if !FlagNoHeader {
		table.SetHeader(header)
	}
	table.AppendBulk(tableData)
	table.Render()
}

func PraseAccountTree(parentTreeRoot treeprint.Tree, account string, accountMap map[string]*protos.AccountInfo) {
	if account == "" {
		return
	}
	if len(accountMap[account].ChildAccounts) == 0 {
		parentTreeRoot.AddNode(account)
	} else {
		branch := parentTreeRoot.AddBranch(account)
		for _, child := range accountMap[account].ChildAccounts {
			PraseAccountTree(branch, child, accountMap)
		}
	}
}

func AddAccount(account *protos.AccountInfo) util.CraneCmdError {
	if account.Name == "=" {
		log.Errorf("Parameter error : account name empty")
		return util.ErrorCmdArg
	}
	if len(account.Name) > 30 {
		log.Errorf("Parameter error : name is too long(up to 30)")
		return util.ErrorCmdArg
	}

	req := new(protos.AddAccountRequest)
	req.Uid = userUid
	req.Account = account
	if account.DefaultQos == "" && len(account.AllowedQosList) > 0 {
		account.DefaultQos = account.AllowedQosList[0]
	}
	if account.DefaultQos != "" {
		find := false
		for _, qos := range account.AllowedQosList {
			if qos == account.DefaultQos {
				find = true
				break
			}
		}
		if !find {
			log.Errorf("Parameter error : default qos %s not contain in allowed qos list", account.DefaultQos)
			return util.ErrorCmdArg
		}
	}

	reply, err := stub.AddAccount(context.Background(), req)
	if err != nil {
		util.GrpcErrorPrintf(err, "Failed to add the account")
		return util.ErrorGrpc
	}
	if reply.GetOk() {
		fmt.Println("Add account success!")
		return util.ErrorSuccess
	} else {
		fmt.Printf("Add account failed: %s\n", reply.GetReason())
		return util.ErrorBackEnd
	}
}

func AddUser(user *protos.UserInfo, partition []string, level string, coordinate bool) util.CraneCmdError {
	if user.Name == "=" {
		log.Errorf("Parameter error : user name empty")
		return util.ErrorCmdArg
	}
	if len(user.Name) > 30 {
		log.Errorf("Parameter error : name is too long(up to 30)")
		return util.ErrorCmdArg
	}

	lu, err := OSUser.Lookup(user.Name)
	if err != nil {
		log.Error(err)
		return util.ErrorCacctmgrUserNotFound
	}

	req := new(protos.AddUserRequest)
	req.Uid = userUid
	req.User = user
	for _, par := range partition {
		user.AllowedPartitionQosList = append(user.AllowedPartitionQosList, &protos.UserInfo_AllowedPartitionQos{PartitionName: par})
	}

	i64, err := strconv.ParseInt(lu.Uid, 10, 64)
	if err == nil {
		user.Uid = uint32(i64)
	}

	if level == "none" {
		user.AdminLevel = protos.UserInfo_None
	} else if level == "operator" {
		user.AdminLevel = protos.UserInfo_Operator
	} else if level == "admin" {
		user.AdminLevel = protos.UserInfo_Admin
	}

	if coordinate {
		user.CoordinatorAccounts = append(user.CoordinatorAccounts, user.Account)
	}

	reply, err := stub.AddUser(context.Background(), req)
	if err != nil {
		util.GrpcErrorPrintf(err, "Failed to add the user")
		return util.ErrorGrpc
	}
	if reply.GetOk() {
		fmt.Println("Add user success!")
		return util.ErrorSuccess
	} else {
		fmt.Printf("Add user failed: %s\n", reply.GetReason())
		return util.ErrorBackEnd
	}
}

func AddQos(qos *protos.QosInfo) util.CraneCmdError {
	if qos.Name == "=" {
		log.Errorf("Parameter error : QOS name empty")
		return util.ErrorCmdArg
	}
	if len(qos.Name) > 30 {
		log.Errorf("Parameter error : name is too long(up to 30)")
		return util.ErrorCmdArg
	}

	req := new(protos.AddQosRequest)
	req.Uid = userUid
	req.Qos = qos

	reply, err := stub.AddQos(context.Background(), req)
	if err != nil {
		util.GrpcErrorPrintf(err, "Failed to add the QoS")
		return util.ErrorGrpc
	}
	if reply.GetOk() {
		fmt.Println("Add qos success!")
		return util.ErrorSuccess
	} else {
		fmt.Printf("Add qos failed: %s\n", reply.GetReason())
		return util.ErrorBackEnd
	}
}

func DeleteAccount(name string) util.CraneCmdError {
	req := protos.DeleteEntityRequest{Uid: userUid, EntityType: protos.EntityType_Account, Name: name}

	reply, err := stub.DeleteEntity(context.Background(), &req)
	if err != nil {
		util.GrpcErrorPrintf(err, "Failed to delete account %s", name)
		return util.ErrorGrpc
	}
	if reply.GetOk() {
		fmt.Printf("Delete account %s success\n", name)
		return util.ErrorSuccess
	} else {
		fmt.Printf("Delete account %s failed: %s\n", name, reply.GetReason())
		return util.ErrorBackEnd
	}
}

func DeleteUser(name string, account string) util.CraneCmdError {
	req := protos.DeleteEntityRequest{Uid: userUid, EntityType: protos.EntityType_User, Name: name, Account: account}

	reply, err := stub.DeleteEntity(context.Background(), &req)
	if err != nil {
		util.GrpcErrorPrintf(err, "Failed to remove user %s", name)
		return util.ErrorGrpc
	}
	if reply.GetOk() {
		fmt.Printf("Remove User %s success\n", name)
		return util.ErrorSuccess
	} else {
		fmt.Printf("Remove User %s failed: %s\n", name, reply.GetReason())
		return util.ErrorBackEnd
	}
}

func DeleteQos(name string) util.CraneCmdError {
	req := protos.DeleteEntityRequest{Uid: userUid, EntityType: protos.EntityType_Qos, Name: name}

	reply, err := stub.DeleteEntity(context.Background(), &req)
	if err != nil {
		util.GrpcErrorPrintf(err, "Failed to delete QoS %s", name)
		return util.ErrorGrpc
	}
	if reply.GetOk() {
		fmt.Printf("Delete Qos %s success\n", name)
		return util.ErrorSuccess
	} else {
		fmt.Printf("Delete Qos %s failed: %s\n", name, reply.GetReason())
		return util.ErrorBackEnd
	}
}

func ModifyAccount(itemLeft string, itemRight string, name string, requestType protos.ModifyEntityRequest_OperatorType) util.CraneCmdError {
	req := protos.ModifyEntityRequest{
		Uid:        userUid,
		Item:       itemLeft,
		Value:      itemRight,
		Name:       name,
		Type:       requestType,
		EntityType: protos.EntityType_Account,
		Force:      FlagForce,
	}

	reply, err := stub.ModifyEntity(context.Background(), &req)
	if err != nil {
		util.GrpcErrorPrintf(err, "Modify information")
		return util.ErrorGrpc
	}
	if reply.GetOk() {
		fmt.Println("Modify information success!")
		return util.ErrorSuccess
	} else {
		fmt.Printf("Modify information failed: %s\n", reply.GetReason())
		return util.ErrorBackEnd
	}
}

func ModifyUser(itemLeft string, itemRight string, name string, account string, partition string, requestType protos.ModifyEntityRequest_OperatorType) util.CraneCmdError {
	if itemLeft == "admin_level" {
		if itemRight != "none" && itemRight != "operator" && itemRight != "admin" {
			log.Errorf("Unknown admin_level, please enter one of {none, operator, admin}")
			return util.ErrorCmdArg
		}
	}

	req := protos.ModifyEntityRequest{
		Uid:        userUid,
		Item:       itemLeft,
		Value:      itemRight,
		Name:       name,
		Partition:  partition,
		Type:       requestType,
		EntityType: protos.EntityType_User,
		Account:    account,
		Force:      FlagForce,
	}

	reply, err := stub.ModifyEntity(context.Background(), &req)
	if err != nil {
		util.GrpcErrorPrintf(err, "Failed to modify the uesr information")
		return util.ErrorGrpc
	}
	if reply.GetOk() {
		fmt.Println("Modify information success!")
		return util.ErrorSuccess
	} else {
		fmt.Printf("Modify information failed: %s\n", reply.GetReason())
		return util.ErrorBackEnd
	}
}

func ModifyQos(itemLeft string, itemRight string, name string) util.CraneCmdError {
	req := protos.ModifyEntityRequest{
		Uid:        userUid,
		Item:       itemLeft,
		Value:      itemRight,
		Name:       name,
		Type:       protos.ModifyEntityRequest_Overwrite,
		EntityType: protos.EntityType_Qos,
	}

	reply, err := stub.ModifyEntity(context.Background(), &req)
	if err != nil {
		util.GrpcErrorPrintf(err, "Failed to modify the QoS")
		return util.ErrorGrpc
	}
	if reply.GetOk() {
		fmt.Println("Modify information success!")
		return util.ErrorSuccess
	} else {
		fmt.Printf("Modify information failed: %s\n", reply.GetReason())
		return util.ErrorBackEnd
	}
}

func ShowAccounts() util.CraneCmdError {
	req := protos.QueryEntityInfoRequest{Uid: userUid, EntityType: protos.EntityType_Account}
	reply, err := stub.QueryEntityInfo(context.Background(), &req)
	if err != nil {
		util.GrpcErrorPrintf(err, "Failed to show accounts")
		return util.ErrorGrpc
	}

	if reply.GetOk() {
		PrintAllAccount(reply.AccountList)
		return util.ErrorSuccess
	} else {
		fmt.Println(reply.Reason)
		return util.ErrorBackEnd
	}
}

func ShowUser(name string, account string) util.CraneCmdError {
	req := protos.QueryEntityInfoRequest{Uid: userUid, EntityType: protos.EntityType_User, Name: name, Account: account}
	reply, err := stub.QueryEntityInfo(context.Background(), &req)
	if err != nil {
		util.GrpcErrorPrintf(err, "Failed to show the user")
		return util.ErrorGrpc
	}

	if reply.GetOk() {
		PrintAllUsers(reply.UserList)
		return util.ErrorSuccess
	} else {
		fmt.Println(reply.Reason)
		return util.ErrorBackEnd
	}
}

func ShowQos(name string) util.CraneCmdError {
	req := protos.QueryEntityInfoRequest{Uid: userUid, EntityType: protos.EntityType_Qos, Name: name}
	reply, err := stub.QueryEntityInfo(context.Background(), &req)
	if err != nil {
		util.GrpcErrorPrintf(err, "Failed to show the QoS")
		return util.ErrorGrpc
	}

	if reply.GetOk() {
		PrintAllQos(reply.QosList)
		return util.ErrorSuccess
	} else {
		if name == "" {
			fmt.Printf("Can't find any qos! %s\n", reply.GetReason())
		} else {
			fmt.Printf("Can't find qos %s\n", name)
		}
		return util.ErrorBackEnd
	}
}

func FindAccount(name string) util.CraneCmdError {
	req := protos.QueryEntityInfoRequest{Uid: userUid, EntityType: protos.EntityType_Account, Name: name}
	reply, err := stub.QueryEntityInfo(context.Background(), &req)
	if err != nil {
		util.GrpcErrorPrintf(err, "Failed to find the account")
		return util.ErrorGrpc
	}

	if reply.GetOk() {
		PrintAccountTable(reply.AccountList)
		return util.ErrorSuccess
	} else {
		fmt.Println(reply.Reason)
		return util.ErrorBackEnd
	}
}

func BlockAccountOrUser(name string, entityType protos.EntityType, account string) util.CraneCmdError {
	req := protos.BlockAccountOrUserRequest{Uid: userUid, Block: true, EntityType: entityType, Name: name, Account: account}
	reply, err := stub.BlockAccountOrUser(context.Background(), &req)
	if err != nil {
		util.GrpcErrorPrintf(err, "Failed to block the entity")
		return util.ErrorGrpc
	}

	if reply.GetOk() {
		fmt.Printf("Block %s success!\n", name)
		return util.ErrorSuccess
	} else {
		fmt.Println(reply.Reason)
		return util.ErrorBackEnd
	}
}

func UnblockAccountOrUser(name string, entityType protos.EntityType, account string) util.CraneCmdError {
	req := protos.BlockAccountOrUserRequest{Uid: userUid, Block: false, EntityType: entityType, Name: name, Account: account}
	reply, err := stub.BlockAccountOrUser(context.Background(), &req)
	if err != nil {
		util.GrpcErrorPrintf(err, "Failed to unblock the entity")
		return util.ErrorGrpc
	}

	if reply.GetOk() {
		fmt.Printf("Unblock %s success!\n", name)
		return util.ErrorSuccess
	} else {
		fmt.Println(reply.Reason)
		return util.ErrorBackEnd
	}
}
