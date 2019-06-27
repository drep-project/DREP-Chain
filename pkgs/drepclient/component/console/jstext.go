package console

var Modules = map[string]string{
	//"account":   Personal_JS,
	"rpc": RPC_JS,
}

const RPC_JS = `
drep._extend({
	property: 'rpc',
	methods: [],
	properties: [
		new drep._extend.Property({
			name: 'modules',
			getter: 'rpc_modules'
		}),
	]
});
`
