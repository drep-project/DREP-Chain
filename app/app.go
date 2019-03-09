package app

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/drep-project/drep-chain/common/fileutil"
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"strings"

	"github.com/drep-project/drep-chain/common"
	"gopkg.in/urfave/cli.v1"
)

var (
	// General settings
	ConfigFileFlag = cli.StringFlag{
		Name:  "config",
		Usage: "TODO add config description",
	}
	
	HomeDirFlag = common.DirectoryFlag{
		Name:  "homedir",
		Usage: "Home directory for the datadir logdir and keystore",
	}
)

// DrepApp based on the cli.App, the module service operation is encapsulated.
// The purpose is to achieve the independence of each module and reduce dependencies as far as possible.
type DrepApp struct {
	Context *ExecuteContext
	*cli.App
}

// NewApp create a new app
func NewApp() *DrepApp {
	return &DrepApp{
		Context: &ExecuteContext{
			Quit:make(chan struct{}),
		},
		App:     cli.NewApp(),
	}
}

// AddService add a server into context
func (mApp DrepApp) addServiceInstance(service Service) {
	mApp.Context.AddService(service)
}

// AddServiceType add many services
func (mApp DrepApp) AddServices(serviceInstances ...interface{}) error {
	nilService := reflect.TypeOf((*Service)(nil)).Elem()

	for _, serviceInstance := range  serviceInstances {
		serviceType := reflect.TypeOf(serviceInstance)
		if serviceType.Kind() == reflect.Ptr {
			serviceType = serviceType.Elem()
		}
		serviceVal := reflect.New(serviceType)
		if !serviceVal.Type().Implements(nilService) {
			return errors.New("the service added not match service interface")
		}
		mApp.addService(serviceVal)
	}
	return nil
}

// addServiceType add a service and iterator all service that has added in and fields in current service,
// if exist , set set service in the field
func (mApp DrepApp) addService(serviceValue reflect.Value) {
	serviceType := serviceValue.Type()
	serviceNumFields := serviceType.Elem().NumField()
	for i := 0; i < serviceNumFields; i++{
		serviceValueField := serviceValue.Elem().Field(i)
		serviceTypeField := serviceType.Elem().Field(i)
		if serviceValueField.Type().Implements(reflect.TypeOf((*Service)(nil)).Elem()) {
			refServiceName := GetServiceTag(serviceTypeField)
			preAddServices := mApp.Context.Services
			hasService := false
			for _, addedService := range preAddServices {

				if addedService.Name() == refServiceName {
					//TODO the filed to be set must be set public field, but it wiil be better to set it as a private field ,
					//TODO There are still some technical difficulties that need to be overcome.
					//TODO UnsafePointer may help
					serviceValue.Elem().Field(i).Set(reflect.ValueOf(addedService))
					hasService = true
				}
			}

			if !hasService {
				fmt.Println(fmt.Sprintf("service not exist %s require %s", serviceValue.Interface().(Service).Name(), refServiceName))
				//dlog.Debug("service not exist",  "Service", addedService.Name() ,"RefService", refServiceName)
			}
		}
	}
	mApp.addServiceInstance(serviceValue.Interface().(Service))
}

// GetServiceTag read service tag name to match service that has added in
func GetServiceTag(field reflect.StructField) string {
	serviceTagStr := field.Tag.Get("service")
	if serviceTagStr == "" {
		return field.Name
	}
	serviceName := strings.Split(serviceTagStr, ",")
	if len(serviceName) == 0 {
		return field.Name
	}else {
		return serviceName[0]
	}
}

//TODO need a more graceful  command supporter
//TODO how to get password from terminal in wallet
//Run read the global configuration, parse the global command parameters,
// initialize the main process one by one, and execute the service before the main process starts.
func (mApp DrepApp) Run() error {
	mApp.Before = mApp.before
	mApp.Flags = append(mApp.Flags, ConfigFileFlag)

	allCommands, allFlags := mApp.Context.AggerateFlags()
	for i:= 0; i < len(allCommands); i++ {
		allCommands[i].Flags = append(allCommands[i].Flags, allFlags...)
		allCommands[i].Action = mApp.action
	}
	mApp.Flags = append(mApp.Flags, allFlags...)
	mApp.App.Commands = allCommands
	mApp.Action = mApp.action
	if err := mApp.App.Run(os.Args); err != nil {
		return err
	}
	return nil
}

// action used to init and run each services
func (mApp DrepApp) action(ctx *cli.Context) error {
	defer func() {
		if err:=recover();err!=nil{
			fmt.Println(err)
		}
		length := len(mApp.Context.Services)
		for i:= length; i >0; i-- {
			err := mApp.Context.Services[i - 1].Stop(mApp.Context)
			if err != nil {
				return
			}
		}
	}()
	mApp.Context.Cli = ctx   //NOTE this set is for different commmands
	for _, service := range mApp.Context.Services {
		err := service.Init(mApp.Context)
		if err != nil {
			return err
		}
	}

	for _, service := range mApp.Context.Services {
		err := service.Start(mApp.Context)
		if err != nil {
			return err
		}
	}
	return nil
}

//  read global config before main process
func (mApp DrepApp) before(ctx *cli.Context) error {
	mApp.Context.Cli = ctx

	homeDir := ""
	if ctx.GlobalIsSet(HomeDirFlag.Name) {
		homeDir = ctx.GlobalString(HomeDirFlag.Name)
	} else {
		homeDir = common.AppDataDir(mApp.Name, false)
	}
	mApp.Context.ConfigPath = homeDir

	mApp.Context.CommonConfig = &CommonConfig{
		HomeDir: homeDir,
	}
	phaseConfig, err := loadConfigFile(ctx, homeDir)

	if err != nil {
		return err
	}
	mApp.Context.PhaseConfig = phaseConfig

	return nil
}

//	loadConfigFile sed to read configuration files
func loadConfigFile(ctx *cli.Context, homeDir string) (map[string]json.RawMessage, error) {
	configFile := filepath.Join(homeDir, "config.json")

	if ctx.GlobalIsSet(ConfigFileFlag.Name) {
		file := ctx.GlobalString(ConfigFileFlag.Name)
		if fileutil.IsFileExists(file) {
			//report error when user specify
			return nil, errors.New("specify config file not exist")
		}
		configFile = file
	}

	if !fileutil.IsFileExists(configFile) {
		//use default
		cfg := &CommonConfig{
			HomeDir: homeDir,
			ConfigFile: configFile,
		}
		originConfigBytes, err := json.MarshalIndent(cfg,"", "\t")
		if err != nil {
			return nil, err
		}
		fileutil.EnsureFile(configFile)
		file, err :=  os.OpenFile(configFile, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0777)
		if err != nil {
			return nil, err
		}
		file.Write(originConfigBytes)
	}
	content, err := ioutil.ReadFile(configFile)
	if err != nil {
		return nil, err
	}
	jsonPhase := map[string]json.RawMessage{}
	err = json.Unmarshal(content, &jsonPhase)
	if err != nil {
		return nil, err
	}
	return jsonPhase, nil
}
