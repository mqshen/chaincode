package main
//
//
//import (
//	"errors"
//	"github.com/hyperledger/fabric/core/chaincode/shim"
//)
//
//func insertOrUpdate(stub shim.ChaincodeStubInterface, owner []byte, asset string, expire uint64, amount uint64, out bool, extension string) (bool, error) {
//	var columns []shim.Column
//	col1 := shim.Column{Value: &shim.Column_Bytes{Bytes: owner}}
//	columns = append(columns, col1)
//
//	col2 := shim.Column{Value: &shim.Column_String_{String_: asset}}
//	columns = append(columns, col2)
//
//	col3 := shim.Column{Value: &shim.Column_String_{String_: extension}}
//	columns = append(columns, col3)
//
//	col4 := shim.Column{Value: &shim.Column_Uint64{Uint64: expire}}
//	columns = append(columns, col4)
//
//	row, err := stub.GetRow(assetTableColumn, columns)
//	if err != nil {
//		return false, errors.New("Failed query row.")
//	}
//	if(len(row.GetColumns()) == 0) {
//		myLogger.Debug("start insert row for asset")
//		_, err = stub.InsertRow(
//			assetTableColumn,
//			shim.Row{
//				Columns: []*shim.Column{
//					&shim.Column{Value: &shim.Column_Bytes{Bytes: owner}},
//					&shim.Column{Value: &shim.Column_String_{String_: asset}},
//					&shim.Column{Value: &shim.Column_String_{String_: extension}},
//					&shim.Column{Value: &shim.Column_Uint64{Uint64: expire}},
//					&shim.Column{Value: &shim.Column_Uint64{Uint64: amount}}},
//			})
//		if err != nil {
//			myLogger.Debugf("Failed replace row: %s", err)
//			return false, errors.New("Failed inserting row.")
//		}
//		return true, nil
//	} else {
//		myLogger.Debug("start replace row for asset")
//		balance := row.Columns[4].GetUint64()
//		if(out) {
//			balance = balance - amount
//		} else {
//			balance = balance + amount
//		}
//		_, err = stub.ReplaceRow(
//			assetTableColumn,
//			shim.Row{
//				Columns: []*shim.Column{
//					&shim.Column{Value: &shim.Column_Bytes{Bytes: owner}},
//					&shim.Column{Value: &shim.Column_String_{String_: asset}},
//					&shim.Column{Value: &shim.Column_String_{String_: extension}},
//					&shim.Column{Value: &shim.Column_Uint64{Uint64: expire}},
//					&shim.Column{Value: &shim.Column_Uint64{Uint64: balance}}},
//			})
//		if err != nil {
//			myLogger.Debugf("Failed replace row: %s", err)
//			return false, errors.New("Failed replace row.")
//		}
//		return true, nil
//	}
//}
//
//func  insertOrUpdateIssue(stub shim.ChaincodeStubInterface, owner []byte, asset string, amount uint64) (bool, error) {
//	var columns []shim.Column
//	col1 := shim.Column{Value: &shim.Column_Bytes{Bytes: owner}}
//	columns = append(columns, col1)
//
//	col2 := shim.Column{Value: &shim.Column_String_{String_: asset}}
//	columns = append(columns, col2)
//
//	row, err := stub.GetRow(issueTableColumn, columns)
//	if err != nil {
//		return false, errors.New("Failed query row.")
//	}
//	if(&row == nil || len(row.GetColumns()) == 0) {
//		myLogger.Debug("start insert row for asset")
//		_, err = stub.InsertRow(
//			issueTableColumn,
//			shim.Row{
//				Columns: []*shim.Column{
//					&shim.Column{Value: &shim.Column_Bytes{Bytes: owner}},
//					&shim.Column{Value: &shim.Column_String_{String_: asset}},
//					&shim.Column{Value: &shim.Column_Uint64{Uint64: amount}}},
//			})
//		if err != nil {
//			return false, errors.New("Failed inserting row.")
//		}
//		return true, nil
//	} else {
//		myLogger.Debug("start replace row for asset")
//		balance := row.Columns[2].GetUint64()
//		balance = balance + amount
//		_, err = stub.ReplaceRow(
//			assetTableColumn,
//			shim.Row{
//				Columns: []*shim.Column{
//					&shim.Column{Value: &shim.Column_Bytes{Bytes: owner}},
//					&shim.Column{Value: &shim.Column_String_{String_: asset}},
//					&shim.Column{Value: &shim.Column_Uint64{Uint64: balance}}},
//			})
//		if err != nil {
//			return false, errors.New("Failed replace row.")
//		}
//		return true, nil
//	}
//}