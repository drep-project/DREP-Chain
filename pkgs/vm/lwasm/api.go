package lwasm

import (
	"encoding/json"
	"errors"
	"github.com/drep-project/drep-chain/pkgs/vm/lwasm/abi"
	"io/ioutil"
	"os"
	path2 "path"
)

type VmApi struct {
    VmService *VmService
}


func (vmApi *VmApi) ImportABI(contractName string, path string) error {
	fs, err := 	os.Stat(path)
	if err != nil {
		return nil
	}
	if fs.IsDir() {
		return errors.New("path is not a file")
	}
	contents, err := ioutil.ReadFile(path)
	if err != nil {
		return nil
	}
	_, err = vmApi.VmService.DatabaseApi.GetStorage(contractName, false)
	if err != nil {
		return err
	}
	abi := &abi.ABI{}
	err = json.Unmarshal(contents, abi)
	if err != nil {
		return err
	}
	newPath := path2.Join(vmApi.VmService.Config.AbiPath, contractName)
	fs, err = 	os.Stat(newPath)
	if err != os.ErrNotExist {
		return errors.New("abi has exist")
	}

	err = ioutil.WriteFile(newPath, contents,os.ModePerm)
	if err != nil {
		return nil
	}
	return nil
}

func (vmApi *VmApi) DeleteABI(contractName string) error {
	abiPath := path2.Join(vmApi.VmService.Config.AbiPath, contractName)
	_, err := 	os.Stat(abiPath)
	if err != nil {
		return err
	}
	err = os.Remove(abiPath)
	if err != nil {
		return nil
	}
	return nil
}