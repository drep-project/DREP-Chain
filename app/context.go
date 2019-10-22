package app

import (
	"encoding/json"
	"fmt"
	"github.com/asaskevich/EventBus"
	"github.com/pkg/errors"
	"gopkg.in/urfave/cli.v1"
	"reflect"
)

var (
	CommandHelpTemplate = `{{.cmd.Name}}{{if .cmd.Subcommands}} command{{end}}{{if .cmd.Flags}} [command options]{{end}} [arguments...]
{{if .cmd.Description}}{{.cmd.Description}}
{{end}}{{if .cmd.Subcommands}}
SUBCOMMANDS:
	{{range .cmd.Subcommands}}{{.Name}}{{with .ShortName}}, {{.}}{{end}}{{ "\t" }}{{.Usage}}
	{{end}}{{end}}{{if .categorizedFlags}}
{{range $idx, $categorized := .categorizedFlags}}{{$categorized.Name}} OPTIONS:
{{range $categorized.Flags}}{{"\t"}}{{.}}
{{end}}
{{end}}{{end}}`
)

func init() {
	cli.AppHelpTemplate = `{{.Name}} {{if .Flags}}[global options] {{end}}command{{if .Flags}} [command options]{{end}} [arguments...]

VERSION:
   {{.Version}}

COMMANDS:
   {{range .Commands}}{{.Name}}{{with .ShortName}}, {{.}}{{end}}{{ "\t" }}{{.Usage}}
   {{end}}{{if .Flags}}
GLOBAL OPTIONS:
   {{range .Flags}}{{.}}
   {{end}}{{end}}
`

	cli.CommandHelpTemplate = CommandHelpTemplate
}

// CommonConfig read before app run,this fuction shared by other moudles
type CommonConfig struct {
	HomeDir    string `json:"homeDir,omitempty"`
	ConfigFile string `json:"configFile,omitempty"`
}

// API describes the set of methods offered over the RPC interface
type API struct {
	Namespace string      // namespace under which the rpc methods of Service are exposed
	Version   string      // api version for DApp's
	Service   interface{} // receiver instance which holds the methods
	Public    bool        // indication if the methods must be considered safe for public use
}

// ExecuteContext centralizes all the data and global parameters of application execution,
// and each service can read the part it needs.
type ExecuteContext struct {
	ConfigPath   string
	CommonConfig *CommonConfig //
	PhaseConfig  map[string]json.RawMessage
	Cli          *cli.Context
	LifeBus      EventBus.Bus
	Services     []Service

	GitCommit string
	Usage     string
	Quit      chan struct{}
}

// GetService In addition, there is a dependency relationship between services.
// This method is used to find the dependency services you need in the context.
func (econtext *ExecuteContext) GetService(name string) Service {
	for _, service := range econtext.Services {
		if service.Name() == name {
			return service
		}
	}
	return nil
}

func (econtext *ExecuteContext) replaceService(replaceService, serviceValue Service) {
	preAddServices := econtext.Services
	for index, service := range preAddServices {
		if service == replaceService {
			econtext.Services = []Service{}
			preService := preAddServices[0:index]
			suffService := preAddServices[index:]
			econtext.Services = append(econtext.Services, preService...)
			econtext.Services = append(econtext.Services, serviceValue)
			econtext.Services = append(econtext.Services, suffService...)
			break
		}
	}
}

// addServiceType add a service and iterator all service that has added in and fields in current service,
// if exist , set set service in the field
func (econtext *ExecuteContext) addService(serviceValue reflect.Value) {
	econtext.Services = append(econtext.Services, serviceValue.Interface().(Service))
}

func (econtext *ExecuteContext) resolveService(service Service) {
	serviceValue := reflect.ValueOf(service)
	serviceType := reflect.TypeOf(service)
	serviceNumFields := serviceType.Elem().NumField()
	for i := 0; i < serviceNumFields; i++ {
		serviceTypeField := serviceType.Elem().Field(i)
		refServiceName := GetServiceTag(serviceTypeField)
		if refServiceName != "" {
			preAddServices := econtext.Services
			hasService := false
			for _, addedService := range preAddServices {
				if addedService.Name() == refServiceName {
					serviceValue.Elem().Field(i).Set(reflect.ValueOf(addedService))
					hasService = true
				}
			}

			if !hasService {
				fmt.Println(fmt.Sprintf("service not exist %s require %s", serviceValue.Interface().(Service).Name(), refServiceName))
			}
		}
	}
}

//	GetConfig Configuration is divided into several segments,
//	each service only needs to obtain its own configuration data,
//	and the parsing process is also controlled by each service itself.
func (econtext *ExecuteContext) GetConfig(phaseName string) json.RawMessage {
	phaseConfig, ok := econtext.PhaseConfig[phaseName]
	if ok {
		return phaseConfig
	} else {
		return nil
	}
}

func (econtext *ExecuteContext) FlatConfig(phaseName string) error {
	phaseConfig, ok := econtext.PhaseConfig[phaseName]
	if ok {
		subConfig := make(map[string]json.RawMessage)
		subJson, err := phaseConfig.MarshalJSON()
		if err != nil {
			return err
		}
		err = json.Unmarshal(subJson, &subConfig)
		if err != nil {
			return err
		}

		for key, val := range subConfig {
			subConfigItem, _ := val.MarshalJSON()
			if subConfigItem[0] == '{' {
				econtext.PhaseConfig[key] = val
			}
		}
		return nil
	} else {
		return nil
	}
}

// GetFlags aggregate command configuration items required for each service
func (econtext *ExecuteContext) AggerateFlags() ([]cli.Command, []cli.Flag) {
	allFlags := []cli.Flag{}
	allCommands := []cli.Command{}
	for _, service := range econtext.Services {
		commands, defaultFlags := service.CommandFlags()
		if commands != nil {
			allCommands = append(allCommands, commands...)
		}
		if defaultFlags != nil {
			allFlags = append(allFlags, defaultFlags...)
		}
	}
	return allCommands, allFlags
}

//	GetApis aggregate interface functions for each service to provide for use by RPC services
func (econtext *ExecuteContext) GetApis() []API {
	apis := []API{}
	for _, service := range econtext.Services {
		apis = append(apis, service.Api()...)
	}
	return apis
}

////	GetApis aggregate interface functions for each service to provide for use by RPC services
//func (econtext *ExecuteContext) GetMessages() (map[int]interface{}, error)  {
//	msg := map[int]interface{}{}
//	for _, service := range econtext.Services {
//		for k, v := range service.P2pMessages() {
//			if _, ok := msg[k]; ok {
//				return nil, errors.New("exist p2p message")
//			}
//			msg[k] = v
//		}
//	}
//	return msg, nil
//}

//	RequireService When a service depends on another service, RequireService is used to obtain the dependent service.
func (econtext *ExecuteContext) RequireService(name string) Service {
	for _, service := range econtext.Services {
		if service.Name() == name {
			return service
		}
	}
	panic(errors.Wrap(ErrServiceNotFound, name))
}

func (econtext *ExecuteContext) UnmashalConfig(serviceName string, config interface{}) error {
	service := econtext.GetService(serviceName)
	if service == nil {
		return errors.Wrapf(ErrServiceNotFound, "service name:%s", serviceName)
	}
	phase := econtext.GetConfig(service.Name())
	if phase == nil {
		return nil
	}
	fmt.Println(string(phase))
	err := json.Unmarshal(phase, config)
	if err != nil {
		return err
	}
	return nil
}
