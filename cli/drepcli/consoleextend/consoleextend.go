package consoleextend

var Modules = map[string]string{
	"admin":      Admin_JS,
	"rpc":        RPC_JS,
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

const Admin_JS = `
drep._extend({
	property: 'admin',
	methods: [
		new drep._extend.Method({
			name: 'startRPC',
			call: 'admin_startRPC',
			params: 4,
			inputFormatter: [null, null, null, null]
		}),
		new drep._extend.Method({
			name: 'stopRPC',
			call: 'admin_stopRPC'
		}),
		new drep._extend.Method({
			name: 'startWS',
			call: 'admin_startWS',
			params: 4,
			inputFormatter: [null, null, null, null]
		}),
		new drep._extend.Method({
			name: 'stopWS',
			call: 'admin_stopWS'
		}),
	]
});
`